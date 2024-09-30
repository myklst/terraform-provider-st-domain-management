package domain_management

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/myklst/terraform-provider-st-domain-management/api"
	"github.com/myklst/terraform-provider-st-domain-management/utils"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type domain struct {
	Domain      types.String         `tfsdk:"domain" json:"domain"`
	Labels      jsontypes.Normalized `tfsdk:"labels" json:"labels"`
	Annotations jsontypes.Normalized `tfsdk:"annotations" json:"annotations"`
}

type domainFilterDataSourceModel struct {
	DomainLabels      jsontypes.Normalized `tfsdk:"domain_labels" json:"domain_labels"`
	DomainAnnotations jsontypes.Normalized `tfsdk:"domain_annotations" json:"domain_annotations"`
	Domains           []domain             `tfsdk:"domains" json:"domains"`
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
			"domains": schema.SetNestedAttribute{
				Description: "Set of domain names that match the given filter.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Description: "The name of the domain.",
							Computed:    true,
						},
						"labels": schema.StringAttribute{
							CustomType: jsontypes.NormalizedType{},
							Description: "The JSON encoded string of the labels attached to this domain. " +
								"Wrap this resource in jsondecode() to use it as a Terraform data type.",
							Computed: true,
						},
						"annotations": schema.StringAttribute{
							CustomType: jsontypes.NormalizedType{},
							Description: "The JSON encoded string of the annotations attached to this domain. " +
								"Wrap this resource in jsondecode() to use it as a Terraform data type.",
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"domain_labels": schema.StringAttribute{
				Description: "Labels filter. Only domains that contain these labels will be returned as data source output.",
				CustomType:  jsontypes.NormalizedType{},
				Required:    true,
				Validators: []validator.String{
					utils.MustBeMapOfString{},
				},
			},
			"domain_annotations": schema.StringAttribute{
				Description: "Annotations filter. Only domains that contain these annotations will be returned as data source output.",
				CustomType:  jsontypes.NormalizedType{},
				Optional:    true,
				Validators: []validator.String{
					utils.MustBeMapOfString{},
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

	jsonData, err := json.Marshal(state.Payload())
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	domains, err := d.client.GetDomains(jsonData)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domains: %s", err))
		return
	}

	if len(domains) == 0 {
		resp.Diagnostics.AddWarning("No domains found. Please try again with the correct domain label filters.", "")

		state.Domains = make([]domain, 0)
		resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
		return
	}

	state.Domains, diags = domainApiModelToDataSourceModel(domains)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func domainApiModelToDataSourceModel(httpResp []*api.Domain) (domains []domain, diags diag.Diagnostics) {
	domains = make([]domain, 0)
	for _, domainResp := range httpResp {
		labels, err := json.Marshal(domainResp.Metadata.Labels)
		if err != nil {
			diags.AddError("Cannot marshal JSON", err.Error())
			return nil, diags
		}

		annotations, err := json.Marshal(domainResp.Metadata.Annotations)
		if err != nil {
			diags.AddError("Cannot marshal JSON", err.Error())
			return nil, diags
		}

		domains = append(domains, domain{
			Domain: types.StringValue(domainResp.Domain),
			Labels: jsontypes.NewNormalizedValue(string(labels)),
			Annotations: jsontypes.NewNormalizedValue(string(annotations)),
		})
	}

	return domains, nil
}
