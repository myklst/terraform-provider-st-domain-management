data "st-domain-management_subdomain_filter" "example" {
  domain_labels = {
    include = {
      "common/brand" = "sige" # Data source will return only domains that belong to sige AND
      "common/env"   = "test" # domains that are used in testing env only.
    }
    exclude = {
      "common/status" = "deleted" # Data source filter will ignore any domains whose status label is deleted 
    }
  }

  subdomain_labels = {
    include = {
      "uncommon/testing" = true # Only return subdomains that contains this label and this value
    }
    exclude = { # Don't exclude any labels
    }
  }
}
