data "domain-management_domain_filter" "example" {
  domain_labels = {
    "common/status" = "new"
  }
  domain_tags = {
    brand   = "pg"
    country = "default"
  }
}
