---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "st-domain-management_domain_filter Data Source - st-domain-management"
subcategory: ""
description: |-
  Query domains that satisfy the filter using Terraform Data Source.
---

# st-domain-management_domain_filter (Data Source)

Query domains that satisfy the filter using Terraform Data Source.

## Example Usage

```terraform
data "st-domain-management_domain_filter" "example" {
  domain_labels = jsonencode({
    "common/brand"   = "brand-A"
    "common/status"  = "new"
    "common/project" = "project-B"
  })
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `domain_labels` (String) Labels filter. Only domains that contain these labels will be returned as data source output.

### Optional

- `domain_annotations` (String) Annotations filter. Only domains that contain these annotations will be returned as data source output.

### Read-Only

- `domains` (Dynamic) List of domains that match the given filter.
Each domain has a metadata object that can be accessed via is dot notation.
e.g. `domains[0].metadata.labels["common/env"]`
