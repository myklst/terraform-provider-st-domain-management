package domain_management

import (
	"context"
	"encoding/json"
	"fmt"
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
	DomainLabels      *internal.Filters      `tfsdk:"domain_labels" json:"domain_labels"`
	DomainAnnotations *internal.Filters      `tfsdk:"domain_annotations" json:"domain_annotations"`
	SubdomainLabels   *internal.Filters      `tfsdk:"subdomain_labels" json:"subdomains_labels"`
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
				Required:       false,
				Optional:       true,
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
				Required:       false,
				Optional:       true,
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

	// Create a temporary domain filter data source
	// so that we can re-use the request.Payload() method.
	domainRequest := internal.DomainFilterDataSourceModel{
		DomainLabels:      state.DomainLabels,
		DomainAnnotations: state.DomainAnnotations,
	}
	domainsFullBytes, err := d.client.GetDomainsFull(domainRequest.Payload())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domains, got error: %s", err))
		return
	}

	domainsFull := []api.DomainFull{}
	err = json.Unmarshal(domainsFullBytes, &domainsFull)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	// TODO Determine if need to give warning if no subdomains found

	// Early return if no domains are found.
	if len(domainsFull) == 0 {
		resp.Diagnostics.AddWarning("No domains found.", "Please try again with the correct domain filters.")
		state.Domains = types.DynamicNull()
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		return
	}

	bytes, err := json.Marshal(domainsFull)
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

			subdomainFiltered.Fqdn = strings.Join([]string{subdomainFiltered.Name, domain.Domain}, ".")
			subdomains = append(subdomains, *subdomainFiltered)
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
