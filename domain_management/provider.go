package domain_management

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &DomainManagementProvider{}
var _ provider.ProviderWithFunctions = &DomainManagementProvider{}

type DomainManagementProvider struct {
	version string
}

type DomainManagementProviderModel struct {
	Endpoint types.String `tfsdk:"endpoint"`
}

func New() provider.Provider {
	return &DomainManagementProvider{}
}

func (p *DomainManagementProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "domain-management"
	resp.Version = p.version
}

func (p *DomainManagementProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"endpoint": schema.StringAttribute{
				MarkdownDescription: "The MongoDB endpoint",
				Required:            true,
			},
		},
	}
}

func (p *DomainManagementProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data DomainManagementProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Unknown endpoint",
			"",
		)
	}

	cfg := Config{
		Endpoint: data.Endpoint.ValueString(),
	}

	client, err := cfg.Client()
	if err != nil {
		resp.Diagnostics.AddError("Create Domain Management API client Error", err.Error())
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *DomainManagementProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainAnnotationResource,
	}
}

func (p *DomainManagementProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDomainDataSource,
	}
}

func (p *DomainManagementProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}
