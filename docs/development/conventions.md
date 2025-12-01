---
layout: default
title: Conventions
parent: Development
nav_order: 4
---
# Conventions
Coding conventions and best practices for Mithril development.

## Code Style
All Go code must adhere to the following standards:

- Use `gofmt` for automatic code formatting: https://pkg.go.dev/cmd/gofmt
- Follow the official Go guidelines: https://go.dev/doc/effective_go
- Use Google's Go style decisions: https://google.github.io/styleguide/go/decisions
- Run `go vet` to catch common mistakes
- Use `golint` or `golangci-lint` in CI/CD for code quality checks

**Pre-commit setup:** Configure your editor to run `gofmt` on save to maintain consistency across the team.

## Naming Conventions

### Packages
- Use short, concise names in lowercase (no underscores)
- Avoid generic names like `util`, `helper`, or `common`
- Examples: `parser`, `client`, `config`, `storage`

### Functions and Methods
- Use `CamelCase` for exported functions (start with uppercase)
- Use `camelCase` for unexported functions (start with lowercase)
- Use descriptive names: `ParseConfig()` not `Parse()`
- Boolean functions should start with `Is`, `Has`, or `Can`: `IsValid()`, `HasPermission()`

### Variables and Constants
- Use short, meaningful variable names within scope
- Exported constants: `CONSTANT_NAME` or `ConstantName` depending on scope
- Unexported constants: `lowercase` or `lowerCamelCase`
- Avoid abbreviations unless widely understood (e.g., `ctx` for context is acceptable)

### Interfaces
- Single method interfaces should be named with `-er` suffix: `Reader`, `Writer`, `Closer`
- Multi-method interfaces use descriptive names: `RequestHandler`, `DataStore`

### Errors
- Define error variables with `Err` prefix: `var ErrNotFound = errors.New("not found")`
- Wrap errors with context: `fmt.Errorf("failed to process user: %w", err)`

## Git Workflow

### Branch Strategy
We follow a **Git Flow** model adapted for our team of 4:

#### Branch Types
- **`main`**: Production-ready code. All commits must be tagged with releases.
- **`develop`**: Integration branch. This is the base for feature development.
- **`feature/*`**: Feature branches for new functionality
  - Naming: `feature/brief-description` (e.g., `feature/add-user-auth`)
  - Created from: `develop`
  - Merged back to: `develop` via pull request
- **`bugfix/*`**: Bug fix branches
  - Naming: `bugfix/issue-number-brief-desc` (e.g., `bugfix/123-crash-on-startup`)
  - Created from: `develop`
  - Merged back to: `develop` via pull request
- **`hotfix/*`**: Critical fixes for production
  - Naming: `hotfix/issue-number-brief-desc`
  - Created from: `main`
  - Merged back to: `main` AND `develop`

### Commit Messages
Follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/):

**Format:** `<type>(<scope>): <subject>`

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, gofmt)
- `refactor`: Code refactoring without functional changes
- `perf`: Performance improvements
- `test`: Adding or updating tests
- `ci`: CI/CD configuration changes
- `chore`: Build, dependency updates, tooling

**Examples:**
```
feat(auth): add JWT token validation
fix(parser): handle malformed JSON gracefully
docs(readme): update installation instructions
test(storage): add unit tests for cache layer
```

**Commit Body (for significant commits):**
- Explain *why*, not *what* (the what should be clear from the code)
- Keep lines under 72 characters
- Use imperative mood: "add feature" not "added feature"
- Reference issues: `Closes #123` or `Fixes #456`

**Example full commit:**
```
feat(storage): add redis cache support

Implement Redis caching for frequently accessed objects to reduce
database load. Includes configurable TTL and automatic cleanup.

Closes #45
```

### Pull Request Workflow

1. **Create a feature branch** from `develop`:
   ```bash
   git checkout develop
   git pull origin develop
   git checkout -b feature/your-feature-name
   ```

2. **Commit regularly** with meaningful messages (follow Conventional Commits above)

3. **Push your branch** and create a pull request:
   ```bash
   git push origin feature/your-feature-name
   ```

4. **PR Requirements** before merging:
   - ✅ All CI/CD checks pass (tests, linting, build)
   - ✅ At least 2 code reviews from team members
   - ✅ All conversations resolved
   - ✅ Branch is up to date with `develop`
   - ✅ Squash commits if history is messy (optional, team decides per PR)

5. **Merge** using "Squash and merge" or "Create a merge commit" (team consensus needed)

### Code Review Checklist
When reviewing PRs, check for:
- Code follows style guidelines (`gofmt`, naming conventions)
- Changes are well-tested with unit/integration tests
- Documentation updated if needed
- No hardcoded values or credentials
- Error handling is appropriate
- Performance impact considered
- Commit messages are clear and follow conventions

### Releasing
1. Create a release branch from `main`:
   ```bash
   git checkout main
   git pull origin main
   git checkout -b release/v1.2.3
   ```

2. Update version numbers and changelog

3. Create a PR to `main`, merge after approval

4. Tag the commit:
   ```bash
   git tag -a v1.2.3 -m "Release version 1.2.3"
   git push origin v1.2.3
   ```

5. Merge back to `develop` to keep in sync

## Documentation Standards

### Code Documentation
- Every exported function and type must have a doc comment
- Doc comments should start with the name of the entity: `// ParseConfig parses...`
- Use proper grammar and complete sentences
- Example:

```go
// User represents a system user with authentication credentials.
// Fields are unexported except for ID which is used in APIs.
type User struct {
    id    string
    name  string
    email string
}

// NewUser creates a new User with the given name and email.
// It returns an error if email is invalid.
func NewUser(name, email string) (*User, error) {
    // implementation
}
```

### README Requirements
Each package or module should have a README explaining:
- Purpose of the package
- Key components and interfaces
- Usage examples
- Dependencies
- Configuration options

### Architecture Documentation
- Maintain a high-level architecture document explaining component relationships
- Update when major changes occur
- Include diagrams where helpful (use Mermaid, ASCII art, or images)

### Changelog
- Maintain a `CHANGELOG.md` file following [Keep a Changelog](https://keepachangelog.com/)
- Update with each release
- Categorize entries: Added, Changed, Deprecated, Removed, Fixed, Security

### Inline Comments
- Comment the *why*, not the *what*
- Avoid obvious comments: `i = i + 1  // increment i` is unnecessary
- Useful comment: `// Use regex to validate email early to fail fast`
- Keep comments up to date with code changes

### Issue and PR Templates
Use GitHub templates for consistency:
- **Issues**: Bug reports should include reproduction steps, expected vs actual behavior
- **PRs**: Reference related issues, describe changes, highlight any breaking changes

## Tools and Automation

### Required Tools
- Go 1.21+ (check `go.mod`)
- `gofmt` (built-in with Go)
- `go vet` (built-in with Go)
- `git` with SSH keys configured

### Recommended Tools
- `golangci-lint`: Comprehensive linting (run locally before pushing)
- `pre-commit`: Git hook framework to enforce checks before commits
- Editor: VS Code with Go extension, GoLand, or vim with vim-go

### CI/CD
All pushes and PRs must pass:
- Go build (`go build ./...`)
- Tests with coverage (`go test -cover ./...`)
- Linting (`golangci-lint run`)
- Go vet (`go vet ./...`)

---
