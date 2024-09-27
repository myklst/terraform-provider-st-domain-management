data "st-domain-management_domain_filter" "example" {
  domain_labels = jsonencode({
    "common/brand"   = "brand-A"
    "common/status"  = "new"
    "common/project" = "project-B"
  })
}
