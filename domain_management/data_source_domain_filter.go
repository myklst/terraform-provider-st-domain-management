package domain_management

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/myklst/terraform-provider-st-domain-management/api"
	"github.com/myklst/terraform-provider-st-domain-management/utils"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

type domainFilterDataSourceModel struct {
	DomainLabels      jsontypes.Normalized `tfsdk:"domain_labels" json:"domain_labels"`
	DomainAnnotations jsontypes.Normalized `tfsdk:"domain_annotations" json:"domain_annotations"`
	Domains           types.Set            `tfsdk:"domains" json:"domains"`
}

func (d *domainFilterDataSourceModel) Payload() (payload map[string]any) {
	domains := map[string]any{}
	labels := map[string]any{}
	annotations := map[string]any{}

	err := d.DomainLabels.Unmarshal(&labels)
	if err != nil {
		panic(err)
	}
	domains["labels"] = labels

	if !d.DomainAnnotations.IsNull() {
		err := d.DomainAnnotations.Unmarshal(&annotations)
		if err != nil {
			panic(err)
		}
		domains["annotations"] = annotations
	}

	payload = map[string]any{}
	payload["metadata"] = domains

	return
}

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
			"domains": schema.SetAttribute{
				Description: "Set of domain names that match the given filter.",
				ElementType: types.StringType,
				Computed:    true,
			},
			"domain_labels": schema.StringAttribute{
				Description: "Labels filter. Only domains that contain these labels will be returned as data source output.",
				CustomType:  jsontypes.NormalizedType{},
				Required:    true,
				Validators: []validator.String{
					utils.MustBeMapOfString{},
					utils.MustNotBeNull{},
				},
			},
			"domain_annotations": schema.StringAttribute{
				Description: "Annotations filter. Only domains that contain these annotations will be returned as data source output.",
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Validators: []validator.String{
					utils.MustBeMapOfString{},
					utils.MustNotBeNull{},
				},
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
	var state domainFilterDataSourceModel

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

	response, err := d.client.GetOnlyDomain(jsonData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domains, got error: %s", err))
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, err := io.ReadAll(response.Body)
		if err != nil {
			resp.Diagnostics.AddError("Read response Error", err.Error())
			return
		}

		jsonBody := map[string]interface{}{}
		err = json.Unmarshal(body, &jsonBody)
		if err != nil {
			resp.Diagnostics.AddError("JSON Unmarshal Error", err.Error())
			return
		}

		resp.Diagnostics.AddError("No domains found. Please try again with the correct domain label filters.",
			fmt.Sprintf("Got response %s: %s", strconv.Itoa(response.StatusCode), jsonBody["err"]))
		return
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		resp.Diagnostics.AddError("Can not read response body", "")
		return
	}

	var domains struct {
		Domains []string `json:"dt"`
	}
	domains.Domains = []string{}

	if err := json.Unmarshal(body, &domains); err != nil {
		resp.Diagnostics.AddError("Can not unmarshal JSON", err.Error())
		return
	}

	state.Domains, diags = types.SetValueFrom(ctx, types.StringType, domains.Domains)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Trace(ctx, "Done reading domain data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
