# SendGrid Terraform Provider Examples

This directory contains practical examples for all resources provided by the SendGrid Terraform provider.

## Available Resource Examples

### Core Resources

- [sendgrid_api_key](resources/sendgrid_api_key/) - API key management with different permission levels
- [sendgrid_teammate](resources/sendgrid_teammate/) - Team member management including SSO users
- [sendgrid_subuser](resources/sendgrid_subuser/) - Subuser account creation and management

### Email Authentication & Branding

- [sendgrid_domain_authentication](resources/sendgrid_domain_authentication/) - Domain authentication setup
- [sendgrid_domain_authentication_subuser](resources/sendgrid_domain_authentication_subuser/) - Assign a parent's authenticated domain to a subuser
- [sendgrid_link_branding](resources/sendgrid_link_branding/) - Link branding for click tracking

### Templates & Content

- [sendgrid_template](resources/sendgrid_template/) - Email template creation
- [sendgrid_template_version](resources/sendgrid_template_version/) - Template version management
- [sendgrid_unsubscribe_group](resources/sendgrid_unsubscribe_group/) - Unsubscribe group configuration

### Webhooks & Integrations

- [sendgrid_event_webhook](resources/sendgrid_event_webhook/) - Event webhook configuration
- [sendgrid_parse_webhook](resources/sendgrid_parse_webhook/) - Inbound email parsing

### Single Sign-On (SSO)

- [sendgrid_sso_integration](resources/sendgrid_sso_integration/) - SSO provider integration
- [sendgrid_sso_certificate](resources/sendgrid_sso_certificate/) - SSO certificate management

## Directory Structure

Each resource example directory contains:

- `resource.tf` - Basic and advanced usage examples
- `import.sh` - Import instructions for existing resources
- Additional scenario files (where applicable)

## Getting Started

1. **Set up your provider configuration:**

   ```hcl
   terraform {
     required_providers {
       sendgrid = {
         source  = "arslanbekov/sendgrid"
         version = "~> 2.0"
       }
     }
   }

   provider "sendgrid" {
     api_key = var.sendgrid_api_key
   }
   ```

2. **Copy examples to your project:**

   ```bash
   cp -r examples/resources/sendgrid_api_key/ ./
   ```

3. **Customize the configuration:**

   - Update email addresses, domains, and URLs
   - Adjust permissions and settings for your use case
   - Set appropriate variable values

4. **Apply the configuration:**

   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

## Common Patterns

### Using Variables

```hcl
variable "company_domain" {
  description = "Company domain for SendGrid configuration"
  type        = string
  default     = "mycompany.com"
}

variable "sendgrid_api_key" {
  description = "SendGrid API key"
  type        = string
  sensitive   = true
}
```

### Resource Dependencies

```hcl
# Create domain authentication first
resource "sendgrid_domain_authentication" "main" {
  domain = var.company_domain
  # ... other configuration
}

# Then create link branding that references it
resource "sendgrid_link_branding" "main" {
  domain = "links.${var.company_domain}"
  # ... other configuration

  depends_on = [sendgrid_domain_authentication.main]
}
```

### Using Timeouts

For resources that may take time to process (like teammates with many scopes):

```hcl
resource "sendgrid_teammate" "admin" {
  # ... configuration

  timeouts {
    create = "10m"
    update = "10m"
    delete = "5m"
  }
}
```

## Best Practices

1. **Use descriptive resource names** that indicate their purpose
2. **Add comments** explaining complex configurations
3. **Use variables** for reusable values like domains and emails
4. **Set timeouts** for operations that might be slow
5. **Validate configurations** with `terraform plan` before applying
6. **Use import scripts** to bring existing resources under management

## Troubleshooting

- Check [troubleshooting guide](../docs/troubleshooting.md) for common issues
- Use `terraform refresh` to sync state with SendGrid
- Enable debug logging: `TF_LOG=DEBUG terraform apply`

## Contributing

When adding new examples:

1. Include both basic and advanced use cases
2. Add clear comments explaining configuration options
3. Provide import instructions
4. Test examples before submitting
