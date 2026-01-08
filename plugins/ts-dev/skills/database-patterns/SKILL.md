---
name: ts-database-patterns
description: This skill should be used when the user asks about "Drizzle ORM", "TypeScript database", "type-safe queries", or mentions "database schema", "migrations", "Drizzle".
allowed-tools: [Read, Grep, Glob, Write, Edit]
---

# TypeScript Database Patterns

Type-safe database access with Drizzle ORM (2025).

## Why Drizzle (NOT Prisma)

| Aspect | Drizzle | Prisma |
|--------|---------|--------|
| Bundle size | ~100KB | 5MB+ |
| Cold start | Fast | Slow |
| Approach | SQL-first | Schema-first |
| Learning curve | Know SQL | Learn PSL |
| Performance | 2-3x faster | More overhead |

**Use Drizzle for**: Serverless, edge, new projects, SQL-familiar teams.

## Setup

```bash
pnpm add drizzle-orm postgres
pnpm add -D drizzle-kit
```

## Schema Definition

### Basic Tables

```typescript
// src/db/schema.ts
import {
  pgTable,
  serial,
  varchar,
  text,
  timestamp,
  integer,
  boolean,
  pgEnum,
} from 'drizzle-orm/pg-core';

// Enum
export const roleEnum = pgEnum('role', ['admin', 'user', 'guest']);

// Users table
export const users = pgTable('users', {
  id: serial('id').primaryKey(),
  email: varchar('email', { length: 255 }).notNull().unique(),
  name: varchar('name', { length: 100 }).notNull(),
  role: roleEnum('role').default('user').notNull(),
  isActive: boolean('is_active').default(true).notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
  updatedAt: timestamp('updated_at').defaultNow().notNull(),
});

// Posts table
export const posts = pgTable('posts', {
  id: serial('id').primaryKey(),
  title: varchar('title', { length: 255 }).notNull(),
  content: text('content'),
  authorID: integer('author_id')
    .references(() => users.id, { onDelete: 'cascade' })
    .notNull(),
  published: boolean('published').default(false).notNull(),
  createdAt: timestamp('created_at').defaultNow().notNull(),
});
```

### Relations

```typescript
// src/db/relations.ts
import { relations } from 'drizzle-orm';
import { users, posts, comments } from './schema';

export const usersRelations = relations(users, ({ many }) => ({
  posts: many(posts),
  comments: many(comments),
}));

export const postsRelations = relations(posts, ({ one, many }) => ({
  author: one(users, {
    fields: [posts.authorID],
    references: [users.id],
  }),
  comments: many(comments),
}));
```

## Database Connection

```typescript
// src/db/index.ts
import { drizzle } from 'drizzle-orm/postgres-js';
import postgres from 'postgres';
import * as schema from './schema';
import * as relations from './relations';

const connectionString = process.env.DATABASE_URL!;

const client = postgres(connectionString, {
  max: 10,
  idle_timeout: 20,
  connect_timeout: 10,
});

export const db = drizzle(client, {
  schema: { ...schema, ...relations },
});

export type Database = typeof db;
```

## Type Inference

```typescript
import { InferSelectModel, InferInsertModel } from 'drizzle-orm';
import { users, posts } from './schema';

export type User = InferSelectModel<typeof users>;
export type NewUser = InferInsertModel<typeof users>;

export type Post = InferSelectModel<typeof posts>;
export type NewPost = InferInsertModel<typeof posts>;
```

## Queries

### Select

```typescript
import { eq, and, like, gt, desc } from 'drizzle-orm';

// Simple select
const allUsers = await db.select().from(users);

// Select specific columns
const emails = await db
  .select({ id: users.id, email: users.email })
  .from(users);

// Where conditions
const activeUsers = await db
  .select()
  .from(users)
  .where(eq(users.isActive, true));

// Multiple conditions
const filteredUsers = await db
  .select()
  .from(users)
  .where(
    and(
      eq(users.role, 'admin'),
      gt(users.createdAt, new Date('2024-01-01'))
    )
  );

// Ordering and pagination
const paginatedUsers = await db
  .select()
  .from(users)
  .orderBy(desc(users.createdAt))
  .limit(10)
  .offset(0);
```

### Insert

```typescript
// Single insert
const [newUser] = await db
  .insert(users)
  .values({
    email: 'john@example.com',
    name: 'John Doe',
  })
  .returning();

// Bulk insert
await db.insert(users).values([
  { email: 'user1@example.com', name: 'User 1' },
  { email: 'user2@example.com', name: 'User 2' },
]);

// Upsert
await db
  .insert(users)
  .values({ email: 'john@example.com', name: 'John' })
  .onConflictDoUpdate({
    target: users.email,
    set: { name: 'John Updated' },
  });
```

### Update

```typescript
const [updated] = await db
  .update(users)
  .set({ name: 'New Name', updatedAt: new Date() })
  .where(eq(users.id, 1))
  .returning();
```

### Delete

```typescript
const [deleted] = await db
  .delete(users)
  .where(eq(users.id, 1))
  .returning();
```

### Relations Query (Recommended)

```typescript
// With relations (cleaner API)
const usersWithPosts = await db.query.users.findMany({
  with: {
    posts: true,
  },
});

// Nested relations
const postsWithDetails = await db.query.posts.findMany({
  with: {
    author: true,
    comments: {
      with: {
        author: true,
      },
    },
  },
  where: eq(posts.published, true),
  orderBy: desc(posts.createdAt),
  limit: 10,
});
```

## Transactions

```typescript
await db.transaction(async (tx) => {
  const [user] = await tx
    .insert(users)
    .values({ email: 'john@example.com', name: 'John' })
    .returning();

  await tx.insert(posts).values({
    title: 'First Post',
    authorID: user.id,
  });
});
```

## Migrations

### drizzle.config.ts

```typescript
import { defineConfig } from 'drizzle-kit';

export default defineConfig({
  schema: './src/db/schema.ts',
  out: './drizzle',
  dialect: 'postgresql',
  dbCredentials: {
    url: process.env.DATABASE_URL!,
  },
  verbose: true,
  strict: true,
});
```

### Commands

```bash
# Generate migration
pnpm drizzle-kit generate

# Apply migrations
pnpm drizzle-kit migrate

# Push schema (dev only)
pnpm drizzle-kit push

# Open Drizzle Studio
pnpm drizzle-kit studio
```

## Best Practices

### DO:
- Use relations query API for nested data
- Define types with `InferSelectModel`/`InferInsertModel`
- Use transactions for multi-step operations
- Index frequently queried columns
- Use migrations in production

### DON'T:
- Use `push` in production (use migrations)
- Store connection strings in code
- Ignore query performance
- Skip type inference

## Version History

- 1.0.0 — Initial release (adapted from t3chn/skills)
