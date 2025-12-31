# Naming Conventions

Go naming follows simplicity and consistency principles from Effective Go.

## Packages

| Rule | Example |
|------|---------|
| Lowercase, single word | `http`, `json`, `user` |
| No underscores | `strconv` not `str_conv` |
| No mixedCaps | `httputil` not `httpUtil` |
| Singular | `user` not `users` |
| Short, clear | `ctx` not `context` (when importing) |

```go
// Good
package user
package storage
package httputil

// Bad
package users         // plural
package user_storage  // underscore
package UserStorage   // mixedCaps
```

## Functions and Methods

| Rule | Example |
|------|---------|
| MixedCaps | `GetUser`, `parseJSON` |
| Exported = Capital | `CreateUser` |
| Unexported = lowercase | `validateEmail` |
| Getters: no "Get" prefix | `user.Name()` not `user.GetName()` |
| Setters: "Set" prefix | `user.SetName()` |

```go
// Good
func CreateUser(ctx context.Context, req *CreateRequest) (*User, error)
func (u *User) Name() string        // getter
func (u *User) SetName(name string) // setter

// Bad
func Create_User()    // underscore
func (u *User) GetName() string // redundant Get for getter
```

## Interfaces

| Rule | Example |
|------|---------|
| Single method: -er suffix | `Reader`, `Writer`, `Closer` |
| Multiple methods: descriptive | `UserRepository`, `Storage` |
| No "I" prefix | `Reader` not `IReader` |

```go
// Good
type Reader interface {
    Read(p []byte) (n int, err error)
}

type UserRepository interface {
    FindByID(ctx context.Context, id string) (*User, error)
    Save(ctx context.Context, user *User) error
}

// Bad
type IReader interface { ... }  // Java-style prefix
type ReaderInterface { ... }    // redundant suffix
```

## Receivers

| Rule | Example |
|------|---------|
| Short, 1-2 letters | `u *User`, `s *Service` |
| First letter of type | `func (u *User)`, `func (c *Client)` |
| Consistent across methods | All methods use same receiver name |
| Pointer for mutations | `func (u *User) SetName()` |
| Value for read-only | `func (u User) String()` |

```go
// Good
func (u *User) SetName(name string) { u.Name = name }
func (u User) String() string { return u.Name }
func (s *Service) CreateUser(ctx context.Context) error

// Bad
func (user *User) SetName()   // too long
func (this *User) SetName()   // Java-style
func (self *User) SetName()   // Python-style
```

## Variables

| Scope | Rule | Example |
|-------|------|---------|
| Short scope | Short names | `i`, `n`, `err`, `ctx` |
| Wide scope | Descriptive | `userCount`, `httpClient` |
| Acronyms | ALL CAPS | `userID`, `httpURL`, `xmlParser` |

```go
// Good
for i := 0; i < n; i++ { ... }
if err := doSomething(); err != nil { ... }
userID := uuid.New()
httpURL := "https://example.com"

// Bad
for index := 0; index < count; index++ { ... }  // too verbose for loop
userId := uuid.New()   // wrong: should be userID
httpUrl := "..."       // wrong: should be httpURL
```

### Common Short Names

| Name | Usage |
|------|-------|
| `ctx` | context.Context |
| `err` | error |
| `i`, `j`, `k` | loop indices |
| `n` | count |
| `v` | value |
| `k` | key (in maps) |
| `ok` | boolean result |
| `b` | byte slice |
| `s` | string |
| `r` | reader |
| `w` | writer |

## Constants

| Rule | Example |
|------|---------|
| MixedCaps | `MaxRetries`, `defaultTimeout` |
| NOT SCREAMING_CASE | `MaxRetries` not `MAX_RETRIES` |
| Exported = Capital | `MaxConnections` |
| Unexported = lowercase | `defaultBufferSize` |

```go
// Good
const (
    MaxRetries       = 3
    DefaultTimeout   = 30 * time.Second
    defaultBufSize   = 4096
)

// Bad
const (
    MAX_RETRIES      = 3   // wrong: not Go style
    DEFAULT_TIMEOUT  = 30  // wrong: screaming case
)
```

## Error Variables

```go
// Package-level sentinel errors
var (
    ErrNotFound     = errors.New("not found")
    ErrInvalidInput = errors.New("invalid input")
    ErrUnauthorized = errors.New("unauthorized")
)
```

## Type Names

```go
// Good
type User struct { ... }
type CreateUserRequest struct { ... }
type UserService struct { ... }

// Bad
type UserStruct struct { ... }  // redundant suffix
type TUser struct { ... }       // unnecessary prefix
```

## File Names

| Rule | Example |
|------|---------|
| Lowercase | `user.go`, `http_client.go` |
| Underscores allowed | `user_repository.go` |
| Test files | `user_test.go` |
| Platform-specific | `file_linux.go`, `file_windows.go` |

## Summary

| Element | Convention | Example |
|---------|------------|---------|
| Package | lowercase, singular | `user` |
| Exported func | MixedCaps | `CreateUser` |
| Unexported func | mixedCaps | `validateInput` |
| Interface | -er or descriptive | `Reader`, `UserRepository` |
| Receiver | 1-2 letters | `u`, `s`, `c` |
| Acronyms | ALL CAPS | `userID`, `httpURL` |
| Constants | MixedCaps | `MaxRetries` |
| Errors | Err prefix | `ErrNotFound` |
