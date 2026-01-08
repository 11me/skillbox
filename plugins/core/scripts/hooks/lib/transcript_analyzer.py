"""Transcript analyzer for TDD compliance checking.

Analyzes Claude Code session transcript to determine:
1. Whether code files were modified (Write, Edit, MultiEdit)
2. Whether tests were executed (Bash with test commands)
3. TDD compliance status
"""

import re
from dataclasses import dataclass

# Test file patterns per language
TEST_FILE_PATTERNS = {
    "go": re.compile(r"_test\.go$"),
    "typescript": re.compile(r"\.(test|spec)\.(ts|tsx|js|jsx)$"),
    "python": re.compile(r"(^test_.*\.py$|.*_test\.py$)"),
    "rust": re.compile(r"(_test\.rs$|/tests/.*\.rs$)"),
}

# Test command patterns
TEST_COMMAND_PATTERNS = [
    re.compile(r"\bgo\s+test\b"),
    re.compile(r"\bcargo\s+test\b"),
    re.compile(r"\bnpm\s+(run\s+)?test\b"),
    re.compile(r"\bpnpm\s+(run\s+)?test\b"),
    re.compile(r"\bvitest\b"),
    re.compile(r"\bjest\b"),
    re.compile(r"\bpytest\b"),
    re.compile(r"\bpython\s+-m\s+pytest\b"),
    re.compile(r"\bpython\s+-m\s+unittest\b"),
]


@dataclass
class TDDAnalysisResult:
    """Result of TDD compliance analysis."""

    code_modified: bool
    test_files_modified: bool
    tests_executed: bool
    modified_files: list[str]
    test_commands: list[str]
    is_compliant: bool
    message: str


def is_test_file(file_path: str) -> bool:
    """Check if a file path is a test file."""
    if not file_path:
        return False

    # Check against all language patterns
    for pattern in TEST_FILE_PATTERNS.values():
        if pattern.search(file_path):
            return True

    # Also check common test directories
    if "/test/" in file_path or "/tests/" in file_path or "/__tests__/" in file_path:
        return True

    return False


def is_test_command(command: str) -> bool:
    """Check if a command is a test execution command."""
    if not command:
        return False

    for pattern in TEST_COMMAND_PATTERNS:
        if pattern.search(command):
            return True

    return False


def analyze_transcript(transcript_content: str) -> TDDAnalysisResult:
    """Analyze transcript for TDD compliance.

    Args:
        transcript_content: Raw transcript text content

    Returns:
        TDDAnalysisResult with compliance status
    """
    modified_files: set[str] = set()
    test_files_modified: set[str] = set()
    test_commands: list[str] = []

    # Look for Write/Edit tool usage patterns in transcript
    # Pattern: tool="Write" or tool="Edit" with file_path
    write_pattern = re.compile(
        r'(?:Write|Edit|MultiEdit).*?(?:file_path|filePath)["\']?\s*[:=]\s*["\']([^"\']+)["\']',
        re.IGNORECASE | re.DOTALL,
    )

    # Also look for simpler patterns like: Writing to file: /path/to/file
    simple_write_pattern = re.compile(
        r'(?:Writing|Editing|Modified|Created).*?(?:file|to):\s*["\']?([^\s"\']+\.[a-z]+)',
        re.IGNORECASE,
    )

    # Look for Bash commands
    bash_pattern = re.compile(
        r'Bash.*?command["\']?\s*[:=]\s*["\']([^"\']+)["\']', re.IGNORECASE | re.DOTALL
    )

    # Extract file modifications
    for match in write_pattern.finditer(transcript_content):
        file_path = match.group(1)
        if is_test_file(file_path):
            test_files_modified.add(file_path)
        else:
            modified_files.add(file_path)

    for match in simple_write_pattern.finditer(transcript_content):
        file_path = match.group(1)
        if is_test_file(file_path):
            test_files_modified.add(file_path)
        else:
            modified_files.add(file_path)

    # Extract test commands
    for match in bash_pattern.finditer(transcript_content):
        command = match.group(1)
        if is_test_command(command):
            test_commands.append(command)

    # Also look for direct test command mentions
    for pattern in TEST_COMMAND_PATTERNS:
        for match in pattern.finditer(transcript_content):
            cmd = match.group(0)
            if cmd not in test_commands:
                test_commands.append(cmd)

    # Determine compliance
    code_modified = len(modified_files) > 0
    tests_executed = len(test_commands) > 0
    has_test_files = len(test_files_modified) > 0

    # TDD compliance rules:
    # 1. If code was modified, tests should have been run
    # 2. Ideally, test files should be modified before/with code
    is_compliant = not code_modified or tests_executed

    # Generate message
    if not code_modified:
        message = "No code modifications detected - TDD check not applicable"
    elif tests_executed:
        if has_test_files:
            message = "TDD compliant: Tests written and executed"
        else:
            message = (
                "Tests executed, but no test files were modified. "
                "Consider writing tests first (Red phase)"
            )
    else:
        message = (
            f"TDD Warning: {len(modified_files)} code file(s) modified "
            "without running tests. Run tests to verify changes."
        )

    return TDDAnalysisResult(
        code_modified=code_modified,
        test_files_modified=has_test_files,
        tests_executed=tests_executed,
        modified_files=list(modified_files),
        test_commands=test_commands,
        is_compliant=is_compliant,
        message=message,
    )


def read_transcript(transcript_path: str) -> str | None:
    """Read transcript file content."""
    try:
        with open(transcript_path, "r", encoding="utf-8") as f:
            return f.read()
    except (IOError, OSError):
        return None
