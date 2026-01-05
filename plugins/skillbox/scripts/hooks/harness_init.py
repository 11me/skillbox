#!/usr/bin/env python3
"""CLI: Initialize harness files directly, bypassing PreToolUse hook.

This script creates harness.json and features.json using direct Python I/O,
which bypasses the PreToolUse guard hook that blocks Write/Edit tool calls.

Usage:
    python3 harness_init.py --project-dir /path/to/project
    python3 harness_init.py --project-dir /path --features '[{"id":"f1","description":"..."}]'
"""

import argparse
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))

from lib.bootstrap import create_harness_state, get_default_verification_command
from lib.features import FeatureList, get_features_path


def main() -> None:
    parser = argparse.ArgumentParser(description="Initialize harness files")
    parser.add_argument("--project-dir", default=".", help="Project root directory")
    parser.add_argument(
        "--features",
        help='JSON array: [{"id": "...", "description": "..."}]',
    )
    parser.add_argument(
        "--no-beads",
        action="store_true",
        help="Skip automatic beads task creation",
    )
    args = parser.parse_args()

    project_dir = Path(args.project_dir).resolve()

    # 1. Create harness.json (uses existing function)
    harness_path = create_harness_state(project_dir)
    print(f"Created: {harness_path}")

    # 2. Create features.json if provided
    if args.features:
        try:
            features_data = json.loads(args.features)
        except json.JSONDecodeError as e:
            print(f"Error parsing features JSON: {e}", file=sys.stderr)
            sys.exit(1)

        feature_list = FeatureList()

        for f in features_data:
            if "id" not in f:
                print(f"Error: feature missing 'id' field: {f}", file=sys.stderr)
                sys.exit(1)

            verification = f.get("verification") or get_default_verification_command(
                project_dir, f["id"]
            )
            feature_list.add_feature(
                feature_id=f["id"],
                description=f.get("description", f["id"]),
                verification=verification,
                auto_create_beads=not args.no_beads,
            )

        features_path = get_features_path(project_dir)
        feature_list.save(features_path)
        print(f"Created: {features_path} ({len(features_data)} features)")


if __name__ == "__main__":
    main()
