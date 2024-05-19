package inkwasm

import (
	"fmt"
)

// VerifyArgs checks if the arguments are supported.
func VerifyArgs(args []interface{}) error {
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
			return fmt.Errorf("unsupported argument type of %T", a)
		}
	}
	return nil
}
