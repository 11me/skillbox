#!/usr/bin/env python3
"""SessionStart hook: inject context at session start.

Provides:
- Current date (critical for YAML generation)
- Project type detection
- Beads tasks integration
- Tool availability checks
- GitOps rules reminder
"""

import subprocess
import sys
from datetime import datetime
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))

from lib.detector import detect_project_types, detect_tdd_mode
from lib.response import session_output
from lib.tmux_state import cleanup_stale_states
from lib.tmux_state import save_state as save_tmux_state


def check_command_exists(cmd: str) -> bool:
    """Check if a command exists in PATH."""
    try:
        subprocess.run(
            ["which", cmd],
            capture_output=True,
            check=True,  # noqa: S603, S607
        )
        return True
    except (subprocess.CalledProcessError, FileNotFoundError):
        return False


def get_beads_ready() -> str | None:
    """Get ready tasks from beads if available."""
    if not check_command_exists("bd"):
        return None

    try:
        result = subprocess.run(
            ["bd", "ready"],  # noqa: S603, S607
            capture_output=True,
            text=True,
            timeout=5,
        )
        output = result.stdout.strip()
        if output and output != "No ready tasks":
            return output
    except (subprocess.TimeoutExpired, subprocess.CalledProcessError, FileNotFoundError):
        pass

    return None


def main() -> None:
    cwd = Path.cwd()
    output_lines: list[str] = []

    # 0. Save tmux state for consistent targeting in notification hooks
    save_tmux_state()

    # 0.5 Cleanup stale state files from orphaned sessions (>24h old)
    cleanup_stale_states()

    # 1. Current date
    today = datetime.now().strftime("%Y-%m-%d")
    output_lines.append(f"**Today:** {today}")
    output_lines.append("")

    # 2. Detect project type
    types = detect_project_types(cwd)
    project_type: str | None = None

    if types["helm"]:
        project_type = "helm"
        output_lines.append("**Project type:** Helm chart")
        output_lines.append("**Active skills:** helm-chart-developer, conventional-commit")
        output_lines.append("")

    elif types["gitops"]:
        project_type = "gitops"
        output_lines.append("**Project type:** GitOps repository")
        output_lines.append("**Commands:** /helm-scaffold, /helm-validate, /checkpoint")
        output_lines.append("")
        output_lines.append("**Layout:**")
        output_lines.append("```")
        output_lines.append("gitops/")
        output_lines.append("├── charts/app/              # Universal Helm chart")
        output_lines.append("├── apps/")
        output_lines.append("│   ├── base/                # Base HelmRelease")
        output_lines.append("│   ├── dev/                 # Dev overlay")
        output_lines.append("│   └── prod/                # Prod overlay")
        output_lines.append("└── infra/                   # Operators")
        output_lines.append("```")
        output_lines.append("")

    elif types["go"]:
        project_type = "go"
        output_lines.append("**Project type:** Go project")
        output_lines.append("")

        # Inject mandatory Go guidelines
        guidelines_path = (
            Path(__file__).parent.parent.parent / "skills/go/go-development/GO-GUIDELINES.md"
        )
        if guidelines_path.exists():
            guidelines = guidelines_path.read_text().strip()
            output_lines.append("## ⛔ MANDATORY - Follow these rules")
            output_lines.append("")
            output_lines.append(guidelines)
            output_lines.append("")
        else:
            # Fallback if file not found
            output_lines.append("**Linter enforces:**")
            output_lines.append("- `userID` not `userId` (var-naming)")
            output_lines.append("- `any` not `interface{}` (use-any)")
            output_lines.append("- No `common/helpers/utils/shared/misc` packages")
            output_lines.append("")
            output_lines.append("→ Run `golangci-lint run` after completing Go tasks")
            output_lines.append("")

        output_lines.append("- Dependencies: always use `@latest` (hook enforces)")
        output_lines.append(
            "- Repository queries: use Filter pattern (`XxxFilter` + `getXxxCondition()`)"
        )
        output_lines.append("")

    elif types["python"]:
        project_type = "python"
        output_lines.append("**Project type:** Python project")
        output_lines.append("")

    elif types["node"]:
        project_type = "node"
        output_lines.append("**Project type:** Node.js project")
        output_lines.append("")

    # Serena detection
    if types["serena"]:
        if not project_type:
            project_type = "serena"
        output_lines.append("**Serena enabled** — semantic navigation available")
        output_lines.append("   Tools: find_symbol, get_symbols_overview, write_memory")
        output_lines.append("")

        # Check for recent checkpoints (manual and auto)
        serena_memories = cwd / ".serena" / "memories"
        if serena_memories.exists():
            # Manual checkpoints
            checkpoints = sorted(serena_memories.glob("checkpoint-*.md"), reverse=True)
            if checkpoints:
                latest = checkpoints[0].name
                output_lines.append(f"**Recent checkpoint:** `read_memory('{latest}')`")

            # Auto-checkpoints (from PreCompact/Stop hooks)
            auto_checkpoints = sorted(serena_memories.glob("auto-checkpoint-*.md"), reverse=True)
            if auto_checkpoints:
                latest_auto = auto_checkpoints[0].name
                output_lines.append(f"**Auto-checkpoint:** `read_memory('{latest_auto}')`")

            if checkpoints or auto_checkpoints:
                output_lines.append("")

    # 3. Check critical tools
    missing_tools: list[str] = []
    if project_type in ("helm", "gitops"):
        if not check_command_exists("helm"):
            missing_tools.append("helm")
        if not check_command_exists("kubectl"):
            missing_tools.append("kubectl")

    if missing_tools:
        output_lines.append(f"**Missing tools:** {' '.join(missing_tools)}")
        output_lines.append("")

    # 4. Beads integration
    bd_installed = check_command_exists("bd")
    beads_initialized = (cwd / ".beads").is_dir()

    if bd_installed and beads_initialized:
        bd_ready = get_beads_ready()
        if bd_ready:
            output_lines.append("**Ready tasks:**")
            output_lines.append("```")
            output_lines.append(bd_ready)
            output_lines.append("```")
            output_lines.append("")
    elif bd_installed and not beads_initialized:
        output_lines.append("**Task tracking:** beads available but not initialized")
        output_lines.append("→ Run `/init-project` for full setup, or `bd init` for beads only")
        output_lines.append("")

    # 5. TDD mode detection and injection
    tdd_status = detect_tdd_mode(cwd)
    if tdd_status["enabled"]:
        tdd_guidelines_path = (
            Path(__file__).parent.parent.parent / "skills/core/tdd-enforcer/TDD-GUIDELINES.md"
        )
        mode_label = "STRICT" if tdd_status["strict"] else "ACTIVE"
        output_lines.append(f"## 🧪 TDD Mode ({mode_label})")
        output_lines.append("")

        if tdd_guidelines_path.exists():
            guidelines = tdd_guidelines_path.read_text().strip()
            output_lines.append(guidelines)
        else:
            # Fallback if file not found
            output_lines.append("**Cycle:** RED → GREEN → REFACTOR")
            output_lines.append("1. Write failing test FIRST")
            output_lines.append("2. Minimal implementation to pass")
            output_lines.append("3. Refactor with tests passing")
        output_lines.append("")

    # 6. GitOps rules reminder
    if project_type in ("helm", "gitops"):
        output_lines.append("**Rules:**")
        output_lines.append("- No literal secrets in values.yaml (use ExternalSecret)")
        output_lines.append("- Use refreshPolicy: OnChange for deterministic ESO updates")
        output_lines.append("- Validate with /helm-validate before completing work")

    session_output("\n".join(output_lines))


if __name__ == "__main__":
    main()
