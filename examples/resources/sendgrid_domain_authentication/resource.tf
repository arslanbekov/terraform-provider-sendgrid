# Basic domain authentication setup
resource "sendgrid_domain_authentication" "main" {
  domain             = "mycompany.com"
  subdomain          = "em"
  is_default         = true
  automatic_security = true
  custom_spf         = false
}

# Domain with custom SPF record
resource "sendgrid_domain_authentication" "marketing" {
  domain             = "marketing.mycompany.com"
  subdomain          = "mail"
  is_default         = false
  automatic_security = false
  custom_spf         = true
}

# DMARC is not part of domain authentication: SendGrid neither creates nor
# returns a _dmarc record, and the policy (p=none|quarantine|reject, rua=...)
# is the domain owner's decision. Publish it at your DNS provider next to the
# records SendGrid returns in `dns`. Route53 is shown; any DNS provider works.
resource "aws_route53_record" "dmarc" {
  zone_id = var.route53_zone_id
  name    = "_dmarc.${sendgrid_domain_authentication.main.domain}"
  type    = "TXT"
  ttl     = 300
  records = ["v=DMARC1; p=none; rua=mailto:dmarc-reports@mycompany.com"]
}
