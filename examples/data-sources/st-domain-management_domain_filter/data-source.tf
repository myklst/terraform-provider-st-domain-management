data "st-domain-management_domain_filter" "example" {
  domain_labels = {
    include = {
      "common/brand" = "a"
    }
    exclude = {
      "common/status" = "deleted"
    }
  }
}
