# Contributing to curl-impersonate-service

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR-USERNAME/curl-impersonate-service.git`
3. Create a branch: `git checkout -b feature/your-feature-name`
4. Make your changes
5. Run tests: `go test ./...`
6. Commit: `git commit -m "Description of changes"`
7. Push: `git push origin feature/your-feature-name`
8. Open a Pull Request

## Development Setup

### Prerequisites

- Go 1.21 or higher
- Docker (for building and testing the full service)
- Git

### Local Development

```bash
# Install dependencies
go mod download

# Run tests
go test -v ./...

# Build
go build -o impersonate-service .

# Run locally (requires curl-impersonate binaries)
TOKEN=test-token ./impersonate-service
```

### Docker Development

```bash
# Build Docker image
docker build -t curl-impersonate-service .

# Run with docker-compose
docker-compose up

# View logs
docker-compose logs -f

# Stop
docker-compose down
```

## Code Style

- Follow standard Go conventions
- Run `go fmt` before committing
- Use meaningful variable and function names
- Add comments for exported functions and complex logic
- Keep functions focused and small

## Testing

- Write tests for new features
- Ensure all tests pass before submitting PR
- Aim for good test coverage
- Include both positive and negative test cases

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run specific package tests
go test -v ./models/
```

## Pull Request Process

1. **Update documentation** - Update README.md if you're adding new features
2. **Add tests** - Include tests for new functionality
3. **Run tests** - Ensure all tests pass
4. **Update CHANGELOG** - Add entry describing your changes
5. **Clean commits** - Use clear, descriptive commit messages
6. **Single concern** - Each PR should address a single concern

### Commit Message Format

```
<type>: <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Adding or updating tests
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Maintenance tasks

Examples:
```
feat: add support for custom TLS settings

fix: handle timeout errors correctly

docs: update API examples in README
```

## Code Review

All submissions require review. We use GitHub pull requests for this purpose.

Reviewers will check for:
- Code quality and style
- Test coverage
- Documentation updates
- Breaking changes
- Security implications

## Reporting Bugs

When reporting bugs, please include:

1. **Description** - Clear description of the issue
2. **Steps to reproduce** - Detailed steps to reproduce the behavior
3. **Expected behavior** - What you expected to happen
4. **Actual behavior** - What actually happened
5. **Environment** - OS, Go version, Docker version
6. **Logs** - Relevant log output
7. **Screenshots** - If applicable

Use the GitHub Issues template when creating bug reports.

## Feature Requests

We welcome feature requests! Please:

1. Check if the feature has already been requested
2. Clearly describe the feature and its use case
3. Explain why this feature would be useful
4. Consider submitting a PR if you can implement it

## Security Issues

**Do not** report security vulnerabilities through public GitHub issues.

Instead, please email security concerns to the maintainers privately.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

Feel free to:
- Open a Discussion on GitHub
- Ask in the Issues section
- Contact the maintainers

Thank you for contributing! 🎉
