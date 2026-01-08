---
name: ts-conventions
description: TypeScript conventions, type-safe APIs (Hono, tRPC, Zod), and best practices. Use when reviewing TypeScript code, building APIs, checking conventions, or writing TypeScript.
allowed-tools: [Read, Grep, Glob, Write, Edit]
---

# TypeScript Conventions

Code review context for TypeScript projects.

## Type Safety

### Prefer Strict Types Over `any`

```typescript
// WRONG
function process(data: any): any

// CORRECT
function process<T extends Record<string, unknown>>(data: T): ProcessedData<T>
```

### Use Type Guards

```typescript
function isUser(value: unknown): value is User {
  return (
    typeof value === 'object' &&
    value !== null &&
    'id' in value &&
    'email' in value
  );
}
```

### Discriminated Unions

```typescript
type Result<T, E = Error> =
  | { success: true; data: T }
  | { success: false; error: E };

// Usage
if (result.success) {
  console.log(result.data);  // TypeScript knows data exists
}
```

## Error Handling

### Custom Error Classes

```typescript
class AppError extends Error {
  constructor(
    message: string,
    public code: string,
    public statusCode: number = 500
  ) {
    super(message);
    this.name = 'AppError';
  }
}

class NotFoundError extends AppError {
  constructor(resource: string) {
    super(`${resource} not found`, 'NOT_FOUND', 404);
  }
}
```

### Result Pattern (No Exceptions for Flow Control)

```typescript
type Result<T, E = Error> =
  | { ok: true; value: T }
  | { ok: false; error: E };

async function fetchUser(id: string): Promise<Result<User>> {
  try {
    const user = await db.users.findUnique({ where: { id } });
    if (!user) {
      return { ok: false, error: new NotFoundError('User') };
    }
    return { ok: true, value: user };
  } catch (e) {
    return { ok: false, error: e as Error };
  }
}
```

## Naming Conventions

| Type | Convention | Example |
|------|------------|---------|
| Interfaces | PascalCase, no `I` prefix | `User`, `ApiResponse` |
| Types | PascalCase | `UserRole`, `RequestConfig` |
| Functions | camelCase | `getUserByID`, `parseConfig` |
| Constants | SCREAMING_SNAKE | `MAX_RETRIES`, `API_URL` |
| React Components | PascalCase | `UserProfile`, `DataTable` |
| Hooks | use* prefix | `useAuth`, `useLocalStorage` |

## Modern Patterns

### Zod for Runtime Validation

```typescript
import { z } from 'zod';

const UserSchema = z.object({
  id: z.string().uuid(),
  email: z.string().email(),
  role: z.enum(['admin', 'user']),
});

type User = z.infer<typeof UserSchema>;

// Validate at runtime
const user = UserSchema.parse(unknownData);
```

### Branded Types for Type Safety

```typescript
type UserID = string & { readonly brand: unique symbol };
type OrderID = string & { readonly brand: unique symbol };

function createUserID(id: string): UserID {
  return id as UserID;
}

// Now getUserByID(orderID) is a compile error!
```

## Async Patterns

### Avoid Floating Promises

```typescript
// WRONG
button.onClick = () => {
  fetchData();  // Floating promise!
};

// CORRECT
button.onClick = () => {
  void fetchData();  // Explicit void
};

// Or handle errors
button.onClick = () => {
  fetchData().catch(handleError);
};
```

### Use Promise.all for Parallel

```typescript
// Sequential (slow)
const user = await getUser(id);
const orders = await getOrders(id);

// Parallel (fast)
const [user, orders] = await Promise.all([
  getUser(id),
  getOrders(id),
]);
```

## Security

- Validate all external input with Zod
- Use parameterized queries (never string SQL)
- Escape HTML output (React does this automatically)
- Never store secrets in client code
- Use HTTPS for all API calls

## Code Review Checklist

- [ ] No `any` types (use `unknown` if needed)
- [ ] Errors handled with Result pattern or try/catch
- [ ] No floating promises
- [ ] Input validated at boundaries (Zod)
- [ ] Types exported from dedicated files
- [ ] Tests cover happy path and error cases
- [ ] No hardcoded secrets or URLs
- [ ] Discriminated unions for state

## API Patterns

### Hono — Edge-First Web Framework

```typescript
import { Hono } from 'hono';
import { cors } from 'hono/cors';
import { logger } from 'hono/logger';
import { zValidator } from '@hono/zod-validator';
import { z } from 'zod';

const createUserSchema = z.object({
  name: z.string().min(1).max(100),
  email: z.string().email(),
});

const users = new Hono()
  .get('/', async (c) => {
    const users = await db.select().from(usersTable);
    return c.json(users);
  })
  .post(
    '/',
    zValidator('json', createUserSchema),
    async (c) => {
      const body = c.req.valid('json'); // Fully typed!
      const user = await createUser(body);
      return c.json(user, 201);
    }
  );

const app = new Hono()
  .use('*', logger())
  .use('*', cors())
  .route('/users', users);

export type AppType = typeof app;
```

### tRPC — End-to-End Type Safety

```typescript
// server/trpc.ts
import { initTRPC, TRPCError } from '@trpc/server';
import { z } from 'zod';

const t = initTRPC.context<Context>().create();

export const router = t.router;
export const publicProcedure = t.procedure;

export const protectedProcedure = t.procedure.use(async ({ ctx, next }) => {
  if (!ctx.user) throw new TRPCError({ code: 'UNAUTHORIZED' });
  return next({ ctx: { user: ctx.user } });
});

// Router definition
export const userRouter = router({
  getByID: publicProcedure
    .input(z.object({ id: z.number() }))
    .query(async ({ input }) => {
      const user = await findUser(input.id);
      if (!user) throw new TRPCError({ code: 'NOT_FOUND' });
      return user;
    }),

  create: protectedProcedure
    .input(z.object({ name: z.string().min(1), email: z.string().email() }))
    .mutation(async ({ input, ctx }) => {
      return await createUser({ ...input, createdBy: ctx.user.id });
    }),
});
```

### Advanced Zod Patterns

```typescript
import { z } from 'zod';

// Transforms
const dateSchema = z.string().transform((str) => new Date(str));

// Refinements
const passwordSchema = z.string()
  .min(8)
  .refine((val) => /[A-Z]/.test(val), 'Must contain uppercase')
  .refine((val) => /[0-9]/.test(val), 'Must contain number');

// Discriminated Union
const responseSchema = z.discriminatedUnion('status', [
  z.object({ status: z.literal('success'), data: userSchema }),
  z.object({ status: z.literal('error'), error: z.string() }),
]);

// Coercion for query params
const querySchema = z.object({
  page: z.coerce.number().default(1),
  limit: z.coerce.number().default(10),
});
```

## Related Skills

- **ts-project-setup** — Project structure and tooling
- **ts-database-patterns** — Drizzle ORM patterns
- **ts-testing-patterns** — Vitest testing
- **ts-type-patterns** — Advanced TypeScript types

## Version History

- 1.1.0 — Merged api-patterns (Hono, tRPC, advanced Zod)
- 1.0.0 — Initial release (adapted from t3chn/skills)
