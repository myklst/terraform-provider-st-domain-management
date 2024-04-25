package domain_management

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/myklst/terraform-provider-domain-management/api"
	"github.com/myklst/terraform-provider-domain-management/structs"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

func NewDomainDataSource() datasource.DataSource {
	return &domainDataSource{}
}

type domainDataSource struct {
	client *api.Client
}

func (d *domainDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_filter"
}

func (d *domainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Query domains that satisfy the filter using Terraform Data Source",
		Attributes: map[string]schema.Attribute{
			"domains": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"domain_tags": schema.DynamicAttribute{
				Required: false,
				Optional: true,
			},
			"domain_labels": schema.DynamicAttribute{
				Required: false,
				Optional: true,
			},
			"domain_annotations": schema.DynamicAttribute{
				Required: false,
				Optional: true,
			},
		},
	}
}

func (d *domainDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *domainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state structs.DomainDataSourceModel

	diags := req.Config.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	payload := state.Payload()
	jsonData, err := json.Marshal(payload)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	response, err := d.client.GetForTerraform(*bytes.NewBuffer(jsonData))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domains, got error: %s", err))
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)

	if err != nil {
		resp.Diagnostics.AddError("Can not read response body", "")
		return
	}

	var res structs.DomainsDefaultResponse
	if err := json.Unmarshal(body, &res); err != nil {
		fmt.Println("Can not unmarshal JSON")
		return
	}

	domainsStringArray := []string{}
	for _, domain := range res.Data.Domains {
		domainsStringArray = append(domainsStringArray, domain.Domain)
	}
	state.Domains, diags = types.ListValueFrom(ctx, types.StringType, domainsStringArray)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Done reading domain data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
