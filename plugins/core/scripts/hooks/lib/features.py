"""JSON feature list management for long-running agent harness.

This module provides immutable feature tracking that persists across sessions.
Features have statuses, verification commands, and optional beads task linking.

Key design principles (from Anthropic's "Effective Harnesses" article):
- JSON format is less likely to be accidentally modified by the model
- Clear status enum prevents ambiguous states
- Verification commands enforce testing before completion
"""

import json
import re
import subprocess
from dataclasses import dataclass
from datetime import datetime
from enum import Enum
from pathlib import Path


class FeatureStatus(str, Enum):
    """Feature lifecycle states.

    pending → in_progress → implemented → verified
                                 ↓
                              failed
                                 ↓
                          (fix & retry)
    """

    PENDING = "pending"
    IN_PROGRESS = "in_progress"
    IMPLEMENTED = "implemented"
    VERIFIED = "verified"
    FAILED = "failed"


@dataclass
class Feature:
    """A single trackable feature."""

    id: str
    description: str
    status: FeatureStatus
    verification: str | None = None
    beads_id: str | None = None
    last_verified: str | None = None
    verification_output: str | None = None

    def to_dict(self) -> dict:
        """Convert to JSON-serializable dict."""
        return {
            "id": self.id,
            "description": self.description,
            "status": self.status.value,
            "verification": self.verification,
            "beads_id": self.beads_id,
            "last_verified": self.last_verified,
            "verification_output": self.verification_output,
        }

    @classmethod
    def from_dict(cls, data: dict) -> "Feature":
        """Create from dict."""
        return cls(
            id=data["id"],
            description=data["description"],
            status=FeatureStatus(data["status"]),
            verification=data.get("verification"),
            beads_id=data.get("beads_id"),
            last_verified=data.get("last_verified"),
            verification_output=data.get("verification_output"),
        )


class FeatureList:
    """Manages a list of features with persistence."""

    def __init__(
        self,
        version: str = "1.0.0",
        created: str | None = None,
        updated: str | None = None,
        features: list[Feature] | None = None,
    ):
        self.version = version
        self.created = created or datetime.now().isoformat()
        self.updated = updated or self.created
        self.features = features or []

    @classmethod
    def load(cls, path: Path) -> "FeatureList | None":
        """Load feature list from JSON file.

        Returns None if file doesn't exist or is invalid.
        """
        if not path.exists():
            return None

        try:
            data = json.loads(path.read_text())
            features = [Feature.from_dict(f) for f in data.get("features", [])]
            return cls(
                version=data.get("version", "1.0.0"),
                created=data.get("created", ""),
                updated=data.get("updated", ""),
                features=features,
            )
        except (json.JSONDecodeError, KeyError, ValueError):
            return None

    def save(self, path: Path) -> None:
        """Save feature list to JSON file."""
        self.updated = datetime.now().isoformat()
        data = {
            "version": self.version,
            "created": self.created,
            "updated": self.updated,
            "features": [f.to_dict() for f in self.features],
        }
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(json.dumps(data, indent=2))

    def add_feature(
        self,
        feature_id: str,
        description: str,
        verification: str | None = None,
        auto_create_beads: bool = True,
    ) -> Feature:
        """Add a new feature with optional beads task creation.

        Args:
            feature_id: Unique identifier (e.g., "auth-login")
            description: Human-readable description
            verification: Command to verify the feature
            auto_create_beads: If True, create linked beads task

        Returns:
            The created Feature instance.
        """
        beads_id = None
        if auto_create_beads:
            beads_id = _create_beads_task(description)

        feature = Feature(
            id=feature_id,
            description=description,
            status=FeatureStatus.PENDING,
            verification=verification,
            beads_id=beads_id,
        )
        self.features.append(feature)
        return feature

    def get_feature(self, feature_id: str) -> Feature | None:
        """Get feature by ID."""
        for f in self.features:
            if f.id == feature_id:
                return f
        return None

    def update_status(
        self,
        feature_id: str,
        status: FeatureStatus,
        verification_output: str | None = None,
    ) -> bool:
        """Update feature status.

        Args:
            feature_id: Feature to update
            status: New status
            verification_output: Output from verification command (for verified/failed)

        Returns:
            True if feature was found and updated.
        """
        feature = self.get_feature(feature_id)
        if not feature:
            return False

        feature.status = status

        if status in (FeatureStatus.VERIFIED, FeatureStatus.FAILED):
            feature.last_verified = datetime.now().isoformat()
            feature.verification_output = verification_output

        # Auto-close beads task when verified
        if status == FeatureStatus.VERIFIED and feature.beads_id:
            _close_beads_task(feature.beads_id, f"Feature {feature_id} verified")

        # Update beads task status
        if status == FeatureStatus.IN_PROGRESS and feature.beads_id:
            _update_beads_status(feature.beads_id, "in_progress")

        return True

    def get_summary(self) -> dict[str, int]:
        """Get count of features by status."""
        summary: dict[str, int] = {s.value: 0 for s in FeatureStatus}
        for f in self.features:
            summary[f.status.value] += 1
        return summary

    def all_verified(self) -> bool:
        """Check if all features are verified."""
        return all(f.status == FeatureStatus.VERIFIED for f in self.features)

    def get_unverified(self) -> list[Feature]:
        """Get features that are not yet verified."""
        return [f for f in self.features if f.status != FeatureStatus.VERIFIED]

    def get_implemented_unverified(self) -> list[Feature]:
        """Get features that are implemented but not verified."""
        return [f for f in self.features if f.status == FeatureStatus.IMPLEMENTED]

    def get_next_feature(self) -> Feature | None:
        """Get highest priority unverified feature.

        Priority: in_progress > pending > implemented (needs verification)
        """
        for status in [
            FeatureStatus.IN_PROGRESS,
            FeatureStatus.PENDING,
            FeatureStatus.IMPLEMENTED,
        ]:
            for f in self.features:
                if f.status == status:
                    return f
        return None


def get_features_path(project_dir: Path | None = None) -> Path:
    """Get path to features.json."""
    if project_dir is None:
        project_dir = Path.cwd()
    return project_dir / ".claude" / "features.json"


def load_features(project_dir: Path | None = None) -> FeatureList | None:
    """Load features from project directory."""
    return FeatureList.load(get_features_path(project_dir))


def _create_beads_task(description: str) -> str | None:
    """Create a beads task and return its ID.

    Returns None if beads is not available or creation fails.
    """
    try:
        result = subprocess.run(
            ["bd", "create", "--title", description, "-t", "feature", "-p", "2"],
            capture_output=True,
            text=True,
            timeout=5,
        )
        if result.returncode != 0:
            return None

        # Parse task ID from output (e.g., "Created: skills-abc123")
        match = re.search(r"(?:Created|created):\s*(\S+)", result.stdout)
        if match:
            return match.group(1)

        # Try parsing just the ID if it's the only output
        output = result.stdout.strip()
        if output and " " not in output:
            return output

        return None
    except (subprocess.TimeoutExpired, FileNotFoundError, OSError):
        return None


def _close_beads_task(task_id: str, reason: str) -> bool:
    """Close a beads task."""
    try:
        result = subprocess.run(
            ["bd", "close", task_id, "--reason", reason],
            capture_output=True,
            timeout=5,
        )
        return result.returncode == 0
    except (subprocess.TimeoutExpired, FileNotFoundError, OSError):
        return False


def _update_beads_status(task_id: str, status: str) -> bool:
    """Update beads task status."""
    try:
        result = subprocess.run(
            ["bd", "update", task_id, "--status", status],
            capture_output=True,
            timeout=5,
        )
        return result.returncode == 0
    except (subprocess.TimeoutExpired, FileNotFoundError, OSError):
        return False
