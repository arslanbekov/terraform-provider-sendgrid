# Contributing to Terraform Provider for SendGrid

Thank you for your interest in contributing to the Terraform Provider for SendGrid! We welcome contributions from the community.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)

## Code of Conduct

This project adheres to a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the [issue tracker](https://github.com/arslanbekov/terraform-provider-sendgrid/issues) as you might find out that you don't need to create one.

When creating a bug report, please include:

- **Clear and descriptive title**
- **Exact steps to reproduce** the problem
- **Expected behavior** and **actual behavior**
- **Terraform version** and **provider version**
- **SendGrid API version** (if applicable)
- **Code samples** or configuration files
- **Error messages** and **logs**

Use the [Bug Report template](.github/ISSUE_TEMPLATE/bug_report.yml) when creating an issue.

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

- **Clear and descriptive title**
- **Detailed description** of the proposed functionality
- **Use cases** that would benefit from this enhancement
- **Examples** of how the feature would be used
- **Potential implementation approach** (if you have one)

Use the [Feature Request template](.github/ISSUE_TEMPLATE/feature_request.yml) when creating an issue.

### Contributing Code

We actively welcome your pull requests:

1. Fork the repo and create your branch from `master`
2. Add tests for any new code you write
3. Update documentation to reflect any changes
4. Ensure the test suite passes
5. Make sure your code follows the existing style
6. Submit your pull request

## Development Setup

### Prerequisites

- [Go](https://golang.org/doc/install) 1.25+ (see [go.mod](go.mod))
- [Terraform](https://www.terraform.io/downloads.html) 1.0+
- [Make](https://www.gnu.org/software/make/) (for using the Makefile)
- SendGrid API key with appropriate permissions for testing

### Setting Up Your Development Environment

```bash
# Clone your fork
git clone https://github.com/<your-username>/terraform-provider-sendgrid.git
cd terraform-provider-sendgrid

# Add upstream remote
git remote add upstream https://github.com/arslanbekov/terraform-provider-sendgrid.git

# Install dependencies
go mod download

# Build the provider
make build

# Run tests
make test
```

### Environment Variables

For acceptance testing, you'll need:

```bash
export SENDGRID_API_KEY="your-sendgrid-api-key"
export TF_ACC=1  # Enable acceptance tests
```

## Pull Request Process

1. **Create a feature branch**

   ```bash
   git checkout -b feature/amazing-feature
   ```

2. **Make your changes**

   - Write clear, self-documenting code
   - Add tests for new functionality
   - Update documentation as needed

3. **Commit your changes**

   ```bash
   git commit -m "feat: add amazing feature"
   ```

   Follow [Conventional Commits](https://www.conventionalcommits.org/) format:

   - `feat:` - New feature
   - `fix:` - Bug fix
   - `docs:` - Documentation changes
   - `test:` - Adding or updating tests
   - `refactor:` - Code refactoring
   - `chore:` - Maintenance tasks

4. **Push to your fork**

   ```bash
   git push origin feature/amazing-feature
   ```

5. **Open a Pull Request**

   - Use the PR template
   - Link related issues
   - Provide clear description of changes
   - Ensure CI checks pass

6. **Code Review**

   - Address review comments
   - Keep PR focused and reasonably sized
   - Be responsive to feedback

7. **Merge**
   - PRs will be merged by maintainers after approval
   - Your contribution will be included in the next release

## Coding Standards

### Go Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` for formatting
- Run `golangci-lint` before submitting
- Keep functions focused and reasonably sized
- Write clear comments for exported functions

```bash
# Format code
make fmt

# Run linter
make lint
```

### Naming Conventions

- **Variables and functions**: `camelCase`
- **Exported functions**: `PascalCase`
- **Constants**: `PascalCase` or `UPPER_CASE`
- **File names**: `snake_case.go`

### Error Handling

- Always check and handle errors
- Provide meaningful error messages
- Use `fmt.Errorf` with context
- Wrap errors when appropriate

```go
if err != nil {
    return fmt.Errorf("failed to create teammate: %w", err)
}
```

## Testing Guidelines

### Unit Tests

- Write unit tests for all new functions
- Use table-driven tests where appropriate
- Mock external dependencies
- Aim for meaningful test coverage

```bash
# Run unit tests
make test

# Run with coverage
make test-coverage
```

### Acceptance Tests

Acceptance tests run against the real SendGrid API:

- Test complete resource lifecycle (Create, Read, Update, Delete)
- Use unique identifiers to avoid conflicts
- Clean up resources after tests
- Mark tests as acceptance tests

```go
func TestAccSendGridTeammate_basic(t *testing.T) {
    // Test implementation
}
```

```bash
# Run acceptance tests (requires SENDGRID_API_KEY)
make testacc
```

### Testing Best Practices

1. **Test names** should clearly describe what they test
2. **Assertions** should be specific and meaningful
3. **Test data** should be realistic but safe
4. **Cleanup** resources properly in tests
5. **Parallel execution** when possible (but be careful with API rate limits)

## Documentation

### Code Documentation

- Add comments for all exported functions and types
- Use godoc format for documentation comments
- Include examples in doc comments when helpful

```go
// CreateTeammate creates a new teammate in SendGrid with the specified permissions.
//
// Example:
//   teammate, err := client.CreateTeammate(&TeammateRequest{
//       Email: "user@example.com",
//       Scopes: []string{"mail.send"},
//   })
func CreateTeammate(req *TeammateRequest) (*Teammate, error) {
    // Implementation
}
```

### Resource Documentation

When adding or modifying resources:

1. Update the resource documentation in `docs/resources/`
2. Add examples to `examples/resources/`
3. Include all attributes and their descriptions
4. Document any import syntax
5. Note any special behaviors or limitations

### Generating Documentation

```bash
# Generate provider documentation
make docs
```

## Additional Resources

- [Terraform Plugin Development](https://www.terraform.io/docs/extend/index.html)
- [Go at Google: Language Design in the Service of Software Engineering](https://talks.golang.org/2012/splash.article)
- [SendGrid API Documentation](https://www.twilio.com/docs/sendgrid/api-reference)
- [Effective Go](https://golang.org/doc/effective_go)

## Getting Help

- **Questions?** Open a [Discussion](https://github.com/arslanbekov/terraform-provider-sendgrid/discussions)
- **Found a bug?** Open an [Issue](https://github.com/arslanbekov/terraform-provider-sendgrid/issues)
- **Need clarification?** Comment on the relevant issue or PR

## Recognition

Contributors are recognized in:

- Release notes
- The project's GitHub contributors page
- CHANGELOG.md (for significant contributions)

Thank you for contributing!
