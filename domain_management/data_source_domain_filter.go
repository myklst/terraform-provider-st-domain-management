package domain_management

import (
	"context"
	"fmt"
	"strings"

	"github.com/myklst/terraform-provider-st-domain-management/api"
	"github.com/myklst/terraform-provider-st-domain-management/domain_management/internal"
	"github.com/myklst/terraform-provider-st-domain-management/utils"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func NewDomainDataSource() datasource.DataSource {
	return &domainFilterDataSource{}
}

type domainFilterDataSource struct {
	client *api.Client
}

func (d *domainFilterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_filter"
}

func (d *domainFilterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Query domains that satisfy the filter using Terraform Data Source.",
		Attributes: map[string]schema.Attribute{
			"domains": schema.DynamicAttribute{
				Description: strings.Join([]string{
					"List of domains that match the given filter.",
					"Each domain has a metadata object that can be accessed via is dot notation.",
					"e.g. `domains[0].metadata.labels[\"common/env\"]`",
				}, "\n"),
				Computed: true,
			},
			"domain_labels": schema.ObjectAttribute{
				Description: "Labels filter. Only domains that contain these labels will be returned as data source output.",
				AttributeTypes: internal.FilterAttributes,
				Required:       true,
			},
			"domain_annotations": schema.ObjectAttribute{
				Description:    "Annotations filter. Only domains that contain these annotations will be returned as data source output.",
				AttributeTypes: internal.FilterAttributes,
				Required:       false,
				Optional:       true,
			},
		},
	}
}

func (d *domainFilterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *domainFilterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state internal.DomainFilterDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domains, err := d.client.GetDomains(state.Payload())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domains: %s", err))
		return
	}

	if len(domains) == 0 {
		resp.Diagnostics.AddWarning("No domains found. Please try again with the correct domain filters.", "")
		state.Domains = types.DynamicNull()
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		return
	}

	state.Domains, err = utils.JSONToTerraformDynamicValue(domains)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
