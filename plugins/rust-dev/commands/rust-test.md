---
name: rust-test
description: Run cargo test with optional features
arguments:
  - name: test
    description: Specific test name pattern (optional)
    required: false
  - name: features
    description: Features to enable (optional)
    required: false
---

# Run Rust Tests

Execute cargo test with output capture.

```bash
#!/bin/bash
set -e

# Build cargo test command
CMD="cargo test"

# Add test filter if specified
if [ -n "$test" ]; then
    CMD="$CMD $test"
fi

# Add features if specified
if [ -n "$features" ]; then
    CMD="$CMD --features $features"
fi

# Add verbose output
CMD="$CMD -- --nocapture"

echo "Running: $CMD"
eval $CMD
```

## Usage Examples

```
/rust-test                              # Run all tests
/rust-test test=user_service            # Run specific test
/rust-test features=postgres            # Run with features
```
