# Terraform Provider for SendGrid

[![Build Status](https://github.com/arslanbekov/terraform-provider-sendgrid/workflows/Tests/badge.svg)](https://github.com/arslanbekov/terraform-provider-sendgrid/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/arslanbekov/terraform-provider-sendgrid)](https://goreportcard.com/report/github.com/arslanbekov/terraform-provider-sendgrid)
[![codecov](https://codecov.io/gh/arslanbekov/terraform-provider-sendgrid/branch/master/graph/badge.svg)](https://codecov.io/gh/arslanbekov/terraform-provider-sendgrid)
[![Go Version](https://img.shields.io/github/go-mod/go-version/arslanbekov/terraform-provider-sendgrid)](https://github.com/arslanbekov/terraform-provider-sendgrid/blob/master/go.mod)
[![License](https://img.shields.io/badge/License-MPL%202.0-blue.svg)](https://opensource.org/licenses/MPL-2.0)
[![OpenTofu Compatible](https://img.shields.io/badge/OpenTofu-Compatible-orange?logo=opentofu)](https://opentofu.org/)

A comprehensive Terraform provider for managing SendGrid resources with enhanced features and reliability.

> **Fully compatible with [OpenTofu](https://opentofu.org/)** - This provider works seamlessly with both Terraform and OpenTofu.

## Key Features

- **Advanced Rate Limiting** - Built-in exponential backoff and retry logic for reliable operations
- **Teammate Management** - Complete lifecycle management, including pending invitations and SSO support
- **Template Management** - Full template and version control with dynamic content
- **Webhook Security** - OAuth and signature verification for parse webhooks
- **Event Webhooks** - Real-time email event notifications with friendly names
- **Multiple Auth Methods** - Environment variables, Terraform variables, and CI/CD integration
- **Rich Documentation** - Extensive examples, troubleshooting guides, and API references

## Quick Start

```bash
# Set your API key
export SENDGRID_API_KEY="SG.your-api-key-here"
```

```hcl
terraform {
  required_providers {
    sendgrid = {
      source  = "arslanbekov/sendgrid"
      version = "~> 2.0"
    }
  }
}

provider "sendgrid" {}

resource "sendgrid_teammate" "example" {
  email    = "teammate@example.com"
  is_admin = false
  is_sso   = false
  scopes   = ["mail.send"]
}
```

```bash
# Using Terraform
terraform init && terraform apply

# Or using OpenTofu
tofu init && tofu apply
```

## Documentation

| Topic                                                                                                             | Description                                            |
| ----------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------ |
| [Installation](https://github.com/arslanbekov/terraform-provider-sendgrid/blob/master/docs/INSTALLATION.md)       | Installation methods and requirements                  |
| [Authentication](https://github.com/arslanbekov/terraform-provider-sendgrid/blob/master/docs/AUTHENTICATION.md)   | All authentication methods and security best practices |
| [Resources](https://github.com/arslanbekov/terraform-provider-sendgrid/blob/master/docs/RESOURCES.md)             | Complete list of resources and data sources            |
| [Troubleshooting](https://github.com/arslanbekov/terraform-provider-sendgrid/blob/master/docs/troubleshooting.md) | Common issues and solutions                            |
| [Rate Limiting](https://github.com/arslanbekov/terraform-provider-sendgrid/blob/master/docs/rate_limiting.md)     | Rate limiting handling and best practices              |
| [Migration Guide](https://github.com/arslanbekov/terraform-provider-sendgrid/blob/master/MIGRATION_GUIDE.md)      | Guide for migrating between major versions             |
| [Testing Guide](https://github.com/arslanbekov/terraform-provider-sendgrid/blob/master/TESTING.md)                | How to run and write tests                             |

## Popular Use Cases

- **Team Management**: Invite and manage teammates with specific permissions
- **Email Templates**: Create and version email templates
- **API Key Management**: Secure API key creation with minimal scopes
- **Domain Setup**: Configure domain authentication and link branding
- **Webhook Configuration**: Set up event and parse webhooks

## Supported Resources

- **Teammate Management**: `sendgrid_teammate` - Manage team members and permissions
- **Templates**: `sendgrid_template`, `sendgrid_template_version` - Email template management
- **API Keys**: `sendgrid_api_key` - Scoped API key management
- **Domain Configuration**: `sendgrid_domain_authentication`, `sendgrid_link_branding` - Domain setup
- **Webhooks**: `sendgrid_event_webhook`, `sendgrid_parse_webhook`, `sendgrid_webhook_security_policy` - Webhook configuration
- **SSO**: `sendgrid_sso_integration`, `sendgrid_sso_certificate` - Single Sign-On setup
- **Subusers**: `sendgrid_subuser` - Subuser account management
- **Unsubscribe Groups**: `sendgrid_unsubscribe_group` - Manage unsubscribe groups
- **Domain Validation**: `sendgrid_domain_authentication_validation` - Domain DNS validation

See [full documentation](docs/RESOURCES.md) for details.

## Quick Links

- [Terraform Registry](https://registry.terraform.io/providers/arslanbekov/sendgrid)
- [OpenTofu Registry](https://search.opentofu.org/provider/arslanbekov/sendgrid/latest)
- [Report Issues](https://github.com/arslanbekov/terraform-provider-sendgrid/issues)
- [Discussions](https://github.com/arslanbekov/terraform-provider-sendgrid/discussions)
- [SendGrid API Documentation](https://www.twilio.com/docs/sendgrid/api-reference)

## Development

### Requirements

- [Go](https://golang.org/doc/install) 1.25+ (see [go.mod](go.mod))
- [Terraform](https://www.terraform.io/downloads.html) 1.0+
- SendGrid API key with appropriate permissions

### Building

```bash
go build -o terraform-provider-sendgrid
```

### Testing

The provider includes unit tests (with mock HTTP servers) and acceptance tests (against real SendGrid API):

```bash
# Run all unit tests (no API key needed)
go test ./... -timeout=60s

# Run acceptance tests (requires SENDGRID_API_KEY)
export SENDGRID_API_KEY="your-key"
export TF_ACC=1
go test -v ./sendgrid/ -run '^TestAcc' -timeout=30m -parallel=1
```

All resources have unit tests with mock HTTP servers covering Read (success, 404, error), schema validation, and key CRUD operations. Acceptance tests run against the real API on the master branch.

See [TESTING.md](TESTING.md) for detailed testing instructions.

### Contributing

Contributions are welcome! We'd love your help improving this provider.

**Before contributing:**

- Read our [Contributing Guide](CONTRIBUTING.md)
- Review the [Code of Conduct](CODE_OF_CONDUCT.md)

**Quick start:**

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Add tests for your changes
4. Run `make fmt` and `make lint`
5. Commit your changes with clear messages ([Conventional Commits](https://www.conventionalcommits.org/))
6. Push to your branch and create a Pull Request

For significant changes, please open an issue first to discuss the proposed changes.

## Commercial Support & Professional Services

Need help with integrating this provider or other Terraform/OpenTofu solutions? I offer professional services including:

- **Provider Integration** - Custom implementation and configuration
- **Migration Assistance** - Smooth migration from manual setup to Infrastructure as Code
- **Terraform/OpenTofu Consulting** - Architecture design and best practices
- **Custom Provider Development** - Terraform providers for your specific needs
- **Training & Workshops** - Team training on Terraform/OpenTofu

For inquiries and collaboration opportunities, please contact: **[sendgrid@arslanbekov.com](mailto:sendgrid@arslanbekov.com)**

## License

This project is licensed under the [Mozilla Public License 2.0](LICENSE).

---

**Disclaimer:** This is an unofficial provider maintained by the community. While it offers enhanced features and comprehensive testing, evaluate thoroughly for production use.
