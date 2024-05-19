package inkwasm

import (
	"fmt"
)

// createArgs is now a no-op
// The JS will be responsible to identify the type of the arguments,
// using Go internal structures.
//
// That is used to avoid wrong types of arguments.
func createArgs(args []interface{}) ([]interface{}, error) {
	for _, a := range args {
		switch a.(type) {
		case Object:
		case string:
		case float32:
		case float64:
		case uintptr:
		case bool:
		case int:
		case uint:
		case uint8:
		case int8:
		case uint16:
		case int16:
		case uint32:
		case int32:
		case uint64:
		case int64:
		case []uint8:
		case []int8:
		case []uint16:
		case []int16:
		case []uint32:
		case []int32:
		case []uint64:
		case []int64:
		default:
			return nil, fmt.Errorf("unsupported argument type of %T", a)
		}
	}
	return args, nil
}
