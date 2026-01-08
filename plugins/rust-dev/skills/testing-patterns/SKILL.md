---
name: testing-patterns
description: Use when the user asks about "Rust tests", "cargo test", "write tests for Rust", "rstest", "mockall", "proptest", "#[test]", "async tests with tokio", or needs guidance on Rust testing patterns.
version: 1.0.0
globs: ["*.rs", "Cargo.toml"]
---

# Rust Testing

Idiomatic Rust testing patterns and practices.

## Test Locations

### 1. Same File (Unit Tests) — for private function testing

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_internal_function() {
        // Can access private functions
    }
}
```

### 2. `tests/` Directory (Integration Tests) — for public API testing

```
tests/
├── integration_test.rs
└── common/
    └── mod.rs
```

### 3. Doc Tests — for documentation examples

```rust
/// Returns the sum of two numbers.
///
/// # Examples
///
/// ```
/// let result = mylib::add(2, 3);
/// assert_eq!(result, 5);
/// ```
pub fn add(a: i32, b: i32) -> i32 {
    a + b
}
```

## Dependencies

```toml
# Cargo.toml
[dev-dependencies]
rstest = "0.18"        # Parametrized tests, fixtures
mockall = "0.12"       # Mocking
proptest = "1.4"       # Property-based testing
tokio = { version = "1", features = ["test-util", "macros", "rt-multi-thread"] }
```

## Patterns

### Basic Test (AAA Pattern)

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_function_with_valid_input() {
        // Arrange
        let input = "test";

        // Act
        let result = process(input);

        // Assert
        assert_eq!(result, expected);
    }
}
```

### Testing Panics

```rust
#[test]
#[should_panic]
fn test_panics_on_invalid_input() {
    process(None);
}

#[test]
#[should_panic(expected = "invalid input")]
fn test_panics_with_message() {
    process("invalid");
}
```

### Testing Result

```rust
#[test]
fn test_returns_error_on_invalid() {
    let result = validate("");

    assert!(result.is_err());
    assert_eq!(
        result.unwrap_err().to_string(),
        "Input cannot be empty"
    );
}

#[test]
fn test_returns_ok_on_valid() {
    let result = validate("valid");

    assert!(result.is_ok());
    assert_eq!(result.unwrap(), "validated");
}
```

### Testing Option

```rust
#[test]
fn test_returns_none_when_not_found() {
    let result = find_user(999);
    assert!(result.is_none());
}

#[test]
fn test_returns_some_when_found() {
    let result = find_user(1);
    assert!(result.is_some());
    assert_eq!(result.unwrap().name, "Alice");
}
```

### Async Tests (Tokio)

```rust
#[tokio::test]
async fn test_async_operation() {
    let result = fetch_data().await;
    assert!(result.is_ok());
}

// With timeout control
#[tokio::test(start_paused = true)]
async fn test_with_timeout() {
    tokio::time::timeout(
        Duration::from_secs(5),
        async_operation()
    ).await.unwrap();
}
```

### Parametrized Tests (rstest)

```rust
use rstest::rstest;

#[rstest]
#[case("hello", 5)]
#[case("", 0)]
#[case("rust", 4)]
fn test_string_length(#[case] input: &str, #[case] expected: usize) {
    assert_eq!(input.len(), expected);
}
```

### Fixtures (rstest)

```rust
use rstest::*;

#[fixture]
fn user() -> User {
    User::new("Test", "test@example.com")
}

#[fixture]
fn db() -> TestDb {
    TestDb::new(":memory:")
}

#[rstest]
fn test_user_save(user: User, db: TestDb) {
    db.save(&user);
    assert!(db.exists(user.id));
}
```

### Mocking (mockall)

```rust
use mockall::automock;

#[automock]
trait Database {
    fn get_user(&self, id: u32) -> Option<User>;
}

#[test]
fn test_with_mock() {
    let mut mock = MockDatabase::new();
    mock.expect_get_user()
        .with(eq(1))
        .times(1)
        .returning(|_| Some(User::new("Test")));

    let service = UserService::new(mock);
    let result = service.find_user(1);

    assert!(result.is_some());
}
```

### Property-Based Testing (proptest)

```rust
use proptest::prelude::*;

proptest! {
    #[test]
    fn test_parse_roundtrip(s in "\\PC*") {
        let parsed = parse(&s);
        let formatted = format(&parsed);
        prop_assert_eq!(s, formatted);
    }

    #[test]
    fn test_addition_commutative(a in any::<i32>(), b in any::<i32>()) {
        prop_assert_eq!(add(a, b), add(b, a));
    }
}
```

## Edge Cases to Test

### Empty/Zero

```rust
#[test]
fn test_empty_string() {
    assert!(process("").is_err());
}

#[test]
fn test_zero_value() {
    assert_eq!(calculate(0), 0);
}
```

### Boundary Conditions

```rust
#[test]
fn test_max_value() {
    assert!(process(i32::MAX).is_ok());
}

#[test]
fn test_min_value() {
    assert!(process(i32::MIN).is_ok());
}
```

### Thread Safety

```rust
#[test]
fn test_concurrent_access() {
    use std::sync::{Arc, Mutex};
    use std::thread;

    let counter = Arc::new(Mutex::new(0));
    let handles: Vec<_> = (0..10)
        .map(|_| {
            let counter = Arc::clone(&counter);
            thread::spawn(move || {
                *counter.lock().unwrap() += 1;
            })
        })
        .collect();

    for handle in handles {
        handle.join().unwrap();
    }

    assert_eq!(*counter.lock().unwrap(), 10);
}
```

## Test Organization

```
src/
├── lib.rs
├── user.rs           # Unit tests in #[cfg(test)] mod tests
├── order.rs
└── payment.rs

tests/
├── common/
│   └── mod.rs        # Shared test utilities
├── user_integration.rs
└── order_flow.rs
```

## Running Tests

```bash
# Run all tests
cargo test

# With output
cargo test -- --nocapture

# Specific test
cargo test test_name

# Specific module
cargo test user::tests

# Only doc tests
cargo test --doc

# Only integration tests
cargo test --test integration_test

# Ignored tests
cargo test -- --ignored

# All tests including ignored
cargo test -- --include-ignored
```

## Best Practices

### DO:
- Use descriptive test names: `test_<function>_<scenario>_<expected>`
- Test one behavior per test
- Use `assert_eq!` over `assert!` when comparing values
- Group related tests in modules
- Use fixtures (`rstest`) for complex setup
- Test error messages, not just error types
- Use `expect("context")` instead of bare `unwrap()` in tests

### DON'T:
- Test private implementation details
- Share mutable state between tests
- Use `thread::sleep` for timing (use tokio test utilities)
- Over-mock (prefer real implementations when fast)
- Ignore test failures
