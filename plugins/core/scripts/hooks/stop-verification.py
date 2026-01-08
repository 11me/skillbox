#!/usr/bin/env python3
"""Stop hook: Enforce verification before session end.

Blocks session end if there are implemented but unverified features.
This is the key enforcement mechanism from Anthropic's article:
"Mandatory self-verification before marking passing".

Without this, agents may declare victory prematurely.
"""

import json
import sys
from pathlib import Path

# Add lib to path
sys.path.insert(0, str(Path(__file__).parent))

from lib.bootstrap import is_harness_initialized
from lib.features import FeatureStatus, load_features
from lib.response import allow, block


def main() -> None:
    """Handle Stop event."""
    try:
        _ = json.load(sys.stdin)  # Consume stdin, not used currently
    except json.JSONDecodeError:
        pass

    cwd = Path.cwd()

    # If harness not initialized, defer to other stop hooks
    if not is_harness_initialized(cwd):
        allow("Stop")
        return

    features = load_features(cwd)
    if not features:
        allow("Stop")
        return

    # Check for implemented but unverified features
    unverified = features.get_implemented_unverified()

    if unverified:
        feature_list = "\n".join(f"  - **{f.id}**: {f.description}" for f in unverified[:5])
        if len(unverified) > 5:
            feature_list += f"\n  ... and {len(unverified) - 5} more"

        block(
            reason="Unverified features detected",
            event="Stop",
            context=(
                f"**{len(unverified)} feature(s) implemented but not verified:**\n"
                f"{feature_list}\n\n"
                "Run verification before ending session:\n"
                "```bash\n"
                "/harness-verify --implemented\n"
                "```\n\n"
                "Or verify individually:\n"
                "```bash\n"
                f"/harness-verify {unverified[0].id}\n"
                "```\n\n"
                "If verification is not possible right now, you can:\n"
                f"- Reset to pending: `/harness-update {unverified[0].id} pending`\n"
                "- Continue implementing"
            ),
        )
        return

    # Check for failed features
    failed = [f for f in features.features if f.status == FeatureStatus.FAILED]
    if failed:
        feature_list = "\n".join(
            f"  - **{f.id}**: {f.verification_output or 'No output'}"[:100] for f in failed[:3]
        )

        block(
            reason="Failed verifications detected",
            event="Stop",
            context=(
                f"**{len(failed)} feature(s) failed verification:**\n"
                f"{feature_list}\n\n"
                "Fix failures before completing session, or reset status:\n"
                f"```bash\n"
                f"/harness-update {failed[0].id} in_progress\n"
                "```"
            ),
        )
        return

    # All clear
    allow("Stop")


if __name__ == "__main__":
    main()
