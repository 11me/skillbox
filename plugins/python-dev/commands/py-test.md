---
name: py-test
description: Run pytest with coverage for the current project
arguments:
  - name: path
    description: Specific test path or pattern (optional)
    required: false
  - name: marker
    description: Run tests with specific marker (slow, integration)
    required: false
---

# Run Python Tests

Execute pytest with coverage reporting.

```bash
#!/bin/bash
set -e

# Build pytest command
CMD="uv run pytest"

# Add coverage
CMD="$CMD --cov --cov-report=term-missing"

# Add marker filter if specified
if [ -n "$marker" ]; then
    CMD="$CMD -m \"$marker\""
fi

# Add path if specified
if [ -n "$path" ]; then
    CMD="$CMD $path"
fi

# Add verbose and fail fast
CMD="$CMD -v -x"

echo "Running: $CMD"
eval $CMD
```

## Usage Examples

```
/py-test                           # Run all tests
/py-test path=tests/unit           # Run unit tests only
/py-test marker=integration        # Run integration tests
/py-test path=tests/test_api.py    # Run specific file
```
