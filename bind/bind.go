package bind

import (
	"fmt"
)

type Hint string

var (
	HintFunc   Hint = "func"
	HintNew    Hint = "new"
	HintGet    Hint = "get"
	HintSet    Hint = "set"
	HintExport Hint = "export"
)

type ArgumentMode uint32

const (
	ModeStatic ArgumentMode = iota + 1
	ModePointer
	ModeArray
	ModeSlice
)

type BridgeFuncInfo struct {
	JS   string
	Size int
}

var BridgeFunc = map[ArgumentMode]map[string]BridgeFuncInfo{
	ModeStatic: {
		"float32":        {JS: "globalThis.inkwasm.Load.Float32", Size: 4},
		"float64":        {JS: "globalThis.inkwasm.Load.Float64", Size: 8},
		"uintptr":        {JS: "globalThis.inkwasm.Load.UintPtr", Size: 8},
		"byte":           {JS: "globalThis.inkwasm.Load.Byte", Size: 1},
		"bool":           {JS: "globalThis.inkwasm.Load.Bool", Size: 1},
		"int":            {JS: "globalThis.inkwasm.Load.Int", Size: 8},
		"uint":           {JS: "globalThis.inkwasm.Load.Uint", Size: 8},
		"uint8":          {JS: "globalThis.inkwasm.Load.Uint8", Size: 1},
		"uint16":         {JS: "globalThis.inkwasm.Load.Uint16", Size: 2},
		"uint32":         {JS: "globalThis.inkwasm.Load.Uint32", Size: 4},
		"uint64":         {JS: "globalThis.inkwasm.Load.Uint64", Size: 8},
		"int8":           {JS: "globalThis.inkwasm.Load.Int8", Size: 1},
		"int16":          {JS: "globalThis.inkwasm.Load.Int16", Size: 2},
		"int32":          {JS: "globalThis.inkwasm.Load.Int32", Size: 4},
		"int64":          {JS: "globalThis.inkwasm.Load.Int64", Size: 8},
		"string":         {JS: "globalThis.inkwasm.Load.String", Size: 16},
		"rune":           {JS: "globalThis.inkwasm.Load.Rune", Size: 8},
		"unsafe.pointer": {JS: "globalThis.inkwasm.Load.UnsafePointer", Size: 8},
		"inkwasm.object": {JS: "globalThis.inkwasm.Load.InkwasmObject", Size: 16},
	},
	ModeArray: {
		"default": {JS: "globalThis.inkwasm.Load.Array", Size: -1},
		"float32": {JS: "globalThis.inkwasm.Load.ArrayFloat32", Size: 4},
		"float64": {JS: "globalThis.inkwasm.Load.ArrayFloat64", Size: 8},
		"uintptr": {JS: "globalThis.inkwasm.Load.ArrayUintPtr", Size: 8},
		"byte":    {JS: "globalThis.inkwasm.Load.ArrayByte", Size: 1},
		"uint8":   {JS: "globalThis.inkwasm.Load.ArrayUint8", Size: 1},
		"uint16":  {JS: "globalThis.inkwasm.Load.ArrayUint16", Size: 2},
		"uint32":  {JS: "globalThis.inkwasm.Load.ArrayUint32", Size: 4},
		"uint64":  {JS: "globalThis.inkwasm.Load.ArrayUint64", Size: 8},
		"int8":    {JS: "globalThis.inkwasm.Load.ArrayInt8", Size: 1},
		"int16":   {JS: "globalThis.inkwasm.Load.ArrayInt16", Size: 2},
		"int32":   {JS: "globalThis.inkwasm.Load.ArrayInt32", Size: 4},
		"int64":   {JS: "globalThis.inkwasm.Load.ArrayInt64", Size: 8},
		"rune":    {JS: "globalThis.inkwasm.Load.ArrayRune", Size: 8},
	},
	ModeSlice: {
		"default": {JS: "globalThis.inkwasm.Load.Slice", Size: 24},
	},
	ModePointer: {
		"default": {JS: "globalThis.inkwasm.Load.Ptr", Size: 8},
	},
}

var ResultFunc = map[ArgumentMode]map[string]BridgeFuncInfo{
	ModeStatic: {
		"float32":        {JS: "globalThis.inkwasm.Set.Float32", Size: 4},
		"float64":        {JS: "globalThis.inkwasm.Set.Float64", Size: 8},
		"uintptr":        {JS: "globalThis.inkwasm.Set.UintPtr", Size: 8},
		"byte":           {JS: "globalThis.inkwasm.Set.Byte", Size: 1},
		"bool":           {JS: "globalThis.inkwasm.Set.Bool", Size: 1},
		"int":            {JS: "globalThis.inkwasm.Set.Int", Size: 8},
		"uint":           {JS: "globalThis.inkwasm.Set.Uint", Size: 8},
		"uint8":          {JS: "globalThis.inkwasm.Set.Uint8", Size: 1},
		"uint16":         {JS: "globalThis.inkwasm.Set.Uint16", Size: 2},
		"uint32":         {JS: "globalThis.inkwasm.Set.Uint32", Size: 4},
		"uint64":         {JS: "globalThis.inkwasm.Set.Uint64", Size: 8},
		"int8":           {JS: "globalThis.inkwasm.Set.Int8", Size: 1},
		"int16":          {JS: "globalThis.inkwasm.Set.Int16", Size: 2},
		"int32":          {JS: "globalThis.inkwasm.Set.Int32", Size: 4},
		"int64":          {JS: "globalThis.inkwasm.Set.Int64", Size: 8},
		"string":         {JS: "globalThis.inkwasm.Set.String", Size: 16},
		"rune":           {JS: "globalThis.inkwasm.Set.Rune", Size: 8},
		"big.int":        {JS: "globalThis.inkwasm.Set.BigInt", Size: 32},
		"unsafe.pointer": {JS: "globalThis.inkwasm.Set.UnsafePointer", Size: 8},
		"inkwasm.object": {JS: "globalThis.inkwasm.Set.InkwasmObject", Size: 16},
	},
	ModeArray: {
		"default": {JS: "globalThis.inkwasm.Set.Array", Size: -1},
		"float32": {JS: "globalThis.inkwasm.Set.Float32", Size: 4},
		"float64": {JS: "globalThis.inkwasm.Set.Float64", Size: 8},
		"uintptr": {JS: "globalThis.inkwasm.Set.UintPtr", Size: 8},
		"byte":    {JS: "globalThis.inkwasm.Set.Byte", Size: 1},
		"bool":    {JS: "globalThis.inkwasm.Set.Bool", Size: 1},
		"int":     {JS: "globalThis.inkwasm.Set.Int", Size: 8},
		"uint":    {JS: "globalThis.inkwasm.Set.Uint", Size: 8},
		"uint8":   {JS: "globalThis.inkwasm.Set.Uint8", Size: 1},
		"uint16":  {JS: "globalThis.inkwasm.Set.Uint16", Size: 2},
		"uint32":  {JS: "globalThis.inkwasm.Set.Uint32", Size: 4},
		"uint64":  {JS: "globalThis.inkwasm.Set.Uint64", Size: 8},
		"int8":    {JS: "globalThis.inkwasm.Set.Int8", Size: 1},
		"int16":   {JS: "globalThis.inkwasm.Set.Int16", Size: 2},
		"int32":   {JS: "globalThis.inkwasm.Set.Int32", Size: 4},
		"int64":   {JS: "globalThis.inkwasm.Set.Int64", Size: 8},
	},
	ModeSlice: {
		"default": {JS: "globalThis.inkwasm.Set.Slice", Size: 16}, // Size is 16 because it's Object
	},
	ModePointer: {},
}

type Argument struct {
	Name    string
	Tag     string
	ArgType ArgumentMode
	Type    string
	SubType *Argument
	Len     uint64 // Len for Array
}

type Package struct {
	Name string
	Dir  string
	Path string
}

type Function struct {
	File   string
	Line   int
	IsTest bool
	FunctionGolang
	FunctionJavascript
}

type FunctionJavascript struct {
	Name string
	Hint Hint
}

type FunctionGolang struct {
	Name      string
	Arguments []Argument
	Result    []Argument
}

func (f *Function) CreateError(format string, a ...interface{}) error {
	return fmt.Errorf("%s:%d: %s", f.File, f.Line, fmt.Sprintf(format, a...))
}
