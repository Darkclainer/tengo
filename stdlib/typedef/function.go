package main

import (
	"fmt"
	"go/types"
	"strings"
)

type Function struct {
	name    string
	input   *Signature
	wrapper *Signature
}

func NewFunction(sign *types.Signature) (*Function, error) {
	inputSign, err := NewSignature(sign)
	if err != nil {
		return nil, err
	}
	functionName, err := assembleFunctionName(inputSign)
	if err != nil {
		return nil, fmt.Errorf("can not assemble function from signature name: %w", err)
	}
	return &Function{
		name:    functionName,
		input:   inputSign,
		wrapper: inputSign,
	}, nil
}

func (f *Function) StringSignature() string {
	return f.input.String()
}

var basicNamedTypesAbbrev = map[string]string{
	// basic
	"bool":    "B",
	"float64": "F",
	"int64":   "I64",
	"int":     "I",
	"string":  "S",
	// named
	"error": "E",
}

func getBasicNamedTypeAbbrev(b Type) (string, error) {
	abbrev, ok := basicNamedTypesAbbrev[b.name]
	if ok == false {
		return "", fmt.Errorf("can not abbreviate basic type %s", b)
	}
	return abbrev, nil
}
func getSliceTypeAbbrev(s Type) (string, error) {
	abbrev, ok := basicNamedTypesAbbrev[s.name]
	if ok == false {
		return "", fmt.Errorf("can not abbreviate slice type %s", s)
	}
	return abbrev + "s", nil
}

func getTypeAbbrev(item Type) (string, error) {
	switch item.kind {
	case TypeBasic, TypeNamed:
		return getBasicNamedTypeAbbrev(item)
	case TypeSlice:
		return getSliceTypeAbbrev(item)
	default:
		return "", fmt.Errorf("can not abbreviate type %s", item)
	}
}

func assembleTupleAbbrev(tuple Tuple) (string, error) {
	abbrevs := make([]string, len(tuple))
	for i, item := range tuple {
		abbrev, err := getTypeAbbrev(item)
		if err != nil {
			return "", fmt.Errorf("can not abbreviate tuple %s: %w", tuple, err)
		}
		abbrevs[i] = abbrev
	}
	return strings.Join(abbrevs, ""), nil
}

func assembleFunctionName(sign *Signature) (string, error) {
	paramsAbbrev, err := assembleTupleAbbrev(sign.params)
	if err != nil {
		return "", fmt.Errorf("can not abbreviate function's %s params: %w", sign, err)
	}
	resultAbbrev, err := assembleTupleAbbrev(sign.results)
	if err != nil {
		return "", fmt.Errorf("can not abbreviate function's %s results: %w", sign, err)
	}
	return fmt.Sprintf("FuncA%sR%s", paramsAbbrev, resultAbbrev), nil
}
