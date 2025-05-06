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
					"Each element contains the following attributes:",
					"  - `domain` - The name of this domain",
					"  - `metadata` - All the metadata of this domain",
					"    - `labels` - JSON key value pair",
					"    - `annotations` - JSON key value pair",
					"  - `subdomains` - List of all the subdomains of this domain. ",
					"If `subdomain_labels` is not null, only filtered subdomains are present. ",
					"Each element contains the following attributes:",
					"    - `name` - The name of this subdomain",
					"    - `fqdn` - The fully qualified domain name.",
					"    - `metadata` - All the metadata of this domain",
					"      - `labels` - JSON key value pair",
					"",
					"Labels or annotations can be accessed via dot notation",
					"e.g. `domains[0].metadata.labels[\"common/env\"]`.",
					"",
					"Labels for each subdomain can also be accessed via dot notation",
					"e.g. `domains[0].subdomains[0].metadata.labels[\"feature_a/enable\"]`.",
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

	domainRequest := internal.FullDomainFilterDataSourceModel{
		DomainLabels:      state.DomainLabels,
		DomainAnnotations: state.DomainAnnotations,
		SubdomainLabels:   state.SubdomainLabels,
	}

	payload, err := domainRequest.Payload()
	if err != nil {
		resp.Diagnostics.AddError("JSON Error", fmt.Sprintf("Cannot convert filter input to json: %s", err))
		return
	}

	domainsFull, err := d.client.GetDomainsFull(payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domains, got error: %s", err))
		return
	}

	// Early return if no domains are found.
	if len(domainsFull) == 0 {
		resp.Diagnostics.AddWarning("No domains found.", "Double check your data source input.")

		// Set the state to an empty list if no domains are found
		emptyList := json.RawMessage([]byte("[]"))
		state.Domains, err = utils.JSONToTerraformDynamicValue(emptyList)
		if err != nil {
			resp.Diagnostics.AddError(err.Error(), "")
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		return
	}

	domainsFull, diags = processDomainFull(domainsFull)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
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

func processDomainFull(httpResp []*api.DomainFull) (domainsFull []*api.DomainFull, diags diag.Diagnostics) {
	domainsFull = make([]*api.DomainFull, 0)

	for _, domain := range httpResp {
		subdomains := []api.Subdomain{}
		for _, subdomain := range domain.Subdomains {
			if len(subdomain.Metadata.Labels) == 0 {
				continue
			}

			subdomain.Fqdn = strings.Join([]string{subdomain.Name, domain.Domain}, ".")
			subdomains = append(subdomains, subdomain)
		}

		if len(subdomains) == 0 {
			diags.AddWarning(
				fmt.Sprintf("%s has no subdomains after filtering", domain.Domain),
				"Please try again with the correct filter",
			)
			continue
		}

		domainFull := &api.DomainFull{
			Domain: domain.Domain,
			Metadata: api.Metadata{
				Labels:      domain.Metadata.Labels,
				Annotations: domain.Metadata.Annotations,
			},
			Subdomains: subdomains,
		}
		domainsFull = append(domainsFull, domainFull)
	}
	return domainsFull, diags
}
