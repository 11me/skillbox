#!/usr/bin/env python3
"""
Validate SKILL.md files for correct YAML frontmatter.

Usage:
    python scripts/validate-skills.py SKILL.md [SKILL.md ...]
    python scripts/validate-skills.py plugins/skillbox/skills/
"""

import re
import sys
from pathlib import Path

# Try to use PyYAML if available, otherwise basic parsing
try:
    import yaml
    HAS_YAML = True
except ImportError:
    HAS_YAML = False


# Valid tools that can be used in allowed-tools
VALID_TOOLS = {"Read", "Write", "Edit", "Grep", "Glob", "Bash"}

# Name pattern: lowercase letters, numbers, hyphens
NAME_PATTERN = re.compile(r"^[a-z0-9-]+$")
MAX_NAME_LENGTH = 64
MAX_DESCRIPTION_LENGTH = 1024


def extract_frontmatter(content: str) -> tuple[str | None, list[str]]:
    """Extract YAML frontmatter from markdown content."""
    errors = []

    if not content.startswith("---"):
        errors.append("File must start with '---' (YAML frontmatter delimiter)")
        return None, errors

    # Find the closing delimiter
    lines = content.split("\n")
    end_idx = None

    for i, line in enumerate(lines[1:], start=1):
        if line.strip() == "---":
            end_idx = i
            break

    if end_idx is None:
        errors.append("Missing closing '---' for YAML frontmatter")
        return None, errors

    frontmatter = "\n".join(lines[1:end_idx])
    return frontmatter, errors


def parse_frontmatter(frontmatter: str) -> tuple[dict | None, list[str]]:
    """Parse YAML frontmatter into dictionary."""
    errors = []

    if HAS_YAML:
        try:
            data = yaml.safe_load(frontmatter)
            if not isinstance(data, dict):
                errors.append("Frontmatter must be a YAML mapping (key: value pairs)")
                return None, errors
            return data, errors
        except yaml.YAMLError as e:
            errors.append(f"Invalid YAML syntax: {e}")
            return None, errors
    else:
        # Basic parsing without PyYAML
        data = {}
        for line in frontmatter.split("\n"):
            line = line.strip()
            if not line or line.startswith("#"):
                continue
            if ":" in line:
                key, _, value = line.partition(":")
                key = key.strip()
                value = value.strip().strip('"').strip("'")
                # Handle arrays
                if value.startswith("[") and value.endswith("]"):
                    value = [v.strip().strip('"').strip("'")
                             for v in value[1:-1].split(",") if v.strip()]
                data[key] = value
        return data, errors


def validate_name(name: str | None) -> list[str]:
    """Validate the 'name' field."""
    errors = []

    if name is None:
        errors.append("Missing required field: 'name'")
        return errors

    if not isinstance(name, str):
        errors.append(f"'name' must be a string, got {type(name).__name__}")
        return errors

    if not NAME_PATTERN.match(name):
        errors.append(
            f"'name' must be kebab-case (lowercase letters, numbers, hyphens): '{name}'"
        )

    if len(name) > MAX_NAME_LENGTH:
        errors.append(
            f"'name' exceeds {MAX_NAME_LENGTH} characters: {len(name)} chars"
        )

    return errors


def validate_description(description: str | None) -> list[str]:
    """Validate the 'description' field."""
    errors = []

    if description is None:
        errors.append("Missing required field: 'description'")
        return errors

    if not isinstance(description, str):
        errors.append(f"'description' must be a string, got {type(description).__name__}")
        return errors

    if len(description) > MAX_DESCRIPTION_LENGTH:
        errors.append(
            f"'description' exceeds {MAX_DESCRIPTION_LENGTH} characters: {len(description)} chars"
        )

    # Check for recommended format: "What it does. Use when..."
    desc_lower = description.lower()
    if "use when" not in desc_lower and "use for" not in desc_lower:
        errors.append(
            "Warning: 'description' should include trigger phrases like 'Use when...' or 'Use for...'"
        )

    return errors


def validate_globs(globs: list | None) -> list[str]:
    """Validate the optional 'globs' field."""
    errors = []

    if globs is None:
        return errors  # Optional field

    if not isinstance(globs, list):
        errors.append(f"'globs' must be an array, got {type(globs).__name__}")
        return errors

    for i, glob in enumerate(globs):
        if not isinstance(glob, str):
            errors.append(f"'globs[{i}]' must be a string, got {type(glob).__name__}")

    return errors


def validate_allowed_tools(tools: str | list | None) -> list[str]:
    """Validate the optional 'allowed-tools' field."""
    errors = []

    if tools is None:
        return errors  # Optional field

    # Parse comma-separated string or list
    if isinstance(tools, str):
        tool_list = [t.strip() for t in tools.split(",")]
    elif isinstance(tools, list):
        tool_list = tools
    else:
        errors.append(f"'allowed-tools' must be a string or array, got {type(tools).__name__}")
        return errors

    for tool in tool_list:
        if tool not in VALID_TOOLS:
            errors.append(
                f"Unknown tool in 'allowed-tools': '{tool}'. "
                f"Valid tools: {', '.join(sorted(VALID_TOOLS))}"
            )

    return errors


def validate_skill_file(filepath: Path) -> list[str]:
    """Validate a single SKILL.md file."""
    errors = []

    try:
        content = filepath.read_text(encoding="utf-8")
    except Exception as e:
        return [f"Cannot read file: {e}"]

    # Extract frontmatter
    frontmatter, extract_errors = extract_frontmatter(content)
    errors.extend(extract_errors)

    if frontmatter is None:
        return errors

    # Parse frontmatter
    data, parse_errors = parse_frontmatter(frontmatter)
    errors.extend(parse_errors)

    if data is None:
        return errors

    # Validate required fields
    errors.extend(validate_name(data.get("name")))
    errors.extend(validate_description(data.get("description")))

    # Validate optional fields
    errors.extend(validate_globs(data.get("globs")))
    errors.extend(validate_allowed_tools(data.get("allowed-tools")))

    return errors


def find_skill_files(path: Path) -> list[Path]:
    """Find all SKILL.md files in a directory."""
    if path.is_file():
        return [path] if path.name == "SKILL.md" else []

    return list(path.rglob("SKILL.md"))


def main() -> int:
    """Main entry point."""
    if len(sys.argv) < 2:
        print("Usage: validate-skills.py <path> [path ...]", file=sys.stderr)
        print("  path: SKILL.md file or directory containing skills", file=sys.stderr)
        return 1

    all_errors: dict[str, list[str]] = {}
    total_files = 0
    warning_count = 0

    for arg in sys.argv[1:]:
        path = Path(arg)

        if not path.exists():
            print(f"Path not found: {path}", file=sys.stderr)
            continue

        skill_files = find_skill_files(path)

        for skill_file in skill_files:
            total_files += 1
            errors = validate_skill_file(skill_file)

            if errors:
                # Separate warnings from errors
                real_errors = [e for e in errors if not e.startswith("Warning:")]
                warnings = [e for e in errors if e.startswith("Warning:")]

                if real_errors:
                    all_errors[str(skill_file)] = real_errors

                if warnings:
                    warning_count += len(warnings)
                    for warning in warnings:
                        print(f"{skill_file}: {warning}")

    # Report results
    if all_errors:
        print(f"\n{'='*60}", file=sys.stderr)
        print(f"Validation failed for {len(all_errors)} file(s):", file=sys.stderr)
        print(f"{'='*60}", file=sys.stderr)

        for filepath, errors in all_errors.items():
            print(f"\n{filepath}:", file=sys.stderr)
            for error in errors:
                print(f"  - {error}", file=sys.stderr)

        return 1

    print(f"Validated {total_files} skill file(s) successfully.")
    if warning_count:
        print(f"  ({warning_count} warning(s))")

    return 0


if __name__ == "__main__":
    sys.exit(main())
