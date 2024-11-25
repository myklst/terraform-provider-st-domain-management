package utils

import (
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Converts a json marshaled byte array into a Terraform Data with Dynamic Type
//
// Adapted from
// https://github.com/magodo/terraform-provider-restful/blob/
// eb875adeb0967a0a3cd7393d8eb2016a2642ac0f/internal/dynamic/dynamic.go#L284
func JSONToTerraformDynamicValue(b []byte) (types.Dynamic, error) {
	if len(b) == 0 {
		return types.DynamicNull(), nil
	}
	_, v, err := jsonToTFTypes(b)
	if err != nil {
		return types.Dynamic{}, err
	}
	return types.DynamicValue(v), nil
}

func jsonToTFTypes(b []byte) (attr.Type, attr.Value, error) {
	if string(b) == "null" {
		return types.DynamicType, types.DynamicNull(), nil
	}

	var object map[string]json.RawMessage
	if err := json.Unmarshal(b, &object); err == nil {
		attrTypes := map[string]attr.Type{}
		attrVals := map[string]attr.Value{}
		for k, v := range object {
			attrTypes[k], attrVals[k], err = jsonToTFTypes(v)
			if err != nil {
				return nil, nil, err
			}
		}
		typ := types.ObjectType{AttrTypes: attrTypes}
		val, diags := types.ObjectValue(attrTypes, attrVals)
		if diags.HasError() {
			diag := diags.Errors()[0]
			return nil, nil, fmt.Errorf("%s: %s", diag.Summary(), diag.Detail())
		}
		return typ, val, nil
	}

	var array []json.RawMessage
	if err := json.Unmarshal(b, &array); err == nil {
		eTypes := []attr.Type{}
		eVals := []attr.Value{}
		for _, e := range array {
			eType, eVal, err := jsonToTFTypes(e)
			if err != nil {
				return nil, nil, err
			}
			eTypes = append(eTypes, eType)
			eVals = append(eVals, eVal)
		}
		typ := types.TupleType{ElemTypes: eTypes}
		val, diags := types.TupleValue(eTypes, eVals)
		if diags.HasError() {
			diag := diags.Errors()[0]
			return nil, nil, fmt.Errorf("%s: %s", diag.Summary(), diag.Detail())
		}
		return typ, val, nil
	}

	// Primitives
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return nil, nil, fmt.Errorf("failed to unmarshal %s: %v", string(b), err)
	}

	switch v := v.(type) {
	case bool:
		return types.BoolType, types.BoolValue(v), nil
	case float64:
		return types.NumberType, types.NumberValue(big.NewFloat(v)), nil
	case string:
		return types.StringType, types.StringValue(v), nil
	default:
		return nil, nil, fmt.Errorf("Unhandled type: %T", v)
	}
}
