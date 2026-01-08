---
name: tdd-enforcer
description: Test-Driven Development workflow patterns for Go, TypeScript, Python, and Rust. Use when writing tests first, following Red-Green-Refactor, or enforcing TDD discipline.
allowed-tools:
  - Read
  - Glob
  - Grep
  - Write
  - Edit
  - Bash
---

# Test-Driven Development Workflow

## Purpose / When to Use

Use this skill when:
- Writing tests first (Red phase)
- Making tests pass with minimal code (Green phase)
- Refactoring while keeping tests green (Refactor phase)
- Need TDD patterns for a specific language
- Want to enforce TDD discipline

## The Red-Green-Refactor Cycle

```
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│   ┌─────────┐     ┌─────────┐     ┌───────────┐            │
│   │   RED   │ ──► │  GREEN  │ ──► │ REFACTOR  │ ──┐        │
│   │  Write  │     │ Minimal │     │  Improve  │   │        │
│   │ Failing │     │  Code   │     │   Code    │   │        │
│   │  Test   │     │ to Pass │     │  Safely   │   │        │
│   └─────────┘     └─────────┘     └───────────┘   │        │
│        ▲                                          │        │
│        └──────────────────────────────────────────┘        │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Phase 1: RED - Write Failing Test

**Purpose:** Define the expected behavior before implementation.

**Rules:**
1. Write exactly ONE test for ONE behavior
2. Test must FAIL before writing any implementation
3. Test must fail for the RIGHT reason (not syntax/import error)
4. Test name describes the expected behavior

**Commit convention:** `test: add test for <behavior>`

### Phase 2: GREEN - Minimal Implementation

**Purpose:** Make the test pass with minimal code.

**Rules:**
1. Write ONLY enough code to make the failing test pass
2. Don't anticipate future requirements
3. Don't refactor yet — ugly code is OK
4. Stop when the test passes

**Commit convention:** `feat: implement <behavior>`

### Phase 3: REFACTOR - Improve with Safety

**Purpose:** Clean up code while tests protect you.

**Rules:**
1. Tests must pass after EVERY change
2. If tests fail, undo immediately
3. Apply design patterns, extract methods, improve names
4. Remove duplication

**Commit convention:** `refactor: improve <aspect>`

---

## Language-Specific Patterns

### Go

**Test file:** `*_test.go` (same directory, same package)
**Test command:** `go test ./...`

```go
// user_test.go
package user

import "testing"

func TestCreateUser_WithValidEmail_ReturnsUser(t *testing.T) {
    // Arrange
    email := "test@example.com"

    // Act
    user, err := CreateUser(email)

    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if user.Email != email {
        t.Errorf("got email %q, want %q", user.Email, email)
    }
}

// Table-driven test (preferred for multiple cases)
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name    string
        email   string
        wantErr bool
    }{
        {"valid email", "user@example.com", false},
        {"empty email", "", true},
        {"no at sign", "invalid", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateEmail(tt.email)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateEmail(%q) error = %v, wantErr %v",
                    tt.email, err, tt.wantErr)
            }
        })
    }
}
```

---

### TypeScript

**Test file:** `*.test.ts` or `*.spec.ts`
**Test command:** `pnpm vitest run`

```typescript
// user.test.ts
import { describe, it, expect } from 'vitest';
import { createUser, validateEmail } from './user';

describe('createUser', () => {
  it('should create user with valid email', () => {
    // Arrange
    const email = 'test@example.com';

    // Act
    const user = createUser(email);

    // Assert
    expect(user.email).toBe(email);
    expect(user.id).toBeDefined();
  });

  it('should throw for invalid email', () => {
    expect(() => createUser('')).toThrow('Invalid email');
  });
});

describe('validateEmail', () => {
  it.each([
    ['user@example.com', true],
    ['', false],
    ['invalid', false],
  ])('validateEmail(%s) should return %s', (email, expected) => {
    expect(validateEmail(email)).toBe(expected);
  });
});
```

---

### Python

**Test file:** `test_*.py` or `*_test.py`
**Test command:** `pytest`

```python
# test_user.py
import pytest
from user import create_user, validate_email

class TestCreateUser:
    def test_creates_user_with_valid_email(self):
        # Arrange
        email = "test@example.com"

        # Act
        user = create_user(email)

        # Assert
        assert user.email == email
        assert user.id is not None

    def test_raises_for_invalid_email(self):
        with pytest.raises(ValueError, match="Invalid email"):
            create_user("")

@pytest.mark.parametrize("email,expected", [
    ("user@example.com", True),
    ("", False),
    ("invalid", False),
])
def test_validate_email(email, expected):
    assert validate_email(email) == expected
```

---

### Rust

**Test file:** Same file with `#[cfg(test)]` module or `tests/*.rs`
**Test command:** `cargo test`

```rust
// user.rs
pub fn create_user(email: &str) -> Result<User, UserError> {
    if !validate_email(email) {
        return Err(UserError::InvalidEmail);
    }
    Ok(User { email: email.to_string(), id: generate_id() })
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_create_user_with_valid_email() {
        let email = "test@example.com";
        let result = create_user(email);
        assert!(result.is_ok());
        assert_eq!(result.unwrap().email, email);
    }

    #[test]
    fn test_create_user_with_invalid_email_returns_error() {
        let result = create_user("");
        assert!(result.is_err());
        assert!(matches!(result, Err(UserError::InvalidEmail)));
    }
}
```

---

## Test Naming Conventions

| Language | Convention | Example |
|----------|-----------|---------|
| Go | `Test<Function>_<Condition>_<Outcome>` | `TestCreateUser_WithInvalidEmail_ReturnsError` |
| TypeScript | `should <outcome> when <condition>` | `'should throw error when email is invalid'` |
| Python | `test_<function>_<condition>_<outcome>` | `test_create_user_with_invalid_email_raises` |
| Rust | `test_<function>_<condition>_<outcome>` | `test_create_user_invalid_email_returns_error` |

---

## Edge Cases to Always Test

### 1. Empty/Null/Zero Values
- Empty strings: ""
- Null/None values
- Zero numbers: 0, 0.0
- Empty collections: [], {}, ()

### 2. Boundary Conditions
- First/last elements
- Min/max values
- Off-by-one: n-1, n, n+1

### 3. Error Cases
- Invalid input
- Missing required fields
- Permission denied

---

## TDD Anti-Patterns to Avoid

| Anti-Pattern | Problem | Fix |
|--------------|---------|-----|
| Writing tests after code | Tests just verify existing implementation | Write test first, watch it fail |
| Testing implementation details | Tests break when refactoring | Test behavior, not implementation |
| Large tests | One test covers multiple behaviors | One test = one behavior |
| Shared mutable state | Tests affect each other | Isolate tests, use fresh fixtures |
| Skipping refactor phase | Technical debt accumulates | Always refactor after green |

---

## Guardrails

**NEVER:**
- Write production code before a failing test
- Skip the refactor phase
- Let tests share mutable state
- Test private implementation details

**ALWAYS:**
- Watch the test fail first (RED)
- Write minimal code to pass (GREEN)
- Clean up with tests protecting you (REFACTOR)
- Use descriptive test names

## Version History

- 1.0.0 — Initial release (adapted from t3chn/skills)
