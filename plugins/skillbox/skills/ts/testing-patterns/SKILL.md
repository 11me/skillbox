---
name: ts-testing-patterns
description: Modern TypeScript testing with Vitest. Use when writing tests, mocking, setting up test coverage, or configuring Vitest.
globs: ["**/*.test.ts", "**/*.spec.ts", "**/vitest.config.*"]
allowed-tools: Read, Grep, Glob, Write, Edit, Bash
---

# TypeScript Testing Patterns

Modern TypeScript testing with Vitest (NOT Jest) (2025).

## Setup

```bash
pnpm add -D vitest @vitest/coverage-v8
```

## Configuration

### vitest.config.ts

```typescript
import { defineConfig } from 'vitest/config';
import tsconfigPaths from 'vite-tsconfig-paths';

export default defineConfig({
  plugins: [tsconfigPaths()],
  test: {
    globals: true,                    // Use describe/it/expect globally
    environment: 'node',              // or 'jsdom' for browser
    include: ['**/*.{test,spec}.{ts,tsx}'],
    coverage: {
      provider: 'v8',
      reporter: ['text', 'json', 'html'],
      exclude: [
        'node_modules/',
        'dist/',
        '**/*.d.ts',
        '**/*.config.*',
      ],
    },
    setupFiles: ['./tests/setup.ts'],
    typecheck: {
      enabled: true,                  // Type-check test files
    },
  },
});
```

### tests/setup.ts

```typescript
import { beforeEach, vi } from 'vitest';

// Reset mocks before each test
beforeEach(() => {
  vi.clearAllMocks();
});
```

## Basic Tests

### Structure

```typescript
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { UserService } from '@/features/users';

describe('UserService', () => {
  let service: UserService;

  beforeEach(() => {
    service = new UserService();
  });

  describe('createUser', () => {
    it('should create user with valid data', async () => {
      const input = { name: 'John', email: 'john@example.com' };

      const user = await service.create(input);

      expect(user).toMatchObject({
        name: 'John',
        email: 'john@example.com',
      });
      expect(user.id).toBeDefined();
    });

    it('should throw on invalid email', async () => {
      const input = { name: 'John', email: 'invalid' };

      await expect(service.create(input)).rejects.toThrow('Invalid email');
    });
  });
});
```

### Assertions

```typescript
// Equality
expect(value).toBe(expected);           // Strict equality (===)
expect(value).toEqual(expected);        // Deep equality
expect(value).toStrictEqual(expected);  // Strict deep equality

// Truthiness
expect(value).toBeTruthy();
expect(value).toBeFalsy();
expect(value).toBeNull();
expect(value).toBeUndefined();
expect(value).toBeDefined();

// Numbers
expect(value).toBeGreaterThan(3);
expect(value).toBeLessThanOrEqual(5);
expect(value).toBeCloseTo(0.3, 5);      // Floating point

// Strings
expect(value).toMatch(/pattern/);
expect(value).toContain('substring');

// Arrays/Objects
expect(array).toContain(item);
expect(object).toHaveProperty('key');
expect(object).toMatchObject({ partial: 'match' });

// Errors
expect(() => fn()).toThrow();
expect(() => fn()).toThrow('message');
expect(() => fn()).toThrow(ErrorClass);

// Async
await expect(promise).resolves.toBe(value);
await expect(promise).rejects.toThrow('error');
```

## Mocking

### Basic Mocks with vi

```typescript
import { vi, describe, it, expect } from 'vitest';

// Mock function
const mockFn = vi.fn();
mockFn.mockReturnValue('mocked');
mockFn.mockResolvedValue('async mocked');
mockFn.mockRejectedValue(new Error('error'));

// Mock implementation
mockFn.mockImplementation((x: number) => x * 2);

// Assertions
expect(mockFn).toHaveBeenCalled();
expect(mockFn).toHaveBeenCalledTimes(2);
expect(mockFn).toHaveBeenCalledWith('arg1', 'arg2');
```

### Module Mocks

```typescript
import { vi, describe, it, expect } from 'vitest';

// Mock entire module
vi.mock('@/lib/database', () => ({
  db: {
    query: vi.fn(),
    insert: vi.fn(),
  },
}));

// Import after mock
import { db } from '@/lib/database';

describe('with mocked db', () => {
  it('should use mocked database', async () => {
    vi.mocked(db.query).mockResolvedValue([{ id: 1 }]);

    const result = await db.query('SELECT * FROM users');

    expect(result).toEqual([{ id: 1 }]);
  });
});
```

### Spies

```typescript
import { vi, describe, it, expect } from 'vitest';

const service = {
  fetchData: async () => ({ data: 'real' }),
};

describe('spying', () => {
  it('should spy on method', async () => {
    const spy = vi.spyOn(service, 'fetchData');
    spy.mockResolvedValue({ data: 'mocked' });

    const result = await service.fetchData();

    expect(spy).toHaveBeenCalled();
    expect(result.data).toBe('mocked');

    spy.mockRestore(); // Restore original
  });
});
```

## Timer Mocks

```typescript
import { vi, describe, it, expect, beforeEach, afterEach } from 'vitest';

describe('timers', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('should handle setTimeout', () => {
    const callback = vi.fn();

    setTimeout(callback, 1000);

    expect(callback).not.toHaveBeenCalled();

    vi.advanceTimersByTime(1000);

    expect(callback).toHaveBeenCalledTimes(1);
  });
});
```

## Test Organization

### File Structure

```
src/
├── features/
│   └── users/
│       ├── service.ts
│       ├── service.test.ts      # Co-located tests
│       └── repository.ts
tests/
├── setup.ts                     # Global setup
├── fixtures/                    # Test data
│   └── users.ts
└── integration/                 # Integration tests
    └── api.test.ts
```

## Running Tests

```bash
# Run all tests
pnpm vitest

# Watch mode
pnpm vitest --watch

# Run once (CI)
pnpm vitest run

# Specific file
pnpm vitest user.test.ts

# With coverage
pnpm vitest --coverage

# UI mode
pnpm vitest --ui
```

## package.json Scripts

```json
{
  "scripts": {
    "test": "vitest",
    "test:run": "vitest run",
    "test:coverage": "vitest run --coverage",
    "test:ui": "vitest --ui"
  }
}
```

## Best Practices

### DO:
- Test behavior, not implementation
- Use descriptive test names
- One assertion concept per test
- Mock external dependencies
- Use factories for test data
- Clean up after tests

### DON'T:
- Test private methods directly
- Share state between tests
- Use `time.sleep()` — use fake timers
- Test framework code
- Write flaky tests

## Version History

- 1.0.0 — Initial release (adapted from t3chn/skills)
