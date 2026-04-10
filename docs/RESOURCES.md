# Resources and Data Sources

This document provides a complete reference for all resources and data sources available in the SendGrid Terraform provider.

## Resources

### sendgrid_teammate

Manages SendGrid teammates with comprehensive support for both regular and SSO users, including pending invitation handling.

**Key Features:**

- Support for both regular and SSO teammates
- Automatic handling of pending invitations
- Comprehensive scope management
- Built-in retry logic for rate limiting
- **Important**: Name fields (`first_name`, `last_name`) are editable only for SSO users; for non-SSO users they are read-only and populated from SendGrid profile

**Example:**

```hcl
resource "sendgrid_teammate" "developer" {
  email      = "developer@example.com"
  first_name = "John"        # Only for SSO users
  last_name  = "Doe"         # Only for SSO users
  is_admin   = false
  is_sso     = true
  scopes     = ["mail.send", "templates.read"]
}
```

**Arguments:**

- `email` (Required) - Email address of the teammate
- `first_name` (Optional) - First name (required for SSO users, read-only for non-SSO users)
- `last_name` (Optional) - Last name (required for SSO users, read-only for non-SSO users)
- `is_admin` (Required) - Whether the teammate has admin privileges
- `is_sso` (Required) - Whether this is an SSO user
- `scopes` (Optional) - List of permission scopes (ignored if is_admin is true)
- `username` (Optional) - Username for the teammate (read-only for pending users)

**Attributes:**

- `user_status` - Status of the user ("active" or "pending")

### sendgrid_template

Manages SendGrid transactional email templates.

**Example:**

```hcl
resource "sendgrid_template" "welcome" {
  name       = "Welcome Email"
  generation = "dynamic"
}
```

### sendgrid_template_version

Manages versions of SendGrid templates.

**Example:**

```hcl
resource "sendgrid_template_version" "welcome_v1" {
  template_id = sendgrid_template.welcome.id
  name        = "Welcome v1"
  subject     = "Welcome to our service!"
  html_content = "<h1>Welcome!</h1>"
  active      = 1
}
```

### sendgrid_api_key

Manages SendGrid API keys with specific scopes.

**Example:**

```hcl
resource "sendgrid_api_key" "mail_send" {
  name   = "Mail Send Key"
  scopes = ["mail.send"]
}
```

### sendgrid_domain_authentication

Manages domain authentication (formerly domain whitelabel).

**Example:**

```hcl
resource "sendgrid_domain_authentication" "example" {
  domain           = "mail.example.com"
  subdomain        = "em"
  is_default       = true
  automatic_security = true
}
```

### sendgrid_link_branding

Manages link branding (formerly link whitelabel).

**Example:**

```hcl
resource "sendgrid_link_branding" "example" {
  domain    = "links.example.com"
  subdomain = "mail"
  is_default = true
}
```

### sendgrid_parse_webhook

Manages inbound parse webhooks.

**Example:**

```hcl
resource "sendgrid_parse_webhook" "example" {
  hostname    = "parse.example.com"
  url         = "https://api.example.com/parse"
  spam_check  = true
  send_raw    = false
}
```

### sendgrid_event_webhook

Manages event webhooks.

**Example:**

```hcl
resource "sendgrid_event_webhook" "example" {
  url     = "https://api.example.com/webhook"
  enabled = true

  delivered          = true
  processed          = true
  open               = true
  click              = true
  bounce             = true
  deferred           = true
  dropped            = true
  spam_report        = true
  unsubscribe        = true
  group_unsubscribe  = true
  group_resubscribe  = true
}
```

### sendgrid_unsubscribe_group

Manages unsubscribe groups.

**Example:**

```hcl
resource "sendgrid_unsubscribe_group" "marketing" {
  name        = "Marketing Emails"
  description = "Marketing and promotional emails"
  is_default  = false
}
```

## Data Sources

### sendgrid_teammate

Retrieves information about an existing teammate.

**Example:**

```hcl
data "sendgrid_teammate" "existing" {
  email = "teammate@example.com"
}
```

### sendgrid_template

Retrieves information about an existing template.

**Example:**

```hcl
data "sendgrid_template" "existing" {
  name = "Welcome Email"
}
```

### sendgrid_domain_authentication

Retrieves information about domain authentication.

**Example:**

```hcl
data "sendgrid_domain_authentication" "existing" {
  domain = "mail.example.com"
}
```

## Important Notes

### Teammate Management

1. **Pending Users**: When creating non-SSO teammates, SendGrid sends invitation emails. Users appear as "pending" until they accept the invitation.

2. **SSO vs Regular Users**:

   - SSO users require `first_name` and `last_name`
   - Regular users don't support these fields
   - The provider automatically handles these differences

3. **Scopes**: Some scopes are managed automatically by SendGrid (`2fa_exempt`, `2fa_required`, `sender_verification_legacy`) and should not be included in your configuration.

### Rate Limiting

All resources include built-in retry logic with exponential backoff to handle SendGrid's rate limits gracefully.

### Authentication

Resources inherit authentication from the provider configuration. See [Authentication Guide](AUTHENTICATION.md) for details.

## Migration Notes

When upgrading from older versions, see the [Migration Guide](../MIGRATION_GUIDE.md) for breaking changes and upgrade instructions.
