#!/bin/bash

# Import by authenticated domain ID and subuser username
terraform import sendgrid_domain_authentication_subuser.team 12345/team-a

# Or by subuser username alone; the domain ID is read from the API
terraform import sendgrid_domain_authentication_subuser.team team-a
