package inkwasm

import (
	_ "unsafe"
)

const (
	TypeUndefined ObjectType = iota
	TypeNull
	TypeBoolean
	TypeNumber
	TypeBigInt
	TypeString
	TypeSymbol
	TypeFunction
	TypeObject
)

func (o ObjectType) String() string {
	switch o {
	case TypeUndefined:
		return "undefined"
	case TypeNull:
		return "null"
	case TypeBoolean:
		return "boolean"
	case TypeNumber:
		return "number"
	case TypeBigInt:
		return "bigint"
	case TypeString:
		return "string"
	case TypeSymbol:
		return "symbol"
	case TypeFunction:
		return "function"
	case TypeObject:
		return "object"
	default:
		return ""
	}
}

var (
	_Global    Object
	_Undefined Object
	_Null      Object
)

type ObjectType uint8

// Global returns an Object of globalThis
// https://developer.mozilla.org/pt-BR/docs/Web/JavaScript/Reference/Global_Objects/globalThis
func Global() Object {
	return _Global
}

// Undefined returns an Object of undefined
// https://developer.mozilla.org/pt-BR/docs/Web/JavaScript/Reference/Global_Objects/undefined
func Undefined() Object {
	return _Undefined
}

// Null returns an Object of null
// https://developer.mozilla.org/pt-BR/docs/Web/JavaScript/Reference/Global_Objects/null
func Null() Object {
	return _Null
}

func init() {
	_Global = getGlobal()
	_Undefined = getUndefined()
	_Null = getNull()

	_Global.protected = true
	_Undefined.protected = true
	_Null.protected = true
}

//inkwasm:get null
func getNull() Object

//inkwasm:get undefined
func getUndefined() Object

//inkwasm:get globalThis
func getGlobal() Object
