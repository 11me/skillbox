# Conventional Commits Reference

Full specification for Conventional Commits 1.0.0.

## Specification

The commit message should be structured as follows:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

### Structure

1. **type**: Required. Describes the category of change.
2. **scope**: Optional. Noun describing the section of codebase.
3. **description**: Required. Short summary of the change.
4. **body**: Optional. Detailed explanation.
5. **footer**: Optional. References, breaking changes.

## Types

### Primary Types (SemVer relevant)

| Type | Description | SemVer |
|------|-------------|--------|
| `feat` | New feature | MINOR |
| `fix` | Bug fix | PATCH |

### Secondary Types

| Type | Description |
|------|-------------|
| `build` | Changes to build system or dependencies |
| `chore` | Maintenance tasks, no production code |
| `ci` | CI/CD configuration changes |
| `docs` | Documentation only changes |
| `perf` | Performance improvements |
| `refactor` | Code change without fix or feature |
| `revert` | Reverts a previous commit |
| `style` | Formatting, whitespace, semicolons |
| `test` | Adding or correcting tests |

## Breaking Changes

Two ways to indicate breaking changes:

### 1. Exclamation Mark

Add `!` after type/scope:

```
feat(api)!: send email on subscription change
```

### 2. Footer

Add `BREAKING CHANGE:` footer:

```
feat(api): allow users to provide custom templates

BREAKING CHANGE: `template` option now required in config
```

Both methods can be combined. Breaking changes trigger MAJOR version bump.

## Scope Guidelines

Scope should be a noun describing the affected area:

### Common Scopes

| Scope | Usage |
|-------|-------|
| `api` | API endpoints, REST/GraphQL |
| `auth` | Authentication, authorization |
| `cli` | Command-line interface |
| `config` | Configuration handling |
| `core` | Core functionality |
| `db` | Database, migrations |
| `deps` | Dependencies |
| `docs` | Documentation |
| `i18n` | Internationalization |
| `perf` | Performance |
| `security` | Security-related |
| `test` | Testing infrastructure |
| `ui` | User interface |

### Language/Framework Specific

**Go:**
- `cmd` - CLI commands
- `internal` - Internal packages
- `pkg` - Public packages

**JavaScript/TypeScript:**
- `components` - React/Vue components
- `hooks` - React hooks
- `store` - State management
- `utils` - Utility functions

**Python:**
- `models` - Data models
- `views` - View functions
- `tests` - Test modules

## Description Rules

1. Use imperative, present tense: "add" not "added" or "adds"
2. Don't capitalize first letter
3. No period at the end
4. Maximum 50 characters (soft limit, 72 hard)

### Good Examples

- `add user authentication`
- `fix memory leak in connection pool`
- `remove deprecated API endpoints`
- `update dependencies to latest versions`

### Bad Examples

- `Added user authentication` (past tense)
- `Fix memory leak.` (period at end)
- `REMOVE DEPRECATED API ENDPOINTS` (caps)

## Body Guidelines

1. Separated from subject by blank line
2. Wrap at 72 characters
3. Explain what and why, not how
4. Can include bullet points

### Example

```
fix(auth): prevent race condition in token refresh

The previous implementation could cause multiple simultaneous
refresh requests when the token expired. This led to:

- Wasted API calls
- Potential rate limiting
- Inconsistent token state

Added mutex lock around the refresh logic to ensure only one
refresh happens at a time.
```

## Footer Guidelines

Footers follow `git trailer` format: `token: value` or `token #value`

### Common Footers

| Footer | Purpose |
|--------|---------|
| `BREAKING CHANGE:` | Describes breaking changes |
| `Fixes #123` | Closes GitHub issue |
| `Closes #123` | Closes GitHub issue |
| `Refs #123` | References issue without closing |
| `Reviewed-by:` | Code reviewer |
| `Co-authored-by:` | Co-author attribution |

### Example

```
feat(api): add webhook support for events

Implement webhook notifications for user events.
Supports both HTTP and HTTPS endpoints.

Fixes #456
Refs #400
Co-authored-by: Jane Doe <jane@example.com>
```

## Gitmoji Reference

Complete emoji mapping for optional gitmoji support:

| Type | Emoji | Unicode | Shortcode |
|------|-------|---------|-----------|
| feat | :sparkles: | U+2728 | `:sparkles:` |
| fix | :bug: | U+1F41B | `:bug:` |
| docs | :memo: | U+1F4DD | `:memo:` |
| style | :lipstick: | U+1F484 | `:lipstick:` |
| refactor | :recycle: | U+267B | `:recycle:` |
| perf | :zap: | U+26A1 | `:zap:` |
| test | :white_check_mark: | U+2705 | `:white_check_mark:` |
| build | :package: | U+1F4E6 | `:package:` |
| ci | :construction_worker: | U+1F477 | `:construction_worker:` |
| chore | :wrench: | U+1F527 | `:wrench:` |
| revert | :rewind: | U+23EA | `:rewind:` |

### Additional Gitmoji

| Emoji | Meaning | Shortcode |
|-------|---------|-----------|
| :fire: | Remove code/files | `:fire:` |
| :truck: | Move/rename files | `:truck:` |
| :lock: | Security fix | `:lock:` |
| :bookmark: | Release/version tag | `:bookmark:` |
| :rotating_light: | Fix linter warnings | `:rotating_light:` |
| :construction: | Work in progress | `:construction:` |
| :arrow_up: | Upgrade dependencies | `:arrow_up:` |
| :arrow_down: | Downgrade dependencies | `:arrow_down:` |
| :pushpin: | Pin dependencies | `:pushpin:` |
| :globe_with_meridians: | Internationalization | `:globe_with_meridians:` |
| :pencil2: | Fix typos | `:pencil2:` |
| :poop: | Bad code (needs improvement) | `:poop:` |
| :beers: | Write drunk code | `:beers:` |
| :card_file_box: | Database changes | `:card_file_box:` |
| :loud_sound: | Add logs | `:loud_sound:` |
| :mute: | Remove logs | `:mute:` |
| :busts_in_silhouette: | Add contributors | `:busts_in_silhouette:` |
| :children_crossing: | Improve UX | `:children_crossing:` |
| :building_construction: | Architectural changes | `:building_construction:` |
| :iphone: | Responsive design | `:iphone:` |
| :clown_face: | Mock something | `:clown_face:` |
| :egg: | Easter egg | `:egg:` |
| :see_no_evil: | Add .gitignore | `:see_no_evil:` |
| :camera_flash: | Add/update snapshots | `:camera_flash:` |
| :alembic: | Experiments | `:alembic:` |
| :mag: | SEO improvements | `:mag:` |
| :label: | Add/update types | `:label:` |
| :seedling: | Add/update seed files | `:seedling:` |
| :triangular_flag_on_post: | Feature flags | `:triangular_flag_on_post:` |
| :goal_net: | Catch errors | `:goal_net:` |
| :dizzy: | Animations/transitions | `:dizzy:` |
| :wastebasket: | Deprecate code | `:wastebasket:` |
| :passport_control: | Authorization code | `:passport_control:` |
| :adhesive_bandage: | Simple fix | `:adhesive_bandage:` |
| :monocle_face: | Data exploration | `:monocle_face:` |
| :coffin: | Remove dead code | `:coffin:` |
| :test_tube: | Add failing test | `:test_tube:` |
| :necktie: | Business logic | `:necktie:` |
| :stethoscope: | Add healthcheck | `:stethoscope:` |
| :bricks: | Infrastructure | `:bricks:` |
| :technologist: | Developer experience | `:technologist:` |

## Revert Commits

When reverting:

```
revert: feat(api): add webhook support

This reverts commit abc123def.

Reason: Webhooks caused performance regression under high load.
```

## Examples by Scenario

### New Feature
```
feat(checkout): add Apple Pay support

Integrate Apple Pay as payment option in checkout flow.
Requires iOS 14+ or Safari 14+.

Closes #789
```

### Bug Fix
```
fix(parser): handle escaped quotes in JSON strings

The parser failed when JSON values contained escaped quotes.
Added proper escape sequence handling.

Fixes #234
```

### Performance
```
perf(db): add index for user lookup queries

Query time reduced from ~200ms to ~5ms for user searches.
Added composite index on (email, status, created_at).
```

### Breaking Change
```
feat(api)!: remove v1 endpoints

Remove deprecated v1 API endpoints as announced in v2.0.0.
All clients should migrate to v2 endpoints.

BREAKING CHANGE: /api/v1/* endpoints no longer available.
See migration guide: https://docs.example.com/v2-migration
```

### Documentation
```
docs(readme): add installation instructions for Windows

Added PowerShell commands and Windows-specific notes.
```

### Dependencies
```
build(deps): upgrade React to 18.2.0

- Enables concurrent rendering
- Improves hydration behavior
- Fixes memory leak in effects
```

### Multiple Scopes

If change affects multiple areas, either:

1. Use broader scope: `feat(core): ...`
2. Omit scope: `feat: ...`
3. Split into multiple commits

## Validation Regex

Pattern to validate commit messages:

```regex
^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(.+\))?!?: .{1,50}
```

Extended pattern with optional body/footer:

```regex
^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\(.+\))?!?: .{1,50}(\n\n.+)?(\n\n(BREAKING CHANGE: .+|.+: .+))*$
```
