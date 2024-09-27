data "st-domain-management_subdomain_filter" "example" {
  domain_labels = jsonencode({
    "common/brand"   = "brand-A"
    "common/status"  = "new"
    "common/project" = "project-B"
  })
  subdomain_labels = jsonencode({
    "module-specific-label/labelA" = true
    "module-specific-label/labelB" = ["a", "b", "c"]
  })
}
