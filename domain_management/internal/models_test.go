package internal

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/myklst/terraform-provider-st-domain-management/api"
	"github.com/myklst/terraform-provider-st-domain-management/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPayloadCreation(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	includeLabels := map[string]interface{}{
		"uncommon/test": true,
	}
	excludeLabels := map[string]interface{}{
		"uncommon/unwanted": false,
	}
	includeAnnotations := map[string]interface{}{
		"annotationA": "valueA",
	}
	excludeAnnotations := map[string]interface{}{
		"annotationB": "valueB",
	}

	includeLabelsBytes, err := json.Marshal(includeLabels)
	require.NoError(err)
	excludeLabelsBytes, err := json.Marshal(excludeLabels)
	require.NoError(err)
	includeAnnotationsBytes, err := json.Marshal(includeAnnotations)
	require.NoError(err)
	excludeAnnotationsBytes, err := json.Marshal(excludeAnnotations)
	require.NoError(err)

	dynamicLabels, err := utils.JSONToTerraformDynamicValue(includeLabelsBytes)
	require.NoError(err)
	dynamicLabelsExclude, err := utils.JSONToTerraformDynamicValue(excludeLabelsBytes)
	require.NoError(err)
	dynamicAnnotations, err := utils.JSONToTerraformDynamicValue(includeAnnotationsBytes)
	require.NoError(err)
	dynamicAnnotationsExclude, err := utils.JSONToTerraformDynamicValue(excludeAnnotationsBytes)
	require.NoError(err)

	expected := api.DomainReq{
		FilterDomains: &api.IncludeExclude{
			Include: &api.Include{
				Metadata: &api.Metadata{
					Labels: includeLabels,
				},
			},
			Exclude: &api.Exclude{
				Metadata: &api.Metadata{},
			},
		},
		FilterSubdomains: &api.IncludeExclude{
			Include: &api.Include{
				Metadata: &api.Metadata{},
			},
			Exclude: &api.Exclude{
				Metadata: &api.Metadata{},
			},
		},
	}
	filter := FullDomainFilterDataSourceModel{
		DomainLabels: &Filters{
			Include: basetypes.NewDynamicValue(dynamicLabels.UnderlyingValue()),
		},
	}
	payload, err := filter.Payload()
	assert.NoError(err)
	assert.Equal(expected, payload, "Payload without annotations and exclude should match expected.")

	expected = api.DomainReq{
		FilterDomains: &api.IncludeExclude{
			Include: &api.Include{
				Metadata: &api.Metadata{
					Labels:      includeLabels,
					Annotations: includeAnnotations,
				},
			},
			Exclude: &api.Exclude{
				Metadata: &api.Metadata{},
			},
		},
		FilterSubdomains: &api.IncludeExclude{
			Include: &api.Include{
				Metadata: &api.Metadata{},
			},
			Exclude: &api.Exclude{
				Metadata: &api.Metadata{},
			},
		},
	}
	filter = FullDomainFilterDataSourceModel{
		DomainLabels: &Filters{
			Include: basetypes.NewDynamicValue(dynamicLabels.UnderlyingValue()),
		},
		DomainAnnotations: &Filters{
			Include: basetypes.NewDynamicValue(dynamicAnnotations.UnderlyingValue()),
		},
	}
	payload, err = filter.Payload()
	assert.NoError(err)
	assert.Equal(expected, payload, "Payload without exclude should match expected.")

	expected = api.DomainReq{
		FilterDomains: &api.IncludeExclude{
			Include: &api.Include{
				Metadata: &api.Metadata{
					Labels:      includeLabels,
					Annotations: includeAnnotations,
				},
			},
			Exclude: &api.Exclude{
				Metadata: &api.Metadata{
					Labels:      excludeLabels,
					Annotations: excludeAnnotations,
				},
			},
		},
		FilterSubdomains: &api.IncludeExclude{
			Include: &api.Include{
				Metadata: &api.Metadata{},
			},
			Exclude: &api.Exclude{
				Metadata: &api.Metadata{},
			},
		},
	}
	filter = FullDomainFilterDataSourceModel{
		DomainLabels: &Filters{
			Include: basetypes.NewDynamicValue(dynamicLabels.UnderlyingValue()),
			Exclude: basetypes.NewDynamicValue(dynamicLabelsExclude.UnderlyingValue()),
		},
		DomainAnnotations: &Filters{
			Include: basetypes.NewDynamicValue(dynamicAnnotations.UnderlyingValue()),
			Exclude: basetypes.NewDynamicValue(dynamicAnnotationsExclude.UnderlyingValue()),
		},
	}
	payload, err = filter.Payload()
	assert.NoError(err)
	assert.Equal(expected, payload, "Payload should match expected.")
}
