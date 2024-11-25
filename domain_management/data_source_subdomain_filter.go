package domain_management

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/myklst/terraform-provider-st-domain-management/api"
	"github.com/myklst/terraform-provider-st-domain-management/utils"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	DomainLabels      jsontypes.Normalized   `tfsdk:"domain_labels" json:"domain_labels"`
	DomainAnnotations jsontypes.Normalized   `tfsdk:"domain_annotations" json:"domain_annotations"`
	SubdomainLabels   jsontypes.Normalized   `tfsdk:"subdomain_labels" json:"subdomains_labels"`
	Domains           basetypes.DynamicValue `tfsdk:"domains" json:"domains"`
}

func (d *subdomainFilterDataSourceModel) Payload() ([]byte, diag.Diagnostics) {
	domains := map[string]any{}
	labels := map[string]any{}
	annotations := map[string]any{}

	diags := d.DomainLabels.Unmarshal(&labels)
	if diags.HasError() {
		return nil, diags
	}
	domains["labels"] = labels

	if !d.DomainAnnotations.IsNull() {
		diags = d.DomainAnnotations.Unmarshal(&annotations)
		if diags.HasError() {
			return nil, diags
		}
		domains["annotations"] = annotations
	}

	payload := map[string]any{}
	payload["metadata"] = domains

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		diags.AddError("Cannot marshal payload. ", err.Error())
		return nil, diags
	}

	return payloadBytes, nil
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
				Computed: true,
			},
			"domain_labels": schema.StringAttribute{
				Description: "Domain labels filter. Only domains that contain these labels will be returned as data source output.",
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
			"subdomain_labels": schema.StringAttribute{
				Description: "Subdomain labels filter. Only subdomains that contain these labels will be returned as data source output",
				CustomType:  jsontypes.NormalizedType{},
				Required:    true,
				Validators: []validator.String{
					utils.MustBeMapOfString{},
				},
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

	payload, diags := state.Payload()
	if diags.HasError() {
		return
	}

	domainsFullBytes, err := d.client.GetDomainsFull(payload)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read domains, got error: %s", err))
		return
	}

	domainsFull := []*api.DomainFull{}
	err = json.Unmarshal(domainsFullBytes, &domainsFull)
	if err != nil {
		resp.Diagnostics.AddError(err.Error(), "")
		return
	}

	domains, diags := domainFullFilter(domainsFull, state.SubdomainLabels.ValueString())
	if diags.HasError() {
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

func domainFullFilter(httpResp []*api.DomainFull, subdomainLabels string) (domainsFull []domainFull, diags diag.Diagnostics) {
	domainsFull = make([]domainFull, 0)
	for _, domainResp := range httpResp {
		subdomains := []subdomain{}
		for _, subdomainResp := range domainResp.Subdomains {
			if len(subdomainResp.Metadata.Labels) == 0 {
				continue
			}

			subdomain, err := subdomainsFilter(subdomainResp, domainResp.Domain, subdomainLabels)
			if err != nil {
				diags = append(diags, diag.NewErrorDiagnostic("Cannot convert subdomain api model to Terraform", err.Error()))
				return
			}

			if subdomain == nil {
				continue
			}

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

func subdomainsFilter(subdomainResp *api.Subdomain, domain string, subdomainLabelsFilter string) (*subdomain, error) {
	// To determine whether the subdomain labels satisfies the filter in the data source input,
	// a three step process is performed.
	// 1. Unmarshal the filter input into a map[string]interface
	// 2. For each map key, use the map key to access the labels map[string] from the api response
	// 3. Ensure that the map[string] from data source and the map[string] from api response is the same
	filter := map[string]interface{}{}
	err := json.Unmarshal([]byte(subdomainLabelsFilter), &filter)
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

	return &subdomain{
		Name:   subdomainResp.Name,
		Fqdn:   strings.Join([]string{subdomainResp.Name, domain}, "."),
		Labels: subdomainResp.Metadata.Labels,
	}, nil
}
