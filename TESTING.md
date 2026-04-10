# Testing Guide

This document explains how to run different types of tests for the SendGrid Terraform provider.

## Test Types

### 1. Unit Tests

Unit tests verify resource behavior using mock HTTP servers — no API key or network access needed.

```bash
# Run all unit tests (SDK + resource layer)
go test ./... -timeout=60s
```

**What they test:**

- Provider configuration and schema validation
- Resource Read functions (success, 404 handling, error propagation)
- Resource Update/Delete with mock HTTP verification (PATCH vs PUT, etc.)
- SDK client functions (Create, Read, Update, Delete for all resources)
- Error handling and status code propagation

**GitHub Actions:** Always run on every PR and push

### 2. Acceptance Tests

Acceptance tests interact with the real SendGrid API and require API credentials.

```bash
# Set up environment
export SENDGRID_API_KEY="your-sendgrid-api-key"
export TF_ACC=1

# Run all acceptance tests
go test -v ./sendgrid/ -run '^TestAcc' -timeout=30m -parallel=1

# Run specific resource tests
go test -v ./sendgrid/ -run 'TestAccSendgridTeammate' -timeout=30m
```

**What they test:**

- Full CRUD operations on real SendGrid resources
- Rate limiting behavior under load
- Integration between multiple resources
- Data source functionality
- Import state round-trip

**GitHub Actions:** Only run on master branch if `SENDGRID_API_KEY` secret is configured

### 3. Test Compilation

Verify that all tests compile correctly without running them.

```bash
go test -c ./sendgrid/ -o /dev/null
```

**GitHub Actions:** Always run to ensure test quality

## Test Categories

### Resource Unit Tests (13/14 resources covered)

Every resource has mock-based unit tests covering Read success, 404 handling, and schema validation:

- `sendgrid_api_key`
- `sendgrid_domain_authentication`
- `sendgrid_domain_authentication_validation`
- `sendgrid_event_webhook` (acceptance tests only — uses list-based Read)
- `sendgrid_link_branding`
- `sendgrid_parse_webhook`
- `sendgrid_sso_certificate`
- `sendgrid_sso_integration`
- `sendgrid_subuser`
- `sendgrid_teammate`
- `sendgrid_template`
- `sendgrid_template_version`
- `sendgrid_unsubscribe_group`
- `sendgrid_webhook_security_policy`

### SDK Unit Tests

- `sdk/client_test.go` — Client creation, JSON body serialization
- `sdk/domain_authentication_test.go` — Parse, Read, Validate functions
- `sdk/parse_webhook_test.go` — Full CRUD with method verification
- `sdk/errors_test.go` — Error parsing, enhancement, retry logic

### Acceptance Tests (14 resources)

All resources have acceptance tests covering Create, Read, Update, Delete, and Import operations.

### Data Source Tests (4/4 covered)

- `sendgrid_template`
- `sendgrid_template_version`
- `sendgrid_teammate`
- `sendgrid_unsubscribe_group`

### Special Test Suites

- **Rate Limiting Tests** — High-volume scenarios
- **Integration Tests** — Multi-resource workflows

## Running Tests Locally

### Prerequisites

1. **Go 1.25+** installed
2. **SendGrid API Key** with appropriate permissions (for acceptance tests only)
3. **Test SendGrid Account** (recommended for acceptance tests)

### Basic Test Run

```bash
# 1. Clone the repository
git clone https://github.com/arslanbekov/terraform-provider-sendgrid
cd terraform-provider-sendgrid

# 2. Install dependencies
go mod download

# 3. Run unit tests (no API key needed)
go test ./... -timeout=60s

# 4. Run acceptance tests (API key required)
export SENDGRID_API_KEY="your-api-key"
export TF_ACC=1
go test -v ./sendgrid/ -run '^TestAcc' -timeout=30m -parallel=1
```

### Running Tests with Coverage

```bash
# Generate coverage report
go test ./... -timeout=60s -coverprofile=coverage.txt -covermode=atomic

# View coverage in browser
go tool cover -html=coverage.txt

# View per-function coverage
go tool cover -func=coverage.txt | grep resource_sendgrid
```

### Rate Limiting Considerations

When running acceptance tests:

- **Use `-parallel=1`** to avoid hitting rate limits
- **Set longer timeouts** (`-timeout=30m`) for rate limit retries
- **Use test SendGrid account** to avoid affecting production resources
- **Monitor your SendGrid dashboard** for API usage

### Test Environment Setup

For consistent testing, you can create a `.env` file:

```bash
# .env (don't commit this file)
export SENDGRID_API_KEY="SG.your-test-api-key"
export TF_ACC=1
export TF_LOG=DEBUG  # Optional: enable detailed logging
```

Then source it before running tests:

```bash
source .env
go test -v ./sendgrid/ -run '^TestAcc' -timeout=30m -parallel=1
```

## GitHub Actions Behavior

### On Pull Requests

- Unit tests run (SDK + resource layer)
- Test compilation verification
- Coverage report generated and uploaded to Codecov
- Acceptance tests skipped (no API access)

### On Master Branch

- Unit tests run
- Test compilation verification
- Acceptance tests run (if `SENDGRID_API_KEY` secret is configured)
- Full coverage validation

## Writing Tests

### Unit Tests (Mock HTTP Server)

All resource unit tests follow this pattern:

```go
package sendgrid

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestMyResourceRead_Success(t *testing.T) {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/endpoint/123", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte(`{"id":"123","name":"test"}`))
    })

    server := httptest.NewServer(mux)
    defer server.Close()
    config := &Config{APIKey: "test-key", Host: server.URL}

    r := resourceSendgridMyResource()
    d := r.TestResourceData()
    d.SetId("123")

    diags := resourceSendgridMyResourceRead(context.Background(), d, config)

    if diags.HasError() {
        t.Fatalf("Read() returned unexpected error: %v", diags)
    }
    if d.Get("name") != "test" {
        t.Errorf("Read() name = %v, want %q", d.Get("name"), "test")
    }
}

func TestMyResourceRead_NotFound(t *testing.T) {
    mux := http.NewServeMux()
    mux.HandleFunc("/api/endpoint/999", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusNotFound)
        _, _ = w.Write([]byte(`{"errors":[{"message":"not found"}]}`))
    })

    server := httptest.NewServer(mux)
    defer server.Close()
    config := &Config{APIKey: "test-key", Host: server.URL}

    r := resourceSendgridMyResource()
    d := r.TestResourceData()
    d.SetId("999")

    diags := resourceSendgridMyResourceRead(context.Background(), d, config)

    if diags.HasError() {
        t.Fatalf("Read() should not error on 404")
    }
    if d.Id() != "" {
        t.Errorf("Read() should clear ID on 404, got: %s", d.Id())
    }
}
```

### Acceptance Tests

Follow the standard terraform-plugin-sdk v2 test pattern:

```go
package sendgrid_test

func TestAccSendgridMyResource_basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:     func() { testAccPreCheck(t) },
        Providers:    testAccProviders,
        CheckDestroy: testAccCheckMyResourceDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccMyResourceConfig(),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckMyResourceExists("sendgrid_my_resource.test"),
                ),
            },
        },
    })
}
```

## Troubleshooting

### "Acceptance tests skipped unless env 'TF_ACC' set"

**Solution:** Set `TF_ACC=1` environment variable

### "HTTP 429 Too Many Requests"

**Solution:** Use `-parallel=1` flag and ensure proper rate limiting

### "API key permissions error"

**Solution:** Ensure your API key has all necessary scopes:

- `mail.send`
- `templates.read`, `templates.write`
- `teammates.read`, `teammates.write`
- `user.read`, `user.write`
- And others depending on resources being tested

### Tests hang or timeout

**Solution:**

- Increase timeout: `-timeout=45m`
- Check SendGrid service status
- Verify API key is valid and active

---

For more information, see the main [README.md](README.md) and [Rate Limiting Documentation](docs/rate_limiting.md).
