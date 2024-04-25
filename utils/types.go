package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

func JSONToTerraformDynamicValue(input map[string]any) (types.Dynamic, diag.Diagnostics) {
	obj, diags := JSONToTerraformObject(input)
	if diags.HasError() {
		return types.DynamicNull(), diags
	}

	return types.DynamicValue(obj), nil
}

func JSONToTerraformObject(input map[string]any) (output types.Object, diags diag.Diagnostics) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				if strings.HasPrefix(x, "ObjectValueMust received error(s): Error | Missing Object Attribute Value") {
					diags.AddError("Complex types do not allow objects with different keys. All objects in list/set/tuple must have the same keys and nested keys.", x)
				} else if strings.EqualFold(x, "Array is empty") {
					diags.AddError("List / Tuple / Set / Array cannot be empty", x)
				} else if strings.EqualFold(x, "Null or nil is not allowed") {
					diags.AddError("Null or nil is not allowed", x)
				} else {
					diags.AddError("Unknown Panic", x)
				}
			default:
				diags.AddError("Unknown Panic", "")
			}
		}
	}()

	attrType, attrValue := jsonToTFTypes(input)

	return types.ObjectValue(attrType, attrValue)
}

// Converts a map[string]interface{} to a pair of map of attr.Type and attr.Value.
func jsonToTFTypes(input map[string]interface{}) (map[string]attr.Type, map[string]attr.Value) {
	resTypes := make(map[string]attr.Type)
	result := make(map[string]attr.Value)

	for key, value := range input {
		switch v := value.(type) {
		case nil:
			panic("Null or nil is not allowed")
		case bool:
			resTypes[key] = types.BoolType
			result[key] = types.BoolValue(v)
		case int:
			resTypes[key] = types.Int64Type
			result[key] = types.Int64Value(int64(v))
		case int64:
			resTypes[key] = types.Int64Type
			result[key] = types.Int64Value(int64(v))
		case float64:
			resTypes[key] = types.Float64Type
			result[key] = types.Float64Value(v)
		case string:
			resTypes[key] = types.StringType
			result[key] = types.StringValue(v)
		case []interface{}:
			if len(v) > 0 {
				switch v[0].(type) {
				case bool:
					resTypes[key] = types.ListType{
						ElemType: types.BoolType,
					}
					result[key], _ = types.ListValueFrom(context.Background(), types.BoolType, v)
				case int:
					resTypes[key] = types.ListType{
						ElemType: types.Int64Type,
					}
					result[key], _ = types.ListValueFrom(context.Background(), types.Int64Type, v)
				case int64:
					resTypes[key] = types.ListType{
						ElemType: types.Int64Type,
					}
					result[key], _ = types.ListValueFrom(context.Background(), types.Int64Type, v)
				case float64:
					resTypes[key] = types.ListType{
						ElemType: types.Float64Type,
					}
					result[key], _ = types.ListValueFrom(context.Background(), types.Float64Type, v)
				case string:
					resTypes[key] = types.ListType{
						ElemType: types.StringType,
					}
					result[key], _ = types.ListValueFrom(context.Background(), types.StringType, v)
				case map[string]interface{}:
					objType, _ := jsonToTFTypes(v[0].(map[string]interface{}))
					objArray := []basetypes.ObjectValue{}

					for _, i := range v {
						_, value := jsonToTFTypes(i.(map[string]interface{}))
						temp := basetypes.NewObjectValueMust(objType, value)
						objArray = append(objArray, temp)
					}

					var diags diag.Diagnostics

					resTypes[key] = types.ListType{
						ElemType: types.ObjectType{
							AttrTypes: objType,
						},
					}
					result[key], diags = types.ListValueFrom(context.Background(), types.ObjectType{
						AttrTypes: objType,
					}, objArray)
					if diags.HasError() {
						log.Println(diags.Errors())
						os.Exit(1)
					}

				default:
				}
			} else {
				panic("Array is empty")
			}

		case map[string]interface{}:
			temp1, temp2 := jsonToTFTypes(v)
			resTypes[key] = types.ObjectType{
				AttrTypes: temp1,
			}
			result[key], _ = types.ObjectValue(temp1, temp2)
		default:
			// Unsupported type, treat it as string
			resTypes[key] = types.StringType
			result[key] = types.StringValue(fmt.Sprintf("%v", value))
		}
	}
	return resTypes, result
}

func TFTypesToJSON(obj basetypes.ObjectValue) map[string]any {
	data := make(map[string]interface{})
	for key, value := range obj.Attributes() {
		data[key] = convertValue(value)
	}

	return data
}

func TFTypesToBytes(obj basetypes.ObjectValue) ([]byte, error) {
	return json.Marshal(TFTypesToJSON(obj))
}

// convertValue recursively converts the value to the appropriate Go type.
func convertValue(value attr.Value) interface{} {
	switch v := value.(type) {
	case types.Bool:
		return v.ValueBool()
	case types.Float64:
		return json.Number(strconv.FormatFloat(v.ValueFloat64(), 'f', -1, 64))
	case types.Int64:
		return json.Number(strconv.FormatInt(v.ValueInt64(), 10))
	case types.List:
		// Recursively convert nested Array
		return convertArray(v.Elements())
	case types.Map:
		// Recursively convert nested Map
		return convertMap(v)
	case types.Number:
		floatValue, _ := v.ValueBigFloat().Float64()
		return json.Number(strconv.FormatFloat(floatValue, 'f', -1, 64))
	case types.Object:
		// Recursively convert nested ObjectType
		return convertObject(v)
	case types.Set:
		// Recursively convert nested Array
		return convertArray(v.Elements())
	case types.String:
		return v.ValueString()
	case types.Tuple:
		// Recursively convert nested Array
		return convertArray(v.Elements())
	default:
		// Unsupported type, return nil or handle accordingly
		return nil
	}
}

// convertArray recursively converts the Array to a slice of appropriate Go types.
func convertArray(arr []attr.Value) []interface{} {
	result := make([]interface{}, len(arr))
	for i, v := range arr {
		result[i] = convertValue(v)
	}
	return result
}

// convertMap recursively converts the Map to a map with string keys and appropriate Go types as values.
func convertMap(m basetypes.MapValue) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range m.Elements() {
		result[key] = convertValue(value)
	}
	return result
}

func convertObject(obj types.Object) map[string]interface{} {
	result := make(map[string]interface{})

	for k, v := range obj.Attributes() {
		result[k] = convertValue(v)
	}
	return result
}
