# Contributing to InstaDeploy

Thank you for considering contributing to InstaDeploy! This document provides guidelines for contributing to the project.

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/yourusername/instadeploy.git`
3. Create a feature branch: `git checkout -b feature/my-feature`
4. Make your changes
5. Test your changes
6. Commit and push
7. Open a Pull Request

## Development Setup

### Prerequisites

- Go 1.22 or higher
- Docker and Docker Compose
- Git

### Building

```bash
# Download dependencies
go mod download

# Build agent
make build-agent

# Run tests
make test

# Format code
make fmt

# Run linter
make vet
```

## Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Write meaningful commit messages
- Add comments for exported functions/types
- Keep functions small and focused

## Testing

- Write tests for new features
- Ensure all tests pass before submitting PR
- Test both success and error cases
- Include integration tests where appropriate

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestValidateProjectName ./agent
```

## Pull Request Guidelines

1. **Title**: Use clear, descriptive titles
   - Good: "Add support for ARM architecture"
   - Bad: "Update code"

2. **Description**: Explain what and why
   - What problem does this solve?
   - How does it work?
   - Any breaking changes?

3. **Tests**: Include tests for new features

4. **Documentation**: Update relevant docs

5. **Commits**: Keep commits atomic and well-described

## Code Review Process

1. Automated checks must pass (tests, linting)
2. At least one maintainer approval required
3. Address review comments
4. Squash commits if requested
5. Rebase on main before merge

## Reporting Issues

### Bug Reports

Include:
- Go version
- Operating system
- Docker version
- Steps to reproduce
- Expected behavior
- Actual behavior
- Error messages/logs

### Feature Requests

Include:
- Use case description
- Proposed solution
- Alternative solutions considered
- Potential impact

## Security Issues

**Do not open public issues for security vulnerabilities.**

Email security@instadeploy.com with:
- Description of vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

## Communication

- GitHub Issues: Bug reports, feature requests
- GitHub Discussions: General questions, ideas
- Discord: Real-time chat (link in README)

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Questions?

Feel free to open a discussion or reach out to maintainers if you have any questions.

Thank you for contributing! 🎉

