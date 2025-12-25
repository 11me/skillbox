---
name: ts-api-patterns
description: Type-safe APIs with Hono, tRPC, and Zod. Use when building TypeScript backends, API validation, or serverless functions.
globs: ["**/src/api/**", "**/routes/**", "**/server/**"]
allowed-tools: Read, Grep, Glob, Write, Edit
---

# TypeScript API Patterns

Type-safe APIs with Hono, tRPC, and Zod (2025).

## Hono — Edge-First Web Framework

### Setup

```bash
pnpm add hono
pnpm add -D @types/node
```

### Basic App

```typescript
import { Hono } from 'hono';
import { serve } from '@hono/node-server';

const app = new Hono();

app.get('/', (c) => c.text('Hello Hono!'));

app.get('/json', (c) => c.json({ message: 'Hello' }));

serve({ fetch: app.fetch, port: 3000 });
```

### Route Groups

```typescript
import { Hono } from 'hono';

const users = new Hono()
  .get('/', async (c) => {
    const users = await db.select().from(usersTable);
    return c.json(users);
  })
  .get('/:id', async (c) => {
    const id = c.req.param('id');
    const user = await findUser(Number(id));
    if (!user) return c.notFound();
    return c.json(user);
  })
  .post('/', async (c) => {
    const body = await c.req.json();
    const user = await createUser(body);
    return c.json(user, 201);
  });

const app = new Hono()
  .route('/users', users)
  .route('/posts', posts);

export type AppType = typeof app;
```

### Middleware

```typescript
import { Hono } from 'hono';
import { cors } from 'hono/cors';
import { logger } from 'hono/logger';
import { secureHeaders } from 'hono/secure-headers';

const app = new Hono();

// Built-in middleware
app.use('*', logger());
app.use('*', cors());
app.use('*', secureHeaders());

// Custom middleware
app.use('*', async (c, next) => {
  const start = Date.now();
  await next();
  const ms = Date.now() - start;
  c.header('X-Response-Time', `${ms}ms`);
});

// Auth middleware
app.use('/api/*', async (c, next) => {
  const token = c.req.header('Authorization')?.replace('Bearer ', '');
  if (!token) return c.json({ error: 'Unauthorized' }, 401);

  try {
    const user = await verifyToken(token);
    c.set('user', user);
    await next();
  } catch {
    return c.json({ error: 'Invalid token' }, 401);
  }
});
```

### Zod Validation with Hono

```typescript
import { Hono } from 'hono';
import { zValidator } from '@hono/zod-validator';
import { z } from 'zod';

const createUserSchema = z.object({
  name: z.string().min(1).max(100),
  email: z.string().email(),
  age: z.number().int().positive().optional(),
});

const app = new Hono()
  .post(
    '/users',
    zValidator('json', createUserSchema),
    async (c) => {
      const body = c.req.valid('json'); // Fully typed!
      const user = await createUser(body);
      return c.json(user, 201);
    }
  );
```

## tRPC — End-to-End Type Safety

### Server Setup

```typescript
// server/trpc.ts
import { initTRPC, TRPCError } from '@trpc/server';
import { z } from 'zod';

const t = initTRPC.context<Context>().create();

export const router = t.router;
export const publicProcedure = t.procedure;

export const protectedProcedure = t.procedure.use(async ({ ctx, next }) => {
  if (!ctx.user) {
    throw new TRPCError({ code: 'UNAUTHORIZED' });
  }
  return next({ ctx: { user: ctx.user } });
});
```

### Router Definition

```typescript
// server/routers/users.ts
import { z } from 'zod';
import { router, publicProcedure, protectedProcedure } from '../trpc';

export const userRouter = router({
  getByID: publicProcedure
    .input(z.object({ id: z.number() }))
    .query(async ({ input }) => {
      const user = await findUser(input.id);
      if (!user) throw new TRPCError({ code: 'NOT_FOUND' });
      return user;
    }),

  create: protectedProcedure
    .input(z.object({
      name: z.string().min(1),
      email: z.string().email(),
    }))
    .mutation(async ({ input, ctx }) => {
      return await createUser({ ...input, createdBy: ctx.user.id });
    }),
});
```

### Client Usage

```typescript
import { createTRPCClient, httpBatchLink } from '@trpc/client';
import type { AppRouter } from '../server/routers/_app';

export const trpc = createTRPCClient<AppRouter>({
  links: [
    httpBatchLink({
      url: 'http://localhost:3000/trpc',
    }),
  ],
});

// Usage - fully typed!
const user = await trpc.users.getByID.query({ id: 1 });
```

## Zod — Schema Validation

### Basic Schemas

```typescript
import { z } from 'zod';

const userSchema = z.object({
  id: z.number(),
  name: z.string().min(1).max(100),
  email: z.string().email(),
  age: z.number().int().positive().optional(),
  role: z.enum(['admin', 'user']).default('user'),
  createdAt: z.date(),
});

type User = z.infer<typeof userSchema>;
```

### Advanced Patterns

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

// Coercion
const querySchema = z.object({
  page: z.coerce.number().default(1),
  limit: z.coerce.number().default(10),
});
```

### Error Handling

```typescript
import { z, ZodError } from 'zod';

const result = schema.safeParse(data);
if (result.success) {
  console.log(result.data);
} else {
  console.log(result.error.flatten());
}
```

## Error Handling Pattern

### Result Type

```typescript
type Result<T, E = Error> =
  | { success: true; data: T }
  | { success: false; error: E };

async function getUser(id: number): Promise<Result<User>> {
  try {
    const user = await findUser(id);
    if (!user) {
      return { success: false, error: new Error('User not found') };
    }
    return { success: true, data: user };
  } catch (error) {
    return { success: false, error: error as Error };
  }
}
```

### HTTP Error Responses

```typescript
import { Hono } from 'hono';

class AppError extends Error {
  constructor(
    public statusCode: number,
    message: string,
    public code?: string
  ) {
    super(message);
  }
}

const app = new Hono();

app.onError((err, c) => {
  if (err instanceof AppError) {
    return c.json({
      error: err.message,
      code: err.code,
    }, err.statusCode);
  }

  console.error(err);
  return c.json({ error: 'Internal Server Error' }, 500);
});
```

## Version History

- 1.0.0 — Initial release (adapted from t3chn/skills)
