---
name: ts-type-patterns
description: Advanced TypeScript type patterns for type-safe applications. Use when working with generics, utility types, type guards, conditional types, or template literals.
globs: ["**/*.ts", "**/*.tsx"]
allowed-tools: Read, Grep, Glob
---

# TypeScript Type Patterns

Advanced TypeScript type patterns for type-safe applications (2025).

## Strict Mode Essentials

### Enable Everything

```json
{
  "compilerOptions": {
    "strict": true,
    "noUncheckedIndexedAccess": true,
    "noImplicitOverride": true,
    "exactOptionalPropertyTypes": true
  }
}
```

### Handle `unknown` in Catch

```typescript
try {
  riskyOperation();
} catch (error) {
  // error is unknown, not any
  if (error instanceof Error) {
    console.error(error.message);
  } else {
    console.error('Unknown error:', error);
  }
}
```

### Safe Array Access

```typescript
// With noUncheckedIndexedAccess: true
const arr = [1, 2, 3];
const first = arr[0]; // number | undefined

// Must check before use
if (first !== undefined) {
  console.log(first.toFixed(2));
}
```

## Utility Types

### Built-in Types

```typescript
interface User {
  id: number;
  name: string;
  email: string;
  role: 'admin' | 'user';
}

// Partial - all properties optional
type PartialUser = Partial<User>;

// Required - all properties required
type RequiredUser = Required<PartialUser>;

// Pick - select properties
type UserCredentials = Pick<User, 'email' | 'name'>;

// Omit - exclude properties
type PublicUser = Omit<User, 'email'>;

// Record - object type
type UserRoles = Record<string, User>;

// Readonly - immutable
type ImmutableUser = Readonly<User>;

// Extract / Exclude for unions
type AdminRole = Extract<User['role'], 'admin'>; // 'admin'
type NonAdminRole = Exclude<User['role'], 'admin'>; // 'user'
```

### Custom Utility Types

```typescript
// Make specific properties optional
type PartialBy<T, K extends keyof T> = Omit<T, K> & Partial<Pick<T, K>>;

type CreateUserInput = PartialBy<User, 'id'>; // id is optional

// Make specific properties required
type RequiredBy<T, K extends keyof T> = T & Required<Pick<T, K>>;

// Deep partial
type DeepPartial<T> = {
  [P in keyof T]?: T[P] extends object ? DeepPartial<T[P]> : T[P];
};

// Nullable
type Nullable<T> = T | null;
```

## Generics

### Basic Constraints

```typescript
// Extend constraint
function getProperty<T, K extends keyof T>(obj: T, key: K): T[K] {
  return obj[key];
}

// Multiple constraints
function merge<T extends object, U extends object>(a: T, b: U): T & U {
  return { ...a, ...b };
}

// Default type
function createArray<T = string>(length: number, value: T): T[] {
  return Array(length).fill(value);
}
```

### Generic Interfaces

```typescript
// Repository pattern
interface Repository<T, ID = number> {
  findByID(id: ID): Promise<T | null>;
  findAll(): Promise<T[]>;
  create(data: Omit<T, 'id'>): Promise<T>;
  update(id: ID, data: Partial<T>): Promise<T>;
  delete(id: ID): Promise<void>;
}
```

### Conditional Types

```typescript
// Basic conditional
type IsString<T> = T extends string ? true : false;

// Infer keyword
type ReturnTypeOf<T> = T extends (...args: any[]) => infer R ? R : never;

type ArrayElement<T> = T extends (infer E)[] ? E : never;

// Extract promise value
type Awaited<T> = T extends Promise<infer U> ? Awaited<U> : T;
```

## Type Guards

### Type Predicates (`is`)

```typescript
interface Dog {
  bark(): void;
}

interface Cat {
  meow(): void;
}

type Animal = Dog | Cat;

// Type predicate
function isDog(animal: Animal): animal is Dog {
  return 'bark' in animal;
}

// Usage
function makeSound(animal: Animal) {
  if (isDog(animal)) {
    animal.bark(); // TypeScript knows it's Dog
  } else {
    animal.meow(); // TypeScript knows it's Cat
  }
}
```

### Assertion Functions (`asserts`)

```typescript
function assertIsDefined<T>(value: T): asserts value is NonNullable<T> {
  if (value === null || value === undefined) {
    throw new Error('Value is not defined');
  }
}

function processUser(user: User | null) {
  assertIsDefined(user);
  // TypeScript now knows user is User, not null
  console.log(user.name);
}
```

### Discriminated Unions

```typescript
// Tag each variant
type Result<T, E = Error> =
  | { success: true; data: T }
  | { success: false; error: E };

function handleResult<T>(result: Result<T>) {
  if (result.success) {
    console.log(result.data);
  } else {
    console.error(result.error);
  }
}

// API response pattern
type ApiResponse<T> =
  | { status: 'loading' }
  | { status: 'success'; data: T }
  | { status: 'error'; error: string };
```

## Template Literal Types

```typescript
// String manipulation
type EventName = `on${Capitalize<string>}`;

// Dynamic keys
type CSSProperty = 'margin' | 'padding';
type CSSDirection = 'top' | 'right' | 'bottom' | 'left';
type CSSRule = `${CSSProperty}-${CSSDirection}`;
// 'margin-top' | 'margin-right' | ... | 'padding-left'
```

## Branded Types

```typescript
// Brand symbol
declare const __brand: unique symbol;

type Brand<T, B> = T & { [__brand]: B };

// Branded IDs
type UserID = Brand<number, 'UserID'>;
type PostID = Brand<number, 'PostID'>;

// Constructor functions
const UserID = (id: number): UserID => id as UserID;
const PostID = (id: number): PostID => id as PostID;

// Usage - can't mix up IDs
function getUser(id: UserID): User { /* ... */ }

const userID = UserID(1);
const postID = PostID(1);

getUser(userID); // OK
getUser(postID); // Error: PostID not assignable to UserID
```

## Const Assertions

```typescript
// Without as const
const routes1 = ['/', '/users', '/posts']; // string[]

// With as const - preserves literal types
const routes2 = ['/', '/users', '/posts'] as const;
// readonly ['/', '/users', '/posts']

type Route = (typeof routes2)[number]; // '/' | '/users' | '/posts'

// Object literal
const config = {
  env: 'production',
  port: 3000,
} as const;
// { readonly env: 'production'; readonly port: 3000 }
```

## Anti-patterns

| Wrong | Right |
|-------|-------|
| `any` | `unknown` with type guards |
| `as Type` everywhere | Proper type inference |
| `!` non-null assertion | Null checks or `??` |
| `// @ts-ignore` | Fix the type error |
| Loose function types | Strict generic constraints |

## Version History

- 1.0.0 — Initial release (adapted from t3chn/skills)
