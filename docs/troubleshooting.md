# SendGrid Terraform Provider - Troubleshooting Guide

This guide helps you resolve common issues when using the SendGrid Terraform provider.

## Common Errors and Solutions

### 1. Invalid or Unassignable Scopes Error

**Error Message:**

```
Error: Invalid or unassignable scopes provided. This can happen when:
1. Using invalid scope names (check SendGrid API documentation)
2. Your SendGrid plan doesn't support certain scopes
3. Including automatically managed scopes like '2fa_exempt' or '2fa_required'
```

**Solution:**

- Remove any invalid scopes from your configuration
- Check that your SendGrid plan supports the scopes you're trying to use
- Never include `2fa_exempt` or `2fa_required` - these are managed automatically

### 2. Operation Cancellation Recovery

**Error Message:**

```
Error: operation was canceled or timed out. This can happen when:
1. You pressed Ctrl+C during execution
2. The operation took longer than the configured timeout
3. Network connectivity issues occurred
```

**Solution:**

1. Check your SendGrid dashboard for partially created resources
2. Run `terraform refresh` to update state
3. Re-run the operation: `terraform apply`

The provider passes Terraform's context to every API call, so Ctrl+C aborts an
in-flight request instead of letting it finish. Step 1 matters: an interrupted
create may or may not have reached SendGrid.

### 3. Rate Limiting Issues

**Error Message:**

```
Error: Rate limit exceeded (HTTP 429)
```

**Solution:**

1. Use reduced parallelism: `terraform apply -parallelism=1`
2. Increase timeouts in your resource configuration:

   ```hcl
   resource "sendgrid_teammate" "example" {
     # ... configuration ...

     timeouts {
       create = "30m"
       update = "30m"
       delete = "30m"
     }
   }
   ```

### 4. Network and Connection Failures

**Error Message:**

```
Error: request failed with HTTP 0: ... dial tcp: connect: connection refused
```

`HTTP 0` is not a SendGrid status code. It means the request never got a
response at all -- DNS, TLS, a proxy, or connectivity failed before SendGrid
answered. A genuine SendGrid rejection always carries its own status (401, 403,
404, 429, 5xx), so `0` never overlaps with one.

**Solution:**

1. Check connectivity to `api.sendgrid.com` from wherever Terraform runs
2. Check proxy and firewall settings, including `HTTPS_PROXY`
3. Verify the `host` provider argument or `SENDGRID_HOST` if you set either --
   a wrong host surfaces here rather than as an API error
4. Check the SendGrid status page: https://status.sendgrid.com/
5. Re-run the operation -- only HTTP 429 is retried automatically, transport
   failures are not

## Valid SendGrid Scopes Reference

### Core Mail Operations

- `mail.send` - Send emails
- `mail.batch.create`, `mail.batch.read`, `mail.batch.update`, `mail.batch.delete` - Batch email operations

### Templates

- `templates.create`, `templates.read`, `templates.update`, `templates.delete` - Template management
- `templates.versions.create`, `templates.versions.read`, `templates.versions.update`, `templates.versions.delete` - Template version management
- `templates.versions.activate.create`, `templates.versions.activate.read`, `templates.versions.activate.update`, `templates.versions.activate.delete` - Template activation

### Statistics and Analytics

- `stats.read`, `stats.global.read` - Basic statistics
- `categories.stats.read`, `categories.stats.sums.read` - Category statistics
- `browsers.stats.read`, `devices.stats.read`, `geo.stats.read` - Device/browser/geo statistics
- `clients.stats.read`, `clients.desktop.stats.read`, `clients.phone.stats.read`, `clients.tablet.stats.read`, `clients.webmail.stats.read` - Client statistics
- `mailbox_providers.stats.read` - Mailbox provider statistics

### User and Team Management

- `teammates.create`, `teammates.read`, `teammates.update`, `teammates.delete` - Teammate management
- `user.profile.create`, `user.profile.read`, `user.profile.update`, `user.profile.delete` - User profile
- `user.account.read`, `user.credits.read`, `user.email.read`, `user.username.read` - User information
- `sso.teammates.create`, `sso.teammates.update` - SSO teammate management

### API Keys

- `api_keys.create`, `api_keys.read`, `api_keys.update`, `api_keys.delete` - API key management

### Suppressions and Lists

- `asm.groups.create`, `asm.groups.read`, `asm.groups.update`, `asm.groups.delete` - Unsubscribe groups
- `asm.groups.suppressions.create`, `asm.groups.suppressions.read`, `asm.groups.suppressions.update`, `asm.groups.suppressions.delete` - Group suppressions
- `suppression.create`, `suppression.read`, `suppression.update`, `suppression.delete` - General suppressions
- `suppression.blocks.*`, `suppression.bounces.*`, `suppression.invalid_emails.*`, `suppression.spam_reports.*`, `suppression.unsubscribes.*` - Specific suppression types

### Marketing (Requires Marketing Plans)

- `marketing.read`, `marketing.automation.read` - Marketing features
- `newsletter.create`, `newsletter.read`, `newsletter.update`, `newsletter.delete` - Newsletter management

### Advanced Features (Pro+ Plans)

- `subusers.*` - Subuser management (requires Pro+ plan)
- `billing.*` - Billing operations
- `ips.*` - IP address management
- `whitelabel.*` - Domain authentication
- `credentials.*` - Credential management

### Settings and Configuration

- `mail_settings.*` - Mail settings configuration
- `tracking_settings.*` - Tracking settings
- `partner_settings.*` - Partner integrations
- `access_settings.*` - Access control settings

### Webhooks

- `user.webhooks.event.settings.*` - Event webhook settings
- `user.webhooks.parse.settings.*` - Parse webhook settings

### Validation and Testing

- `validations.email.create`, `validations.email.read` - Email validation
- `email_testing.read`, `email_testing.write` - Email testing

## SendGrid Plan Limitations

### Free Plan

- Basic scopes: `mail.send`, `stats.read`, `templates.read`
- Limited teammate management

### Essentials Plan

- All Free plan scopes
- Extended statistics scopes
- Basic teammate management

### Pro Plan

- All Essentials scopes
- `subusers.*` scopes
- Advanced IP management
- Full teammate management

### Marketing Plans

- `marketing.*` scopes
- `newsletter.*` scopes
- Enhanced analytics

## Best Practices

### 1. Always Validate First

```bash
# Validate configuration
terraform validate

# Plan before applying
terraform plan

# Apply with confirmation
terraform apply
```

### 2. Use Appropriate Timeouts

```hcl
resource "sendgrid_teammate" "example" {
  # ... configuration ...

  timeouts {
    create = "30m"  # Especially important for bulk operations
    update = "30m"
    delete = "30m"
  }
}
```

### 3. Handle Rate Limiting

```bash
# For bulk operations, reduce parallelism
terraform apply -parallelism=2

# For very large operations
terraform apply -parallelism=1
```

### 4. Scope Selection Guidelines

- **Start minimal**: Begin with basic scopes like `mail.send`
- **Add incrementally**: Add scopes as needed
- **Check plan limits**: Verify your SendGrid plan supports the scopes
- **Avoid automatic scopes**: Never include `2fa_exempt` or `2fa_required`

### 5. Error Recovery Workflow

1. **Check the error**: Read the enhanced error message
2. **Verify scopes**: Ensure all scopes are valid for your plan
3. **Refresh state**: Run `terraform refresh` if operation was interrupted
4. **Retry with fixes**: Apply the corrected configuration

## Getting Help

1. **Check this troubleshooting guide** for common issues
2. **Review SendGrid API documentation** for scope requirements
3. **Check SendGrid plan limitations** in your account
4. **Use enhanced error messages** for specific guidance
5. **Open an issue** on the provider repository for bugs

## Automatic Scopes (Never Include These)

These scopes are managed automatically by SendGrid and should never be included in your configuration:

- `2fa_exempt`
- `2fa_required`

## Debugging Tips

### Enable Terraform Debug Logging

```bash
export TF_LOG=DEBUG
terraform apply
```

### Check SendGrid Dashboard

Always verify the state in your SendGrid dashboard when troubleshooting, especially after:

- Failed operations
- Canceled operations
- Timeout errors

### Validate Scope Names

Use the complete scope list above to verify spelling and availability of scopes in your configuration.
