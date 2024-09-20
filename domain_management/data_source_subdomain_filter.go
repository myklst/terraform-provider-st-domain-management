package domain_management

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"

	"github.com/myklst/terraform-provider-st-domain-management/api"
	"github.com/myklst/terraform-provider-st-domain-management/utils"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type metadata struct {
	Labels map[string]interface{} `json:"labels"`
}

type subdomainFull struct {
	Name     string   `json:"name"`
	Metadata metadata `json:"metadata"`
}

type domainFullResp struct {
	Domain     string          `json:"domain"`
	Subdomains []subdomainFull `json:"subdomains"`
}

type domainsArray struct {
	Domains []domainFullResp `json:"dt"`
}

type subdomain struct {
	Name   types.String         `tfsdk:"name" json:"name"`
	Fqdn   types.String         `tfsdk:"fqdn" json:"fqdn"`
	Labels jsontypes.Normalized `tfsdk:"labels" json:"labels"`
}

type domain struct {
	Domain     types.String `tfsdk:"domain" json:"domain"`
	Subdomains []subdomain  `tfsdk:"subdomains" json:"subdomains"`
}

type subdomainFilterDataSourceModel struct {
	DomainLabels      jsontypes.Normalized `tfsdk:"domain_labels" json:"domain_labels"`
	DomainAnnotations jsontypes.Normalized `tfsdk:"domain_annotations" json:"domain_annotations"`
	SubdomainLabels   jsontypes.Normalized `tfsdk:"subdomain_labels" json:"subdomains_labels"`
	Domains           []domain             `tfsdk:"domains" json:"domains"`
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
			"domains": schema.SetNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Description: "The main domain of this result.",
							Computed:    true,
						},
						"subdomains": schema.SetNestedAttribute{
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"name": schema.StringAttribute{
										Description: "The name of subdomain.",
										Computed:    true,
									},
									"fqdn": schema.StringAttribute{
										Description: "The result of joining the subdomain name with the main domain.",
										Computed:    true,
									},
									"labels": schema.StringAttribute{
										CustomType: jsontypes.NormalizedType{},
										Description: "The JSON encoded string of the labels attachd to this subdomain. " +
											"Wrap this resource in jsonencode() to use it as a Terraform resource.",
										Computed: true,
									},
								},
							},
							Computed: true,
						},
					},
				},
				Computed: true,
			},
			"domain_labels": schema.StringAttribute{
				Description: "Domain labels filter. Only domains that contain these labels will be returned as data source output.",
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
			"subdomain_labels": schema.StringAttribute{
				Description: "Subdomain labels filter. Only subdomains that contain these labels will be returned as data source output",
				CustomType:  jsontypes.NormalizedType{},
				Required:    true,
				Validators: []validator.String{
					utils.MustBeMapOfString{},
					utils.MustNotBeNull{},
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

	response, err := d.client.GetDomainsFull(payload)
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

		resp.Diagnostics.AddError("HTTP Error", fmt.Sprintf("Got response %s: %s", strconv.Itoa(response.StatusCode), jsonBody["err"]))
		return
	}

	body, err := io.ReadAll(response.Body)

	if err != nil {
		resp.Diagnostics.AddError("Can not read response body", "")
		return
	}

	var domains domainsArray
	if err := json.Unmarshal(body, &domains); err != nil {
		resp.Diagnostics.AddError("Cannot unmarshal JSON", err.Error())
		return
	}

	for _, element := range domains.Domains {
		filterLabels, err := strconv.Unquote(state.SubdomainLabels.String())
		if err != nil {
			resp.Diagnostics.AddError("String Unquote Error: ", err.Error())
			return
		}

		convertedDomain, err := element.convertToTerraform(filterLabels)
		if err != nil {
			resp.Diagnostics.AddError("Error occured while converting api response to Terraform struct.", err.Error())
			return
		}

		if convertedDomain != nil {
			state.Domains = append(state.Domains, *convertedDomain)
		}
	}

	if len(state.Domains) == 0 {
		resp.Diagnostics.AddError("No subdomains found. Please try again with the correct domain and subdomain label filters.", "")
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Converts a domain response from the API into a domain for Terraform
func (d *domainFullResp) convertToTerraform(subdomainLabels string) (*domain, error) {
	subdomains := []subdomain{}

	for _, subdomain := range d.Subdomains {
		a, err := subdomain.convertToTerraformSubdomain(d.Domain, subdomainLabels)
		if err != nil {
			return nil, err
		}

		if a != nil {
			subdomains = append(subdomains, *a)
		}
	}

	if len(subdomains) == 0 {
		return nil, nil
	}

	return &domain{
		Domain:     types.StringValue(d.Domain),
		Subdomains: subdomains,
	}, nil
}

// Converts a subdomain response from the API into a subdomain for Terraform
func (s *subdomainFull) convertToTerraformSubdomain(domain string, subdomainLabelsFilter string) (*subdomain, error) {
	filter := map[string]interface{}{}
	err := json.Unmarshal([]byte(subdomainLabelsFilter), &filter)
	if err != nil {
		return nil, err
	}

	stringJson, err := json.Marshal(s.Metadata.Labels)
	if err != nil {
		return nil, err
	}

	// To determine whether the subdomain subdomain labels satisfies the data source filter,
	// a three step process is performed.
	// 1. Extract the map keys from the data source
	// 2. Use the same map keys to filter the map[string] from
	// 3. Ensure that the map[string] from data source and the map[string]
	//    from api response is the same
	apiResponse := map[string]interface{}{}
	for k := range filter {
		apiResponse[k] = s.Metadata.Labels[k]
	}

	// Return nil if subdomain labels filter's map contents (data source user input)
	// is not found in the subdomain from the api response
	if !reflect.DeepEqual(filter, apiResponse) {
		return nil, nil
	}

	return &subdomain{
		Name:   types.StringValue(s.Name),
		Fqdn:   types.StringValue(fmt.Sprintf("%s.%s", s.Name, domain)),
		Labels: jsontypes.NewNormalizedValue(string(stringJson)),
	}, nil
}
