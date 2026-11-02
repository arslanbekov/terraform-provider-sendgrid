# Look up an authenticated domain of the parent account by domain name.
# Subuser-owned domains are excluded from the lookup.
data "sendgrid_domain_authentication" "app" {
  domain = "app.example.com"
}

# When several authenticated domains share the same domain, pick one by subdomain.
data "sendgrid_domain_authentication" "marketing" {
  domain    = "app.example.com"
  subdomain = "em5678"
}

output "app_domain_valid" {
  value = data.sendgrid_domain_authentication.app.valid
}
