# ForgeBase — Project Conventions

## Architecture
- Clean Architecture: domain → application → infrastructure → presentation
- Dependency Rule: inner layers know nothing of outer layers
- Manual DI via bootstrap container
- KISS: minimal abstractions, explicit code

## Code Style
- No comments on trivial code (let names speak)
- One comment per non-obvious WHY only
- Functional options for constructors with many params
- Error wrapping: `fmt.Errorf("context: %w", err)`
- Context propagated to all IO operations

## Naming
- Entities: PascalCase, singular (`User`, `Workspace`)
- Repositories: `{Entity}Repository` interface, `{Entity}RepositoryImpl` struct
- Services: `{Entity}Service`
- DTOs: suffixed with `Request`/`Response`
- Handlers: suffixed with `Handler`
- Migrations: `NNNN_direction_description.sql`

## Testing
- Unit tests: `_test.go` co-located with source
- Integration tests: `test/integration/` with testcontainers
- Mocks: `test/mocks/` generated via go generate
- Coverage target: ≥ 80%

## Commits
- Conventional commits: `feat:`, `fix:`, `refactor:`, `test:`, `chore:`
- One epic per commit when possible
- Reference epic number in commit message when applicable

## Required Go Skills
- Always load `samber/cc-skills-golang@golang-how-to` when working on Go code; it routes to the relevant skills (layout, lint, naming, code style, error handling, testing, etc.) for every task.
