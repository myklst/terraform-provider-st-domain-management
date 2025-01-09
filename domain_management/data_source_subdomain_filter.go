package domain_management

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/myklst/terraform-provider-st-domain-management/api"
	"github.com/myklst/terraform-provider-st-domain-management/domain_management/internal"
	"github.com/myklst/terraform-provider-st-domain-management/utils"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type subdomain struct {
	Labels map[string]interface{} `tfsdk:"labels" json:"labels"`
	Name   string                 `tfsdk:"name" json:"name"`
	Fqdn   string                 `tfsdk:"fqdn" json:"fqdn"`
}

type domainDetails struct {
	Labels      map[string]interface{} `tfsdk:"labels" json:"labels"`
	Annotations map[string]interface{} `tfsdk:"annotations" json:"annotations"`
	Name        string                 `tfsdk:"name" json:"domain"`
}

type domainFull struct {
	Domain     domainDetails `tfsdk:"domain" json:"domain"`
	Subdomains []subdomain   `tfsdk:"subdomains" json:"subdomains"`
}

type subdomainFilterDataSourceModel struct {
	DomainLabels      internal.Filters       `tfsdk:"domain_labels" json:"domain_labels"`
	DomainAnnotations *internal.Filters      `tfsdk:"domain_annotations" json:"domain_annotations"`
	SubdomainLabels   internal.Filters       `tfsdk:"subdomain_labels" json:"subdomains_labels"`
	Domains           basetypes.DynamicValue `tfsdk:"domains" json:"domains"`
}

func (d *subdomainFilterDataSourceModel) Payload() api.DomainReq {
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

func NewSubdomainDataSource() datasource.DataSource {
	return &subdomainFilterDataSource{}
}

type subdomainFilterDataSource struct {
	client *api.Client
}

func (d *subdomainFilterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_subdomain_filter"
}

func (d *subdomainFilterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Query subdomains that satisfy the filter using Terraform Data Source.",
		Attributes: map[string]schema.Attribute{
			"domains": schema.DynamicAttribute{
				Description: strings.Join([]string{
					"List of domains that match the given filter.",
					"Each domain has a subdomains list.",
					"Each subdomain in the list has name fqdn, and a labels object that can be accessed via its dot notation.",
					"e.g. `domains[0].subdomains[0].labels[\"common/env\"]`",
					"Additionally, each domain has a metadata object that can be accessed via its dot notation.",
					"e.g. `domains[0].metadata.labels[\"common/env\"]`",
				}, "\n"),
				Computed: true,
			},
			"domain_labels": schema.ObjectAttribute{
				Description: "Labels filter. Only domains that contain these labels will be returned as data source output.",
				AttributeTypes: map[string]attr.Type{
					"include": types.DynamicType,
					"exclude": types.DynamicType,
				},
				Required: true,
			},
			"domain_annotations": schema.ObjectAttribute{
				Description: "Annotations filter. Only domains that contain these annotations will be returned as data source output.",
				AttributeTypes: map[string]attr.Type{
					"include": types.DynamicType,
					"exclude": types.DynamicType,
				},
				Required: false,
				Optional: true,
			},
			"subdomain_labels": schema.ObjectAttribute{
				Description: "Subdomain labels filter. Only subdomains that contain these labels will be returned as data source output",
				AttributeTypes: map[string]attr.Type{
					"include": types.DynamicType,
					"exclude": types.DynamicType,
				},
				Required: true,
			},
		},
	}
}

func (d *subdomainFilterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*api.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *api.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *subdomainFilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state subdomainFilterDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domainsFullBytes, err := d.client.GetDomainsFull(state.Payload())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domains, got error: %s", err))
		return
	}

	if len(domainsFullBytes) == 0 {
		resp.Diagnostics.AddWarning("No domains found. Please try again with the correct domain filters.", "")
		state.Domains = types.DynamicNull()
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		return
	}

	domainsFull := []api.DomainFull{}
	err = json.Unmarshal(domainsFullBytes, &domainsFull)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	domains, diags := domainFullFilter(domainsFull, state.SubdomainLabels)
	if diags.HasError() {
		return
	}

	if len(domains) == 0 {
		resp.Diagnostics.AddWarning("No subdomains found. Please try again with the correct subdomain filters.", "")
		state.Domains = types.DynamicNull()
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		return
	}

	bytes, err := json.Marshal(domains)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	state.Domains, err = utils.JSONToTerraformDynamicValue(bytes)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// This cannot be done at the api level as there is no api support for filtering subdomains
func domainFullFilter(httpResp []api.DomainFull, subdomainLabels internal.Filters) (domainsFull []domainFull, diags diag.Diagnostics) {
	domainsFull = make([]domainFull, 0)
	for _, domainResp := range httpResp {
		subdomains := []subdomain{}
		for _, subdomain := range domainResp.Subdomains {
			if len(subdomain.Metadata.Labels) == 0 {
				continue
			}

			subdomain, err := subdomainsFilter(subdomain, subdomainLabels)
			if err != nil {
				diags = append(diags, diag.NewErrorDiagnostic("Cannot convert subdomain api model to Terraform", err.Error()))
				return
			}

			if subdomain == nil {
				continue
			}

			subdomain.Fqdn = strings.Join([]string{subdomain.Name, domainResp.Domain}, ".")
			subdomains = append(subdomains, *subdomain)
		}

		if len(subdomains) == 0 {
			continue
		}

		domainDetail := domainDetails{
			Name:        domainResp.Domain,
			Labels:      domainResp.Metadata.Labels,
			Annotations: domainResp.Metadata.Annotations,
		}

		domain := domainFull{
			Domain:     domainDetail,
			Subdomains: subdomains,
		}
		domainsFull = append(domainsFull, domain)
	}
	return domainsFull, nil
}

func subdomainsFilter(subdomainResp api.Subdomain, labelsFilter internal.Filters) (*subdomain, error) {
	// To determine whether the subdomain labels satisfies the filter in the data source input,
	// a three step process is performed.
	// 1. Unmarshal the filter input into a map[string]interface
	// 2. For each map key, use the map key to access the labels map[string] from the api response
	// 3. Ensure that the map[string] from data source and the map[string] from api response is the same
	if len(subdomainResp.Metadata.Labels) == 0 {
		return nil, nil
	}

	filter, err := utils.TFTypesToJSON(labelsFilter.Include)
	if err != nil {
		return nil, err
	}

	apiResponse := map[string]interface{}{}
	for k := range filter {
		apiResponse[k] = subdomainResp.Metadata.Labels[k]
	}

	// Return nil if subdomain labels filter's map contents (data source user input)
	// is not found in the subdomain from the api response
	if !reflect.DeepEqual(filter, apiResponse) {
		return nil, nil
	}

	subdomain := &subdomain{
		Name:   subdomainResp.Name,
		Labels: subdomainResp.Metadata.Labels,
	}

	if labelsFilter.Exclude.IsNull() {
		return subdomain, nil
	}

	exclude, err := utils.TFTypesToJSON(labelsFilter.Exclude)
	if err != nil {
		return nil, err
	}

	apiResponse = map[string]interface{}{}
	for k := range exclude {
		apiResponse[k] = subdomainResp.Metadata.Labels[k]
	}

	// Return nil if subdomain labels exclude's map contents (data source user input)
	// is found in the subdomain from the api response
	if reflect.DeepEqual(exclude, apiResponse) {
		return nil, nil
	}

	return subdomain, nil
}
