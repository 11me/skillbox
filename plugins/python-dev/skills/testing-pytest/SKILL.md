---
name: testing-pytest
description: Use when the user asks about "pytest", "Python tests", "write tests", "test fixtures", "mocking", "async tests", "parametrize", "test coverage", or needs guidance on Python testing best practices.
version: 1.0.0
globs: ["*.py", "pyproject.toml", "pytest.ini", "conftest.py"]
---

# Python Testing

Modern Python testing with pytest ecosystem.

## Tooling

**Use:**
- `pytest` - test runner
- `pytest-cov` - coverage
- `pytest-asyncio` - async tests
- `pytest-mock` - mocking (wraps unittest.mock)
- `hypothesis` - property-based testing (when appropriate)
- `respx` - HTTP mocking for httpx
- `aioresponses` - HTTP mocking for aiohttp

**Avoid:**
- `unittest` style (use pytest native)
- `nose` (deprecated)
- `mock` standalone (use pytest-mock)

## Quick Start

### Install

```bash
uv add --dev pytest pytest-cov pytest-asyncio pytest-mock
```

### pyproject.toml

```toml
[tool.pytest.ini_options]
testpaths = ["tests"]
asyncio_mode = "auto"
asyncio_default_fixture_loop_scope = "function"
addopts = ["-ra", "-q", "--strict-markers", "--strict-config"]
markers = [
    "slow: marks tests as slow",
    "integration: marks tests as integration tests",
]

[tool.coverage.run]
source = ["src"]
branch = true

[tool.coverage.report]
exclude_lines = [
    "pragma: no cover",
    "if TYPE_CHECKING:",
    "raise NotImplementedError",
]
```

### Directory Structure

```
project/
├── src/
│   └── mypackage/
│       ├── __init__.py
│       └── service.py
└── tests/
    ├── conftest.py          # Shared fixtures
    ├── unit/
    │   └── test_service.py
    └── integration/
        └── test_api.py
```

## Test Naming

```python
# Pattern: test_<function>_<scenario>_<expected>
def test_calculate_total_with_discount_returns_reduced_price():
    ...

def test_fetch_user_invalid_id_raises_not_found():
    ...
```

## Patterns

### Basic Test (AAA Pattern)

```python
def test_user_creation():
    # Arrange
    user_data = {"name": "John", "email": "john@example.com"}
    service = UserService()

    # Act
    result = service.create(user_data)

    # Assert
    assert result.id is not None
    assert result.name == "John"
```

### Parametrized Tests

```python
import pytest

@pytest.mark.parametrize("input,expected", [
    ("hello", "HELLO"),
    ("World", "WORLD"),
    ("", ""),
])
def test_uppercase(input, expected):
    assert input.upper() == expected
```

### Fixtures

```python
# tests/conftest.py
import pytest

@pytest.fixture
def sample_user():
    """Simple data fixture."""
    return {"id": 1, "name": "Test User", "email": "test@example.com"}


@pytest.fixture
def db():
    """Setup/teardown fixture."""
    database = Database(":memory:")
    database.connect()
    yield database
    database.disconnect()


@pytest.fixture(scope="module")
def expensive_resource():
    """Shared across module (use sparingly)."""
    resource = create_expensive_resource()
    yield resource
    resource.cleanup()
```

### Async Tests

```python
# With asyncio_mode = "auto", no decorator needed
async def test_fetch_user():
    user = await fetch_user(1)
    assert user["id"] == 1


@pytest.fixture
async def async_client():
    async with AsyncClient() as client:
        yield client


async def test_with_async_client(async_client):
    response = await async_client.get("/users")
    assert response.status_code == 200
```

### Mocking

```python
from unittest.mock import AsyncMock

def test_send_email(mocker):
    """Mock external service."""
    mock_send = mocker.patch("mypackage.email.send_email")
    mock_send.return_value = True

    result = notify_user("test@example.com", "Hello")

    assert result is True
    mock_send.assert_called_once_with("test@example.com", "Hello")


async def test_external_api(mocker):
    """Mock async function."""
    mock_fetch = mocker.patch(
        "mypackage.client.fetch_data",
        new_callable=AsyncMock,
        return_value={"data": "mocked"}
    )

    result = await process_data()

    assert result["data"] == "mocked"
    mock_fetch.assert_awaited_once()
```

### HTTP Mocking (httpx with respx)

```python
import respx
import httpx

@respx.mock
async def test_api_call():
    respx.get("https://api.example.com/users/1").respond(
        json={"id": 1, "name": "John"}
    )

    async with httpx.AsyncClient() as client:
        response = await client.get("https://api.example.com/users/1")

    assert response.json()["name"] == "John"
```

### Exception Testing

```python
import pytest

def test_invalid_email_raises():
    with pytest.raises(ValueError) as exc_info:
        validate_email("not-an-email")

    assert "Invalid email format" in str(exc_info.value)
```

### Markers

```python
import pytest

@pytest.mark.slow
def test_complex_calculation():
    """Run with: pytest -m slow"""
    result = heavy_computation()
    assert result is not None


@pytest.mark.integration
async def test_database_connection():
    """Run with: pytest -m integration"""
    async with get_connection() as conn:
        assert await conn.ping()


@pytest.mark.skip(reason="Not implemented yet")
def test_future_feature():
    pass
```

## Running Tests

```bash
# Run all tests
pytest

# With coverage
pytest --cov --cov-report=term-missing

# Specific file/test
pytest tests/unit/test_service.py
pytest tests/unit/test_service.py::test_specific_function

# By marker
pytest -m "not slow"
pytest -m integration

# Verbose with print output
pytest -v -s

# Stop on first failure
pytest -x

# Run last failed
pytest --lf
```

## Coverage

```bash
# Terminal report
pytest --cov=src --cov-report=term-missing

# HTML report
pytest --cov=src --cov-report=html

# Fail if below threshold
pytest --cov=src --cov-fail-under=80
```

## Best Practices

### DO:
- One assertion per test (usually)
- Descriptive test names: `test_<what>_<condition>_<expected>`
- Use fixtures for setup/teardown
- Test edge cases: empty, None, negative, boundary values
- Test error paths, not just happy paths
- Keep tests fast (mock external services)

### DON'T:
- Test implementation details
- Use `time.sleep()` in tests
- Share state between tests
- Test private methods directly
- Write tests that depend on execution order
