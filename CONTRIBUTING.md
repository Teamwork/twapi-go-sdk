# Contributing to Teamwork API Go SDK

Thank you for your interest in contributing to the Teamwork API Go SDK! 🎉 We
welcome contributions of all kinds, including bug fixes, new features, 
documentation improvements, and examples. This document outlines the process 
for contributing to the project and helps ensure a smooth collaboration.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Development Guidelines](#development-guidelines)
- [Testing](#testing)
- [Code Style](#code-style)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)
- [Getting Help](#getting-help)

## Code of Conduct

By participating in this project, you agree to abide by our Code of Conduct. 
Please be respectful and constructive in all interactions.

## Development Guidelines

### Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.24 or later** - You can check your version with:
  ```bash
  go version
  ```
- **Git** - For version control
- **A code editor** - We recommend VS Code with the Go extension

### Initial Setup

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/twapi-go-sdk.git
   cd twapi-go-sdk
   ```

3. **Add the upstream remote** to keep your fork in sync:
   ```bash
   git remote add upstream https://github.com/teamwork/twapi-go-sdk.git
   ```

4. **Install dependencies**:
   ```bash
   go mod tidy
   ```

5. **Verify the setup** by running tests:
   ```bash
   TWAPI_SERVER=https://yourdomain.teamwork.com/ TWAPI_TOKEN=your_api_token go test -v ./...
   ```

## Development Guidelines

### Project Structure

The project is organized as follows:

```
twapi-go-sdk/
├── engine.go          # Core engine implementation
├── helpers.go         # Utility functions
├── types.go           # Common type definitions
├── examples/          # Usage examples
├── internal/          # Internal packages
│   └── browser/       # Browser integration utilities
├── projects/          # Projects API implementation
└── session/           # Authentication methods
```

### Development Workflow

1. **Create a feature branch** from the main branch:
   ```bash
   git checkout main
   git pull upstream main
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** following our coding standards
3. **Add tests** for any new functionality
4. **Run tests** to ensure everything works:
   ```bash
   TWAPI_SERVER=https://yourdomain.teamwork.com/ TWAPI_TOKEN=your_api_token go test -v ./...
   ```

5. **Run linting** to check code quality (we use [`golangci-lint`](https://golangci-lint.run/)):
   ```bash
   golangci-lint -c .golangci.yml run ./...
   ```

6. **Commit your changes** with a descriptive message:
   ```bash
   git add .
   git commit -m "Feature: Add new authentication method"
   ```

7. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

8. **Create a pull request** on GitHub

### Keeping Your Fork Updated

Regularly sync your fork with the upstream repository:

```bash
git checkout main
git pull upstream main
git push origin main
```

## Testing

We maintain high test coverage to ensure reliability. When contributing:

### Running Tests

```bash
# Run all tests
TWAPI_SERVER=https://yourdomain.teamwork.com/ TWAPI_TOKEN=your_api_token  go test -v ./...

# Run tests with coverage
TWAPI_SERVER=https://yourdomain.teamwork.com/ TWAPI_TOKEN=your_api_token  go test -v -cover ./...

# Run tests for a specific package
TWAPI_SERVER=https://yourdomain.teamwork.com/ TWAPI_TOKEN=your_api_token  go test -v ./projects

# Run a specific test
TWAPI_SERVER=https://yourdomain.teamwork.com/ TWAPI_TOKEN=your_api_token  go test -v -run TestSpecificFunction ./projects
```

### Writing Tests

- **Unit tests** should be placed alongside the code they test (e.g., `project_test.go` for `project.go`)
- **Example tests** demonstrate usage and are located in `*_example_test.go` files

## Code Style

We follow Go best practices and conventions:

### Formatting and Linting

- Use `go fmt` for consistent formatting
- Use `go vet` to catch common mistakes
- Consider using `golangci-lint` for additional checks

### Naming Conventions

- **Packages**: lowercase, single word when possible
- **Functions**: CamelCase, exported functions start with uppercase
- **Variables**: camelCase for local variables, CamelCase for exported
- **Constants**: CamelCase or UPPER_CASE for package-level constants

### Documentation

- All exported functions, types, and constants must have doc comments
- Doc comments should start with the name of the item being documented
- Use complete sentences and proper grammar

Example:
```go
// ProjectCreateRequest represents the data needed to create a new project.
type ProjectCreateRequest struct {
    Name        string `json:"name"`
    Description string `json:"description"`
}

// Create creates a new project using the provided session and request data.
// It returns the created project information or an error if the operation fails.
func Create(session *session.Session, req ProjectCreateRequest) (*Project, error) {
    // implementation
}
```

### Error Handling

- Return errors rather than panicking
- Wrap errors with context using `fmt.Errorf` or similar
- Use meaningful error messages

## Submitting Changes

### Pull Request Guidelines

1. **Title**: Use a descriptive title following conventional commits format:
   - `Feature:` for new features
   - `Fix:` for bug fixes
   - `Docs:` for documentation changes
   - `Test:` for test additions/changes
   - `Refactor:` for code refactoring
   - `Enhancement:` for improvements
   - `Chore:` for maintenance tasks

2. **Description**: Include:
   - What changes were made and why
   - Any breaking changes
   - Related issue numbers (if applicable)
   - Screenshots or examples (if relevant)

3. **Checklist**: Ensure your PR:
   - [ ] Passes all tests
   - [ ] Includes tests for new functionality
   - [ ] Updates documentation if needed
   - [ ] Follows the project's coding standards
   - [ ] Has a clear, descriptive title and description

### Review Process

- All PRs require at least one review from a maintainer
- Be responsive to feedback and questions
- Make requested changes promptly
- Keep discussions constructive and professional

## Reporting Issues

### Bug Reports

When reporting bugs, please include:

1. **Go version**: `go version` output
2. **SDK version**: Version or commit hash being used
3. **Operating system**: e.g., macOS 15.1, Ubuntu 22.04
4. **Expected behavior**: What should happen
5. **Actual behavior**: What actually happens
6. **Reproduction steps**: Minimal code to reproduce the issue
7. **Error messages**: Full error output if applicable

### Feature Requests

When requesting features:

1. **Use case**: Describe the problem you're trying to solve
2. **Proposed solution**: How you think it should work
3. **Alternatives**: Other solutions you've considered
4. **API compatibility**: Consider impact on existing users

## Getting Help

If you need help or have questions:

- **GitHub Discussions**: [Ask questions and discuss ideas](https://github.com/teamwork/twapi-go-sdk/discussions)
- **GitHub Issues**: [Report bugs or request features](https://github.com/teamwork/twapi-go-sdk/issues)
- **Documentation**: Check the [API documentation](https://apidocs.teamwork.com/) and [examples](examples/)

## Recognition

Contributors will be recognized in our changelog. Thank you for helping make
this project better! 🙏