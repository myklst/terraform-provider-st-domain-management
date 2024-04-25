package utils

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// unit test scenario
/*
	1. create new key at root
	2. remove existing key at root
	3. modify value at root
	4. modify key at root -> two separate steps

	5. create new key in nested path -> should trigger update
	6. remove existing key in nested path -> should trigger update
	7. modify value in nested path -> should trigger update
*/

func TestCreateNewRoot(t *testing.T) {
	plan := json.RawMessage(`{"annotationA": "Hello", "annotationB": 69, "annotationC/annotationD": {"annotationE":"Bye"}, "annotationF":"newValue"}`)
	state := json.RawMessage(`{"annotationA": "Hello", "annotationB": 69, "annotationC/annotationD": {"annotationE":"Bye"}}`)
	var stateMapOfString map[string]any
	var planMapOfString map[string]any

	if err := json.Unmarshal(state, &stateMapOfString); err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal(plan, &planMapOfString); err != nil {
		t.Error(err)
	}

	test, err := JSONDiffToTerraformOperations(stateMapOfString, planMapOfString)
	if err != nil {
		t.Error(err)
	}

	assert := assert.New(t)
	assert.Equal(0, len(test.Update), "Count should be zero.")
	assert.Equal(0, len(test.Delete), "Count should be zero.")
	assert.Equal(1, len(test.Create), "Count should be one.")

	for k := range test.Create {
		assert.Equal("annotationF", k, "")
	}
}

func TestDeleteExistingRoot(t *testing.T) {
	plan := json.RawMessage(`{"annotationA": "Hello", "annotationB": 69}`)
	state := json.RawMessage(`{"annotationA": "Hello", "annotationB": 69, "annotationC/annotationD": {"annotationE":"Bye"}}`)
	var stateMapOfString map[string]any
	var planMapOfString map[string]any

	if err := json.Unmarshal(state, &stateMapOfString); err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal(plan, &planMapOfString); err != nil {
		t.Error(err)
	}

	test, err := JSONDiffToTerraformOperations(stateMapOfString, planMapOfString)
	if err != nil {
		t.Error(err)
	}

	assert := assert.New(t)
	assert.Equal(0, len(test.Create), "Count should be zero.")
	assert.Equal(0, len(test.Update), "Count should be zero.")
	assert.Equal(1, len(test.Delete), "Count should be one.")

	for k := range test.Create {
		assert.Equal("annotationC/annotationD", k, "")
	}
}

func TestUpdateExistingRootValue(t *testing.T) {
	plan := json.RawMessage(`{"annotationA": "Hello", "annotationB": 70, "annotationC/annotationD": {"annotationE":"Bye"}}`)
	state := json.RawMessage(`{"annotationA": "Hello", "annotationB": 69, "annotationC/annotationD": {"annotationE":"Bye"}}`)
	var stateMapOfString map[string]any
	var planMapOfString map[string]any

	if err := json.Unmarshal(state, &stateMapOfString); err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal(plan, &planMapOfString); err != nil {
		t.Error(err)
	}

	test, err := JSONDiffToTerraformOperations(stateMapOfString, planMapOfString)
	if err != nil {
		t.Error(err)
	}

	assert := assert.New(t)
	assert.Equal(0, len(test.Create), "Count should be zero.")
	assert.Equal(0, len(test.Delete), "Count should be zero.")
	assert.Equal(1, len(test.Update), "Count should be one.")

	for k := range test.Update {
		assert.Equal("annotationB", k, "")
	}
}

func TestModifyKeyAtRoot(t *testing.T) {
	plan := json.RawMessage(`{"annotationG": "Hello", "annotationB": 69, "annotationC/annotationD": {"annotationE":"Bye"}}`)
	state := json.RawMessage(`{"annotationA": "Hello", "annotationB": 69, "annotationC/annotationD": {"annotationE":"Bye"}}`)
	var stateMapOfString map[string]any
	var planMapOfString map[string]any

	if err := json.Unmarshal(state, &stateMapOfString); err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal(plan, &planMapOfString); err != nil {
		t.Error(err)
	}

	test, err := JSONDiffToTerraformOperations(stateMapOfString, planMapOfString)
	if err != nil {
		t.Error(err)
	}

	assert := assert.New(t)
	assert.Equal(1, len(test.Create), "Count should be one.")
	assert.Equal(1, len(test.Delete), "Count should be one.")
	assert.Equal(0, len(test.Update), "Count should be zero.")

	for k := range test.Create {
		assert.Equal("annotationG", k, "")
	}

	for k := range test.Delete {
		assert.Equal("annotationA", k, "")
	}
}

func TestCreateNewKeyInNestedPath(t *testing.T) {
	plan := json.RawMessage(`{"annotationA": "Hello", "annotationB": 69, "annotationC/annotationD": {"annotationE":"Bye", "annotationH":"World"}}`)
	state := json.RawMessage(`{"annotationA": "Hello", "annotationB": 69, "annotationC/annotationD": {"annotationE":"Bye"}}`)
	var stateMapOfString map[string]any
	var planMapOfString map[string]any

	if err := json.Unmarshal(state, &stateMapOfString); err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal(plan, &planMapOfString); err != nil {
		t.Error(err)
	}

	test, err := JSONDiffToTerraformOperations(stateMapOfString, planMapOfString)
	if err != nil {
		t.Error(err)
	}

	assert := assert.New(t)
	assert.Equal(0, len(test.Create), "Count should be zero.")
	assert.Equal(0, len(test.Delete), "Count should be zero.")
	assert.Equal(1, len(test.Update), "Count should be one.")

	for k := range test.Update {
		assert.Equal("annotationC/annotationD", k, "")
	}
}

func TestRemoveKeyInNestedPath(t *testing.T) {
	plan := json.RawMessage(`{"annotationA": "Hello", "annotationB": 69, "annotationC/annotationD": {}}`)
	state := json.RawMessage(`{"annotationA": "Hello", "annotationB": 69, "annotationC/annotationD": {"annotationE":"Bye"}}`)
	var stateMapOfString map[string]any
	var planMapOfString map[string]any

	if err := json.Unmarshal(state, &stateMapOfString); err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal(plan, &planMapOfString); err != nil {
		t.Error(err)
	}

	test, err := JSONDiffToTerraformOperations(stateMapOfString, planMapOfString)
	if err != nil {
		t.Error(err)
	}

	assert := assert.New(t)
	assert.Equal(0, len(test.Create), "Count should be zero.")
	assert.Equal(0, len(test.Delete), "Count should be zero.")
	assert.Equal(1, len(test.Update), "Count should be one.")

	for k := range test.Update {
		assert.Equal("annotationC/annotationD", k, "")
	}
}

func TestModifyValueInNestedPath(t *testing.T) {
	plan := json.RawMessage(`{"annotationA": "Hello", "annotationB": 69, "annotationC/annotationD": {"annotationE":"Welcome Back"}}`)
	state := json.RawMessage(`{"annotationA": "Hello", "annotationB": 69, "annotationC/annotationD": {"annotationE":"Bye"}}`)
	var stateMapOfString map[string]any
	var planMapOfString map[string]any

	if err := json.Unmarshal(state, &stateMapOfString); err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal(plan, &planMapOfString); err != nil {
		t.Error(err)
	}

	test, err := JSONDiffToTerraformOperations(stateMapOfString, planMapOfString)
	if err != nil {
		t.Error(err)
	}

	assert := assert.New(t)
	assert.Equal(0, len(test.Create), "Count should be zero.")
	assert.Equal(0, len(test.Delete), "Count should be zero.")
	assert.Equal(1, len(test.Update), "Count should be one.")

	for k := range test.Update {
		assert.Equal("annotationC/annotationD", k, "")
	}
}

func TestPathAdhereToRFC6902(t *testing.T) {
	plan := json.RawMessage(`{"annotationA": "Hello", "annotationB": 69, "annotationB~1annotationC~/annotationD": {"annotationE":"Welcome Back"}}`)
	state := json.RawMessage(`{"annotationA": "Hello", "annotationB": 69, "annotationC/annotationD": {"annotationE":"Bye"}}`)
	var stateMapOfString map[string]any
	var planMapOfString map[string]any

	if err := json.Unmarshal(state, &stateMapOfString); err != nil {
		t.Error(err)
	}

	if err := json.Unmarshal(plan, &planMapOfString); err != nil {
		t.Error(err)
	}

	test, err := JSONDiffToTerraformOperations(stateMapOfString, planMapOfString)
	if err != nil {
		t.Error(err)
	}

	assert := assert.New(t)
	assert.Equal(1, len(test.Create), "Count should be one.")
	assert.Equal(1, len(test.Delete), "Count should be one.")
	assert.Equal(0, len(test.Update), "Count should be zero.")

	for k := range test.Delete {
		assert.Equal("annotationC/annotationD", k, "")
	}

	for k := range test.Create {
		assert.Equal("annotationB~1annotationC~/annotationD", k, "")
	}
}
