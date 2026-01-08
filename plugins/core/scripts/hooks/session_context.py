#!/usr/bin/env python3
"""SessionStart hook: inject context at session start.

Provides:
- Current date (critical for YAML generation)
- Project type detection
- Serena auto-activation instruction
- Harness workflow rules (context reinforcement)
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

from lib.bootstrap import is_harness_initialized
from lib.detector import detect_flux, detect_project_types, detect_tdd_mode
from lib.response import session_output
from lib.tmux_state import cleanup_stale_states
from lib.tmux_state import save_state as save_tmux_state


def get_harness_rules() -> list[str]:
    """Get harness workflow rules for context reinforcement."""
    return [
        "## 🔄 Active Workflow Rules",
        "- **Harness**: Check `/harness-status` before work, verify after features",
        "- **Tasks**: Use beads (`bd create/update/close`) for tracking",
        "- **Checkpoints**: Save with `/checkpoint` before risky operations",
        "",
    ]


def get_task_enforcement_rules() -> list[str]:
    """Rules for task creation interruption handling.

    These rules ensure agents don't bypass task tracking when
    bd commands are interrupted or fail.
    """
    return [
        "## ⚠️ Task Creation Enforcement",
        "- If `bd create` is interrupted, MUST retry or ask user before proceeding",
        "- Never start implementation without confirmed active task",
        "- On any beads command failure, resolve before coding",
        "- Write/Edit operations WILL BE BLOCKED without active in_progress task",
        "",
    ]


def get_no_active_task_warning() -> str | None:
    """Check if active task exists and return warning if not.

    Returns a strong warning message if beads is initialized but
    there's no task in in_progress status. This is a soft gate -
    warns but doesn't block session start.
    """
    try:
        result = subprocess.run(
            ["bd", "list", "--status", "in_progress", "--json"],
            capture_output=True,
            text=True,
            timeout=5,
        )
        if result.returncode == 0:
            import json

            tasks = json.loads(result.stdout)
            if tasks:
                return None  # Has active task - all good

        # No active task - return warning
        return """## ⚠️ NO ACTIVE TASK

Workflow mode is active but **no task is in progress**.

**Before implementing anything, create or select a task:**
```bash
bd create --title "Your task description" -p 2
bd update <id> --status in_progress
```

**Or view ready tasks:** `bd ready`

> ⛔ Write/Edit operations will be BLOCKED until a task is active.
"""
    except Exception:
        # Don't warn on error - the guard will catch it later
        return None


def get_serena_project_name(cwd: Path) -> str:
    """Get Serena project name from config or directory name."""
    config_path = cwd / ".serena" / "config.yaml"
    if config_path.exists():
        try:
            import yaml

            config = yaml.safe_load(config_path.read_text())
            if config and "project_name" in config:
                return config["project_name"]
        except Exception:
            pass
    # Fallback to directory name
    return cwd.name


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

        # Get project name for activation
        project_name = get_serena_project_name(cwd)

        output_lines.append("**Serena enabled** — semantic navigation available")
        output_lines.append("   Tools: find_symbol, get_symbols_overview, write_memory")
        output_lines.append("")

        # Auto-activation instruction
        output_lines.append("**⚡ Serena Auto-Activate:**")
        output_lines.append(
            f'   → Call `mcp__plugin_serena_serena__activate_project` with project="{project_name}"'
        )
        output_lines.append("   This is REQUIRED before using semantic tools.")
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

    # 5.5 Harness workflow rules (context reinforcement)
    if is_harness_initialized(cwd):
        output_lines.extend(get_harness_rules())

    # 5.6 Task enforcement rules and warning (when beads is active)
    if beads_initialized:
        output_lines.extend(get_task_enforcement_rules())

        # Show warning if no active task
        no_task_warning = get_no_active_task_warning()
        if no_task_warning:
            output_lines.append(no_task_warning)

    # 6. GitOps rules reminder
    if project_type in ("helm", "gitops"):
        output_lines.append("**Rules:**")
        output_lines.append("- No literal secrets in values.yaml (use ExternalSecret)")
        output_lines.append("- Use refreshPolicy: OnChange for deterministic ESO updates")
        output_lines.append("- Validate with /helm-validate before completing work")
        output_lines.append("")

    # 7. K8s/Flux version enforcement via Context7
    is_flux = detect_flux(cwd)
    if project_type in ("helm", "gitops") or is_flux:
        output_lines.append("**Flux/K8s project detected**")
        output_lines.append("   CRITICAL: Use Context7 for ALL Helm chart versions")
        output_lines.append("   Workflow: resolve-library-id → query-docs")
        output_lines.append(
            "   Components: cert-manager, ingress-nginx, external-secrets, external-dns"
        )
        output_lines.append("   NEVER hardcode versions — always fetch from Context7")
        output_lines.append("")

    session_output("\n".join(output_lines))


if __name__ == "__main__":
    main()
