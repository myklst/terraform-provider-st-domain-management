package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var transformJSON = cmp.FilterValues(func(x, y []byte) bool {
	return json.Valid(x) && json.Valid(y)
}, cmp.Transformer("ParseJSON", func(in []byte) (out interface{}) {
	if err := json.Unmarshal(in, &out); err != nil {
		panic(err) // should never occur given previous filter to ensure valid JSON
	}
	return out
}))

func TestDynamicValueObject(t *testing.T) {
	jsonObj := map[string]interface{}{
		"boolValue":        true,
		"intValue":         42,
		"floatValue":       3.14,
		"stringValue":      "hello",
		"objectValue":      map[string]interface{}{"key": map[string]interface{}{"key2": map[string]interface{}{"key3": true}}},
		"arrayIntValue":    []interface{}{1, 2, 3},
		"arrayBoolValue":   []interface{}{true, false, true},
		"arrayStringValue": []interface{}{"hi", "there", "bye"},
		"arrayObjectValue": []interface{}{
			map[string]interface{}{"key4": map[string]interface{}{"key5": map[string]interface{}{"key6": true}}},
			map[string]interface{}{"key4": map[string]interface{}{"key5": map[string]interface{}{"key6": true}}},
			map[string]interface{}{"key4": map[string]interface{}{"key5": map[string]interface{}{"key6": true}}},
		},
	}

	types, value := jsonToTFTypes(jsonObj)
	obj, diags := basetypes.NewObjectValue(types, value)
	if diags.HasError() {
		log.Println(diags.Errors())
		os.Exit(1)
	}
	val, err := TFTypesToBytes(obj)
	if err != nil {
		t.Error(err)
	}

	bytes, err := json.Marshal(jsonObj)
	if err != nil {
		t.Error(err)
	}
	if !cmp.Equal(bytes, val) {
		fmt.Println(cmp.Diff(bytes, val, transformJSON))
		t.Error(err)
	}
}

func TestInconsistentArrayType(t *testing.T) {
	jsonObj := map[string]interface{}{
		"boolValue": true,
		"arrayObjectValue": []interface{}{
			map[string]interface{}{"key4": map[string]interface{}{"key5": map[string]interface{}{"key6": true}}},
			map[string]interface{}{"key4": map[string]interface{}{"key5": map[string]interface{}{"key6": true}}},
			map[string]interface{}{"key7": map[string]interface{}{"key8": map[string]interface{}{"key9": true}}},
		},
	}

	defer func() {
		r := recover()
		if r == nil {
			t.Errorf("The function should have panicked")
		}
		_ = r
	}()

	jsonToTFTypes(jsonObj)
}

func TestNullData(t *testing.T) {
	jsonObj := map[string]interface{}{
		"boolValue": true,
		"nullValue": nil,
	}

	_, diags := JSONToTerraformDynamicValue(jsonObj)

	if !diags.HasError() {
		t.Error("Diagnostics should have an error")
	}

	if diags[0].Summary() != "Null or nil is not allowed" {
		t.Error("Diagnostics error is unrelated. Not interested in this error.")
	}
}

func TestEmptyList(t *testing.T) {
	jsonObj := map[string]interface{}{
		"boolValue":        true,
		"arrayObjectValue": []interface{}{},
	}

	_, diags := JSONToTerraformDynamicValue(jsonObj)

	if !diags.HasError() {
		t.Error("Diagnostics should have an error")
	}

	if diags[0].Summary() != "List / Tuple / Set / Array cannot be empty" {
		t.Error("Diagnostics error is unrelated. Not interested in this error.")
	}
}

func TestEmptyObject(t *testing.T) {
	jsonObj := map[string]interface{}{
		"boolValue":        true,
		"emptyObjectValue": map[string]interface{}{"key": map[string]interface{}{"key2": map[string]interface{}{"key3": nil}}},
	}

	_, diags := JSONToTerraformDynamicValue(jsonObj)

	if !diags.HasError() {
		t.Error("Diagnostics should have an error")
	}

	if diags[0].Summary() != "Null or nil is not allowed" {
		t.Error("Diagnostics error is unrelated. Not interested in this error.")
	}
}

func TestHandlePanicCorrectly(t *testing.T) {
	jsonObj := map[string]interface{}{
		"boolValue": true,
		"arrayObjectValue": []interface{}{
			map[string]interface{}{"key4": map[string]interface{}{"key5": map[string]interface{}{"key6": true}}},
			map[string]interface{}{"key4": map[string]interface{}{"key5": map[string]interface{}{"key6": true}}},
			map[string]interface{}{"key7": map[string]interface{}{"key8": map[string]interface{}{"key9": true}}},
		},
	}

	_, diags := JSONToTerraformDynamicValue(jsonObj)

	if !diags.HasError() {
		t.Error("Diagnostics should have an error")
	}

	if diags[0].Summary() != "Complex types do not allow objects with different keys. All objects in list/set/tuple must have the same keys and nested keys." {
		t.Error("Diagnostics error is unrelated. Not interested in this error.")
	}

}
