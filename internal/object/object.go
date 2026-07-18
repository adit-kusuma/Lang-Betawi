package object

import (
	"fmt"
	"strconv"
	"strings"

	"language-betawi/internal/ast"
)

type ObjectType string

const (
	INTEGER_OBJ      ObjectType = "INTEGER"
	FLOAT_OBJ        ObjectType = "FLOAT"
	STRING_OBJ       ObjectType = "STRING"
	BOOLEAN_OBJ      ObjectType = "BOOLEAN"
	NULL_OBJ         ObjectType = "NULL"
	ARRAY_OBJ        ObjectType = "ARRAY"
	MAP_OBJ          ObjectType = "MAP"
	FUNCTION_OBJ     ObjectType = "FUNCTION"
	RETURN_VALUE_OBJ ObjectType = "RETURN_VALUE"
	ERROR_OBJ        ObjectType = "ERROR"
)

var BetawiTypeName = map[ObjectType]string{
	INTEGER_OBJ:      "biji",
	FLOAT_OBJ:        "biji desimal",
	STRING_OBJ:       "bacotan",
	BOOLEAN_OBJ:      "bener/kagak",
	NULL_OBJ:         "zonk",
	ARRAY_OBJ:        "kropak",
	MAP_OBJ:          "kropak berlabel",
	FUNCTION_OBJ:     "gaya",
	RETURN_VALUE_OBJ: "balikan",
	ERROR_OBJ:        "kapiran",
}

func DisplayName(t ObjectType) string {
	if name, ok := BetawiTypeName[t]; ok {
		return name
	}
	return string(t)
}

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct{ Value int64 }

func (i *Integer) Type() ObjectType { return INTEGER_OBJ }
func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }

type Float struct{ Value float64 }

func (f *Float) Type() ObjectType { return FLOAT_OBJ }
func (f *Float) Inspect() string  { return strconv.FormatFloat(f.Value, 'f', -1, 64) }

type String struct{ Value string }

func (s *String) Type() ObjectType { return STRING_OBJ }
func (s *String) Inspect() string  { return s.Value }

type Boolean struct{ Value bool }

func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }
func (b *Boolean) Inspect() string {
	if b.Value {
		return "bener"
	}
	return "kagak"
}

type Null struct{}

func (n *Null) Type() ObjectType { return NULL_OBJ }
func (n *Null) Inspect() string  { return "zonk" }

type Array struct {
	Elements []Object
}

func (a *Array) Type() ObjectType { return ARRAY_OBJ }
func (a *Array) Inspect() string {
	var els []string
	for _, e := range a.Elements {
		els = append(els, e.Inspect())
	}
	return "[" + strings.Join(els, ", ") + "]"
}

type Map struct {
	Keys  []string
	Pairs map[string]Object
}

func NewMap() *Map {
	return &Map{Pairs: make(map[string]Object)}
}

func (m *Map) Set(key string, val Object) {
	if _, exists := m.Pairs[key]; !exists {
		m.Keys = append(m.Keys, key)
	}
	m.Pairs[key] = val
}

func (m *Map) Type() ObjectType { return MAP_OBJ }
func (m *Map) Inspect() string {
	var pairs []string
	for _, k := range m.Keys {
		pairs = append(pairs, fmt.Sprintf("%s: %s", k, m.Pairs[k].Inspect()))
	}
	return "{" + strings.Join(pairs, ", ") + "}"
}

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Type() ObjectType { return FUNCTION_OBJ }
func (f *Function) Inspect() string {
	var params []string
	for _, p := range f.Parameters {
		params = append(params, p.Value)
	}
	return "bikin_gaya(" + strings.Join(params, ", ") + ") { ... }"
}

type ReturnValue struct{ Value Object }

func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }
func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }

type Error struct {
	Message string
	Line    int
	Column  int
}

func (e *Error) Type() ObjectType { return ERROR_OBJ }
func (e *Error) Inspect() string  { return e.Message }
