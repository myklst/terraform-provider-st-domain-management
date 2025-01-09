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

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

type subdomainFilterDataSourceModel struct {
	DomainLabels      internal.Filters       `tfsdk:"domain_labels" json:"domain_labels"`
	DomainAnnotations *internal.Filters      `tfsdk:"domain_annotations" json:"domain_annotations"`
	SubdomainLabels   internal.Filters       `tfsdk:"subdomain_labels" json:"subdomains_labels"`
	Domains           basetypes.DynamicValue `tfsdk:"domains" json:"domains"`
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
				Description:    "Labels filter. Only domains that contain these labels will be returned as data source output.",
				AttributeTypes: internal.FilterAttributes,
				Required:       true,
			},
			"domain_annotations": schema.ObjectAttribute{
				Description:    "Annotations filter. Only domains that contain these annotations will be returned as data source output.",
				AttributeTypes: internal.FilterAttributes,
				Required:       false,
				Optional:       true,
			},
			"subdomain_labels": schema.ObjectAttribute{
				Description:    "Subdomain labels filter. Only subdomains that contain these labels will be returned as data source output",
				AttributeTypes: internal.FilterAttributes,
				Required:       true,
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

	var domainRequest = internal.DomainFilterDataSourceModel{
		DomainLabels:      state.DomainLabels,
		DomainAnnotations: state.DomainAnnotations,
	}
	domainsFullBytes, err := d.client.GetDomainsFull(domainRequest.Payload())
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

func domainFullFilter(httpResp []api.DomainFull, subdomainLabels internal.Filters) (domainsFull []api.DomainFull, diags diag.Diagnostics) {
	domainsFull = make([]api.DomainFull, 0)
	for _, domain := range httpResp {
		subdomains := []api.Subdomain{}
		for _, subdomain := range domain.Subdomains {
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

			subdomain.Fqdn = strings.Join([]string{subdomain.Name, domain.Domain}, ".")
			subdomains = append(subdomains, *subdomain)
		}

		if len(subdomains) == 0 {
			continue
		}

		domainFull := api.DomainFull{
			Domain: domain.Domain,
			Metadata: api.Metadata{
				Labels:      domain.Metadata.Labels,
				Annotations: domain.Metadata.Annotations,
			},
			Subdomains: subdomains,
		}
		domainsFull = append(domainsFull, domainFull)
	}
	return domainsFull, nil
}

func subdomainsFilter(subdomainResp api.Subdomain, labelsFilter internal.Filters) (*api.Subdomain, error) {
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

	subdomain := &api.Subdomain{
		Name: subdomainResp.Name,
		Metadata: api.Metadata{
			Labels: subdomainResp.Metadata.Labels,
		},
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
