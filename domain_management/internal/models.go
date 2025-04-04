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

type DomainFilterDataSourceModel struct {
	DomainLabels      *Filters               `tfsdk:"domain_labels" json:"domain_labels"`
	DomainAnnotations *Filters               `tfsdk:"domain_annotations" json:"domain_annotations"`
	Domains           basetypes.DynamicValue `tfsdk:"domains" json:"domains"`
}

func (d *DomainFilterDataSourceModel) Payload() (api.DomainReq, error) {
	var err error

	includeLabels := map[string]interface{}{}
	excludeLabels := map[string]interface{}{}
	includeAnnotations := map[string]interface{}{}
	excludeAnnotations := map[string]interface{}{}

	if d.DomainLabels != nil {
		if !d.DomainLabels.Include.IsNull() {
			includeLabels, err = utils.TFTypesToJSON(d.DomainLabels.Include)
			if err != nil {
				return api.DomainReq{}, err
			}
		}

		if !d.DomainLabels.Exclude.IsNull() {
			excludeLabels, err = utils.TFTypesToJSON(d.DomainLabels.Exclude)
			if err != nil {
				return api.DomainReq{}, err
			}
		}
	}

	if d.DomainAnnotations != nil {
		if !d.DomainAnnotations.Include.IsNull() {
			includeAnnotations, err = utils.TFTypesToJSON(d.DomainAnnotations.Include)
			if err != nil {
				return api.DomainReq{}, err
			}
		}

		if !d.DomainAnnotations.Exclude.IsNull() {
			excludeAnnotations, err = utils.TFTypesToJSON(d.DomainAnnotations.Exclude)
			if err != nil {
				return api.DomainReq{}, err
			}
		}
	}
	request := api.DomainReq{
		FilterDomains: &api.IncludeExclude{
			Include: &api.Include{
				Metadata: &api.Metadata{
					Labels:      includeLabels,
					Annotations: includeAnnotations,
				},
			},
			Exclude: &api.Exclude{
				Metadata: &api.Metadata{
					Labels:      excludeLabels,
					Annotations: excludeAnnotations,
				},
			},
		},
	}

	return request, nil
}

type FullDomainFilterDataSourceModel struct {
	DomainLabels         *Filters               `tfsdk:"domain_labels" json:"domain_labels"`
	DomainAnnotations    *Filters               `tfsdk:"domain_annotations" json:"domain_annotations"`
	SubdomainLabels      *Filters               `tfsdk:"subdomain_labels" json:"subdomain_labels"`
	SubdomainAnnotations *Filters               `tfsdk:"subdomain_annotations" json:"subdomain_annotations"`
	Domains              basetypes.DynamicValue `tfsdk:"domains" json:"domains"`
}

// Returns a result that is suitable for use in api requests.
func (d *FullDomainFilterDataSourceModel) Payload() (api.DomainReq, error) {
	var err error

	includeLabels := map[string]interface{}{}
	excludeLabels := map[string]interface{}{}
	includeAnnotations := map[string]interface{}{}
	excludeAnnotations := map[string]interface{}{}

	if d.DomainLabels != nil {
		if !d.DomainLabels.Include.IsNull() {
			includeLabels, err = utils.TFTypesToJSON(d.DomainLabels.Include)
			if err != nil {
				return api.DomainReq{}, err
			}
		}

		if !d.DomainLabels.Exclude.IsNull() {
			excludeLabels, err = utils.TFTypesToJSON(d.DomainLabels.Exclude)
			if err != nil {
				return api.DomainReq{}, err
			}
		}
	}

	if d.DomainAnnotations != nil {
		if !d.DomainAnnotations.Include.IsNull() {
			includeAnnotations, err = utils.TFTypesToJSON(d.DomainAnnotations.Include)
			if err != nil {
				return api.DomainReq{}, err
			}
		}

		if !d.DomainAnnotations.Exclude.IsNull() {
			excludeAnnotations, err = utils.TFTypesToJSON(d.DomainAnnotations.Exclude)
			if err != nil {
				return api.DomainReq{}, err
			}
		}
	}

	subdomainIncludeLabels := map[string]interface{}{}
	subdomainExcludeLabels := map[string]interface{}{}

	if d.SubdomainLabels != nil {
		if !d.SubdomainLabels.Include.IsNull() {
			subdomainIncludeLabels, err = utils.TFTypesToJSON(d.SubdomainLabels.Include)
			if err != nil {
				return api.DomainReq{}, err
			}
		}

		if !d.SubdomainLabels.Exclude.IsNull() {
			subdomainExcludeLabels, err = utils.TFTypesToJSON(d.SubdomainLabels.Exclude)
			if err != nil {
				return api.DomainReq{}, err
			}
		}
	}

	request := api.DomainReq{
		FilterDomains: &api.IncludeExclude{
			Include: &api.Include{
				Metadata: &api.Metadata{
					Labels:      includeLabels,
					Annotations: includeAnnotations,
				},
			},
			Exclude: &api.Exclude{
				Metadata: &api.Metadata{
					Labels:      excludeLabels,
					Annotations: excludeAnnotations,
				},
			},
		},
		FilterSubdomains: &api.IncludeExclude{
			Include: &api.Include{Metadata: &api.Metadata{
				Labels: subdomainIncludeLabels,
			}},
			Exclude: &api.Exclude{Metadata: &api.Metadata{
				Labels: subdomainExcludeLabels,
			}},
		},
	}

	return request, nil
}
