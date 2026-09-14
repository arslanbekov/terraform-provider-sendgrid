# Basic subuser for a separate application
resource "sendgrid_subuser" "app_subuser" {
  username = "app-emails"
  email    = "app-emails@mycompany.com"
  password = "SecurePassword123!"
  ips      = ["192.168.1.100"]
}

# Disabled subuser (can be enabled later)
resource "sendgrid_subuser" "staging_subuser" {
  username = "staging-app"
  email    = "staging@mycompany.com"
  password = "StagingPass456!"
  ips      = ["192.168.1.101"]
  disabled = true
}

# EU data residency subuser (requires a dedicated EU IP; region cannot be changed after creation)
resource "sendgrid_subuser" "eu_subuser" {
  username = "app-emails-eu"
  email    = "app-emails-eu@mycompany.com"
  password = "SecurePassword789!"
  region   = "eu"
  ips      = ["192.168.1.102"]
}
