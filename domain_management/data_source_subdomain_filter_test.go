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
						"keyA":             "valueA",
						"uncommon/testing": true,
					},
				},
			},
			{
				Fqdn: "api.a.com",
				Name: "api",
				Metadata: api.Metadata{
					Labels: map[string]interface{}{
						"keyB":             "valueB",
						"uncommon/testing": false,
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
						"keyC":             "valueC",
						"uncommon/testing": true,
					},
				},
			},
			{
				Fqdn: "app.b.com",
				Name: "app",
				Metadata: api.Metadata{
					Labels: map[string]interface{}{
						"keyD":             "valueD",
						"uncommon/testing": false,
					},
				},
			},
		},
	},
}

func TestFilterSubdomainsByIncludeLabels(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	labelA := map[string]interface{}{
		"keyA": "valueA",
	}

	labelABytes, err := json.Marshal(labelA)
	require.NoError(err)

	dynamicLabelsA, err := utils.JSONToTerraformDynamicValue(labelABytes)
	require.NoError(err)

	expected := []api.DomainFull{{
		Domain:   domains[0].Domain,
		Metadata: domains[0].Metadata,
		Subdomains: []api.Subdomain{
			domains[0].Subdomains[0],
		},
	}}

	actual, diags := domainFullFilter(domains, internal.Filters{
		Include: dynamicLabelsA,
	})
	require.False(diags.HasError())
	assert.Equal(expected, actual, "Should return domain a.com with one subdomain (m).")
}

func TestFilterSubdomainsByExcludeLabels(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	labelB := map[string]interface{}{
		"keyB": "valueB",
	}

	labelBBytes, err := json.Marshal(labelB)
	require.NoError(err)

	dynamicLabelsB, err := utils.JSONToTerraformDynamicValue(labelBBytes)
	require.NoError(err)

	expected := []api.DomainFull{{
		Domain:   domains[0].Domain,
		Metadata: domains[0].Metadata,
		Subdomains: []api.Subdomain{
			domains[0].Subdomains[0],
		},
	}, domains[1]}

	actual, diags := domainFullFilter(domains, internal.Filters{
		Exclude: dynamicLabelsB,
	})
	require.False(diags.HasError())
	assert.Equal(expected, actual, "Should return domain a.com with one subdomains (m), and b.com with two subdomains.")
}

func TestFilterSubdomainsByExcludeMultipleLabels(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	labels := map[string]interface{}{
		"keyE":             "valueE",
		"uncommon/testing": false,
	}

	labelsBytes, err := json.Marshal(labels)
	require.NoError(err)

	dynamicLabels, err := utils.JSONToTerraformDynamicValue(labelsBytes)
	require.NoError(err)

	expected := domains
	expected[0].Subdomains = []api.Subdomain{domains[0].Subdomains[0]}
	expected[1].Subdomains = []api.Subdomain{domains[1].Subdomains[0]}

	actual, diags := domainFullFilter(domains, internal.Filters{
		Exclude: dynamicLabels,
	})
	require.False(diags.HasError())
	assert.Equal(expected, actual, "Should return domain a.com with one subdomains, and b.com with one subdomain.")
}

func TestFilterSubdomainsByWrongIncludeLabels(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	labelE := map[string]interface{}{
		"keyE": "valueE",
	}

	labelEBytes, err := json.Marshal(labelE)
	require.NoError(err)

	dynamicLabelsE, err := utils.JSONToTerraformDynamicValue(labelEBytes)
	require.NoError(err)

	expected := []api.DomainFull{}

	actual, diags := domainFullFilter(domains, internal.Filters{
		Include: dynamicLabelsE,
	})
	require.False(diags.HasError())
	assert.Equal(expected, actual, "Filtering for a wrong/non-existant include label should return an empty slice.")
}

func TestIncludeExcludeSameLabels(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	labels := map[string]interface{}{
		"keyA": "valueA",
	}

	labelsBytes, err := json.Marshal(labels)
	require.NoError(err)

	dynamicLabels, err := utils.JSONToTerraformDynamicValue(labelsBytes)
	require.NoError(err)

	expected := []api.DomainFull{}
	actual, diags := domainFullFilter(domains, internal.Filters{
		Include: dynamicLabels,
		Exclude: dynamicLabels,
	})
	require.False(diags.HasError())
	assert.Equal(expected, actual, "Should return empty list.")
}

func TestFilterWithBothIncludeAndExclude(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	includeLabels := map[string]interface{}{
		"keyA": "valueA",
	}
	excludeLabels := map[string]interface{}{
		"keyA":             "valueA",
		"uncommon/testing": false,
		"testLabel1":       1,
	}

	includeBytes, err := json.Marshal(includeLabels)
	require.NoError(err)
	excludeBytes, err := json.Marshal(excludeLabels)
	require.NoError(err)

	dynamicInclude, err := utils.JSONToTerraformDynamicValue(includeBytes)
	require.NoError(err)
	dynamicExclude, err := utils.JSONToTerraformDynamicValue(excludeBytes)
	require.NoError(err)

	expected := []api.DomainFull{}

	actual, diags := domainFullFilter(domains, internal.Filters{
		Include: dynamicInclude,
		Exclude: dynamicExclude,
	})

	require.False(diags.HasError())
	assert.Equal(expected, actual, "Should return empty list.")
}
