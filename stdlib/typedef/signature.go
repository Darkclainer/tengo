package main

import (
	"fmt"
	"go/types"
	"strings"
)

type Tuple []Type

func (t Tuple) String() string {
	result := make([]string, len(t))
	for i, item := range t {
		result[i] = item.String()
	}
	return strings.Join(result, ", ")
}

type Signature struct {
	params  Tuple
	results Tuple
}

func NewSignature(sign *types.Signature) (*Signature, error) {
	params, err := convertTuple(sign.Params())
	if err != nil {
		return nil, fmt.Errorf("can not convert params of signature %s: %w", sign, err)
	}
	results, err := convertTuple(sign.Results())
	if err != nil {
		return nil, fmt.Errorf("can not convert results of signature %s: %w", sign, err)
	}
	return &Signature{
		params:  params,
		results: results,
	}, nil

}
func (s *Signature) String() string {
	return fmt.Sprintf("(%s) -> (%s)", s.params, s.results)
}

func convertTuple(tuple *types.Tuple) (Tuple, error) {
	result := make([]Type, tuple.Len())
	for i := 0; i < tuple.Len(); i++ {
		srcType := tuple.At(i).Type()
		dstType, err := convertType(srcType)
		if err != nil {
			return nil, fmt.Errorf("can not convert tuple %s to []Type: %w", tuple.String(), err)
		}
		result[i] = dstType

	}
	return result, nil
}
