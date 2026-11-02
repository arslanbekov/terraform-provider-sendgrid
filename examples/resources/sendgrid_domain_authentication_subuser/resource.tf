# Look up a domain that is already authenticated and validated in the parent account.
data "sendgrid_domain_authentication" "app" {
  domain    = "app.example.com"
  subdomain = "em1234"
}

resource "sendgrid_subuser" "team" {
  username = "team-a"
  email    = "team-a@example.com"
  password = var.subuser_password
  ips      = ["192.0.2.10"]
}

# Assign the parent's validated domain to the subuser. The subuser sends from
# the parent's domain without new DNS records, unlike
# `sendgrid_domain_authentication` with `sub_user_on_behalf_of`, which
# authenticates a separate domain owned by the subuser.
#
# The call is made by the parent account with the provider's API key and host,
# not on behalf of the subuser, so `sub_user_on_behalf_of` does not apply.
# A subuser holds one assigned domain through this resource; changing
# `domain_id` replaces the association.
resource "sendgrid_domain_authentication_subuser" "team" {
  domain_id = data.sendgrid_domain_authentication.app.id
  username  = sendgrid_subuser.team.username
}
