package domain_management

import (
	"context"
	"os"

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
				MarkdownDescription: "The Domain Management server endpoint",
				Optional:            true,
			},
		},
	}
}

func (p *DomainManagementProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config DomainManagementProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value
	if config.Endpoint.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Provider endpoint cannot be unknown",
			"",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	var (
		endpoint string
	)

	if !config.Endpoint.IsNull() {
		endpoint = config.Endpoint.ValueString()
	} else {
		endpoint = os.Getenv("DOMAIN_MANAGEMENT_ENDPOINT")
	}

	if endpoint == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("endpoint"),
			"Missing Domain Management endpoint",
			"The provider cannot create the Domain Management API client as there is a "+
				"missing or empty value for the Domain Management endpoint. Set the "+
				"endpoint value in the configuration or use the DOMAIN_MANAGEMENT_ENDPOINT "+
				"environment variable. If either is already set, ensure the value "+
				"is not empty.",
		)
		return
	}

	cfg := Config{
		Endpoint: endpoint,
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
