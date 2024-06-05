package utils

import (
	"errors"
	"strings"

	"gomodules.xyz/jsonpatch/v2"
)

type UpdateOperations struct {
	Create map[string]jsonpatch.Operation
	Update map[string]jsonpatch.Operation
	Delete map[string]jsonpatch.Operation
}

// Calculates the diff between state and plan objects. The output of this
// function is arrays of jsondiff.Ops that has been modified to suit the needs of
// Terraform CRUD operation. The following modifications are:
// Updating a JSON object can be classified into three distinct operations.
//  1. Creating a new key value pair at the root.
//  2. Updating a root value. Creating a new key value pair in a nested path will
//     be considered as updating a root value.
//  3. Deleting an existing key value pair from the root.
//  4. Values inside the jsondiff.Operation are discarded and disregarded.
//  5. Only the paths are important.
//  6. It is considered a root key if the path contains only one slash "/xxx"
func JSONDiffToTerraformOperations(state, plan []byte) (UpdateOperations, error) {
	patch, err := jsonpatch.CreatePatch(state, plan)
	if err != nil {
		return UpdateOperations{}, err
	}

	op := UpdateOperations{
		Create: map[string]jsonpatch.Operation{},
		Update: map[string]jsonpatch.Operation{},
		Delete: map[string]jsonpatch.Operation{},
	}
	for _, v := range patch {

		// First, trim the starting empty element.
		// Check the length of the json path. If it is longer than 1,
		// It is considered as a nested object.
		// String replacements are required. Refer to RFC6902.
		array := strings.Split(strings.Trim(v.Path, "/"), "/")
		if len(array) > 1 {
			finalPath := ProcessString(array[0])
			operation := jsonpatch.Operation{}
			op.Update[finalPath] = operation
		}

		// If length is one, then it is considered a root key
		if len(array) == 1 {
			v.Path = ProcessString(array[0])
			switch v.Operation {
			case "add":
				{
					op.Create[v.Path] = v
				}
			case "remove":
				{
					op.Delete[v.Path] = v
				}
			case "copy":
				{
					return UpdateOperations{}, errors.New("JsonDiff OperationCopy not supported")
				}
			case "move":
				{
					return UpdateOperations{}, errors.New("JsonDiff OperationMove not supported")
				}
			case "replace":
				{
					op.Update[v.Path] = v
				}
			}
		}
	}

	return op, nil
}

// Process string according to RFC6902 standard
// 1. "~0" will be converted back to "~"
// 2. "~1" will be converted back to "/"
func ProcessString(input string) (output string) {
	output = strings.ReplaceAll(input, "~1", "/")
	output = strings.ReplaceAll(output, "~0", "~")
	return
}
