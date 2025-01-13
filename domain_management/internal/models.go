package internal

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/myklst/terraform-provider-st-domain-management/api"
	"github.com/myklst/terraform-provider-st-domain-management/utils"
)

// Generic Filter type with the Include and Exclude syntax.
type Filters struct {
	Include basetypes.DynamicValue `tfsdk:"include" json:"include"`
	Exclude basetypes.DynamicValue `tfsdk:"exclude" json:"exclude"`
}

// The Terraform Types version of Filters. Used in schema implementation.
var FilterAttributes = map[string]attr.Type{
	"include": types.DynamicType,
	"exclude": types.DynamicType,
}

// During Terraform Read phase, Terraform attribute values will be unmarshaled
// into this DomainFilter struct. Usage is shared between GetDomains() and
// GetDomainsFull().
type DomainFilterDataSourceModel struct {
	DomainLabels      Filters                `tfsdk:"domain_labels" json:"domain_labels"`
	DomainAnnotations *Filters               `tfsdk:"domain_annotations" json:"domain_annotations"`
	Domains           basetypes.DynamicValue `tfsdk:"domains" json:"domains"`
}

// Returns a result that is suitable for use in api requests.
func (d *DomainFilterDataSourceModel) Payload() api.DomainReq {
	var err error

	includeLabels, err := utils.TFTypesToJSON(d.DomainLabels.Include)
	if err != nil {
		panic(err)
	}

	excludeLabels, err := utils.TFTypesToJSON(d.DomainLabels.Exclude)
	if err != nil {
		panic(err)
	}

	filter := api.Metadata{
		Labels: includeLabels,
	}

	excludeFilter := api.Metadata{
		Labels: excludeLabels,
	}

	if d.DomainAnnotations != nil {
		includeAnnotations, err := utils.TFTypesToJSON(d.DomainAnnotations.Include)
		if err != nil {
			panic(err)
		}
		filter.Annotations = includeAnnotations

		excludeAnnotations, err := utils.TFTypesToJSON(d.DomainAnnotations.Exclude)
		if err != nil {
			panic(err)
		}
		excludeFilter.Annotations = excludeAnnotations
	}

	request := api.DomainReq{
		Filter:  filter,
		Exclude: excludeFilter,
	}

	return request
}
