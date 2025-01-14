package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var jsonObj = map[string]interface{}{
	"boolValue":        true,
	"intValue":         float64(42),
	"floatValue":       3.14,
	"stringValue":      "hello",
	"objectValue":      map[string]interface{}{"key": map[string]interface{}{"key2": map[string]interface{}{"key3": true}}},
	"arrayIntValue":    []interface{}{float64(1), float64(2), float64(3)},
	"arrayBoolValue":   []interface{}{true, false, true},
	"arrayStringValue": []interface{}{"hi", "there", "bye"},
	"arrayObjectValue": []interface{}{
		map[string]interface{}{"key4": map[string]interface{}{"key5": map[string]interface{}{"key6": true}}},
		map[string]interface{}{"key4": map[string]interface{}{"key5": map[string]interface{}{"key6": true}}},
		map[string]interface{}{"key4": map[string]interface{}{"key5": map[string]interface{}{"key6": true}}},
	},
}

func TestDynamicValueObject(t *testing.T) {
	jsonBytes, err := json.Marshal(jsonObj)
	if err != nil {
		t.Error(err)
	}
	_, err = JSONToTerraformDynamicValue(jsonBytes)
	if err != nil {
		t.Error(err)
	}
}

func TestInconsistentArrayType(t *testing.T) {
	jsonObj := map[string]interface{}{
		"boolValue": true,
		"arrayObjectValue": []interface{}{
			map[string]interface{}{"key4": map[string]interface{}{"key5": map[string]interface{}{"key6": true}}},
			map[string]interface{}{"key4": map[string]interface{}{"key5": map[string]interface{}{"key6": true}}},
			map[string]interface{}{"key7": []bool{true, false, true}},
			[]bool{false, true, false},
		},
	}

	bytes, err := json.Marshal(jsonObj)
	assert.Nil(t, err)

	_, err = JSONToTerraformDynamicValue(bytes)
	assert.Nil(t, err)
}

func TestNullData(t *testing.T) {
	jsonObj := map[string]interface{}{
		"boolValue": true,
		"nullValue": nil,
	}

	bytes, err := json.Marshal(jsonObj)
	assert.Nil(t, err)

	_, err = JSONToTerraformDynamicValue(bytes)
	assert.Nil(t, err)
}

func TestEmptyList(t *testing.T) {
	jsonObj := map[string]interface{}{
		"boolValue":        true,
		"arrayObjectValue": []interface{}{},
	}

	bytes, err := json.Marshal(jsonObj)
	assert.Nil(t, err)

	_, err = JSONToTerraformDynamicValue(bytes)
	assert.Nil(t, err)
}

func TestEmptyObject(t *testing.T) {
	jsonObj := map[string]interface{}{
		"boolValue":        true,
		"emptyObjectValue": map[string]interface{}{"key": map[string]interface{}{"key2": map[string]interface{}{"key3": nil}}},
	}

	bytes, err := json.Marshal(jsonObj)
	assert.Nil(t, err)

	_, err = JSONToTerraformDynamicValue(bytes)
	assert.Nil(t, err)
}

func TestTFTypesToJSON(t *testing.T) {
	require := require.New(t)
	jsonObjBytes, err := json.Marshal(jsonObj)
	require.NoError(err)

	dynamicTFObj, err := JSONToTerraformDynamicValue(jsonObjBytes)
	require.NoError(err)

	actualJson, err := TFTypesToJSON(dynamicTFObj)
	require.NoError(err)

	assert.Equal(t, jsonObj, actualJson)
}
