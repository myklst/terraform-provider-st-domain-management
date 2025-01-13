package domain_management

import (
	"encoding/json"
	"testing"

	"github.com/myklst/terraform-provider-st-domain-management/api"
	"github.com/myklst/terraform-provider-st-domain-management/domain_management/internal"
	"github.com/myklst/terraform-provider-st-domain-management/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var domains = []api.DomainFull{
	{
		Domain: "a.com",
		Metadata: api.Metadata{
			Labels: map[string]interface{}{
				"common/brand": "sige",
				"common/env":   "basic",
			},
			Annotations: map[string]interface{}{
				"uncommon/test": false,
			},
		},
		Subdomains: []api.Subdomain{
			{
				Fqdn: "m.a.com",
				Name: "m",
				Metadata: api.Metadata{
					Labels: map[string]interface{}{
						"keyA": "valueA",
					},
				},
			},
			{
				Fqdn: "api.a.com",
				Name: "api",
				Metadata: api.Metadata{
					Labels: map[string]interface{}{
						"keyB": "valueB",
					},
				},
			},
		},
	},
	{
		Domain: "b.com",
		Metadata: api.Metadata{
			Labels: map[string]interface{}{
				"common/brand": "pg",
				"common/env":   "test",
			},
			Annotations: map[string]interface{}{
				"uncommon/test": false,
			},
		},
		Subdomains: []api.Subdomain{
			{
				Fqdn: "static.b.com",
				Name: "static",
				Metadata: api.Metadata{
					Labels: map[string]interface{}{
						"keyC": "valueC",
					},
				},
			},
			{
				Fqdn: "app.b.com",
				Name: "app",
				Metadata: api.Metadata{
					Labels: map[string]interface{}{
						"keyD": "valueD",
					},
				},
			},
		},
	},
}

func TestDomainFilter(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)
	labelA := map[string]interface{}{
		"keyA": "valueA",
	}
	labelB := map[string]interface{}{
		"keyB": "valueB",
	}
	labelE := map[string]interface{}{
		"keyE": "valueE",
	}

	labelABytes, err := json.Marshal(labelA)
	require.NoError(err)
	labelBBytes, err := json.Marshal(labelB)
	require.NoError(err)
	labelEBytes, err := json.Marshal(labelE)
	require.NoError(err)

	dynamicLabelsA, err := utils.JSONToTerraformDynamicValue(labelABytes)
	require.NoError(err)
	dynamicLabelsB, err := utils.JSONToTerraformDynamicValue(labelBBytes)
	require.NoError(err)
	dynamicLabelsE, err := utils.JSONToTerraformDynamicValue(labelEBytes)
	require.NoError(err)

	// Various test scenarios below
	expected := []api.DomainFull{}
	expected = append(expected, api.DomainFull{
		Domain:   domains[0].Domain,
		Metadata: domains[0].Metadata,
		Subdomains: []api.Subdomain{
			domains[0].Subdomains[0],
		},
	})
	actual, diags := domainFullFilter(domains, internal.Filters{
		Include: dynamicLabelsA,
	})
	require.False(diags.HasError())
	assert.Equal(expected, actual, "Should return domain a.com with one subdomain (m).")

	expected = []api.DomainFull{}
	expected = append(expected, api.DomainFull{
		Domain:   domains[0].Domain,
		Metadata: domains[0].Metadata,
		Subdomains: []api.Subdomain{
			domains[0].Subdomains[0],
		},
	})
	expected = append(expected, domains[1])
	actual, diags = domainFullFilter(domains, internal.Filters{
		Exclude: dynamicLabelsB,
	})
	require.False(diags.HasError())
	assert.Equal(expected, actual, "Should return domain a.com with one subdomains, and b.com with two subdomains.")

	expected = []api.DomainFull{}
	actual, diags = domainFullFilter(domains, internal.Filters{
		Include: dynamicLabelsE,
	})
	require.False(diags.HasError())
	assert.Equal(expected, actual, "Filtering for a wrong/non-existant label should return an empty slice.")

	actual, diags = domainFullFilter(domains, internal.Filters{
		Exclude: dynamicLabelsE,
	})
	require.False(diags.HasError())
	assert.Equal(domains, actual, "Excluding a wrong/non-existant label should not filer any results (return all results).")
}
