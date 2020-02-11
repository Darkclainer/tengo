package main

import (
	"fmt"
	"go/types"
)

type TypeKind int

const (
	TypeUnknow TypeKind = iota
	TypeBasic
	TypeNamed
	TypeSlice
)

type Type struct {
	name string
	kind TypeKind
}

func (t Type) String() string {
	switch t.kind {
	case TypeBasic, TypeNamed:
		return t.name
	case TypeSlice:
		return "[]" + t.name
	default:
		return fmt.Sprintf("Uknown Type.kind: %d", t.kind)
	}
}

func convertType(item types.Type) (Type, error) {
	switch t := item.(type) {
	case *types.Basic:
		return convertBasicType(t)
	case *types.Slice:
		return convertSliceType(t)
	case *types.Named:
		return convertNamedType(t)
	default:
		return Type{}, fmt.Errorf("can not convert type %s", item.String())
	}
}

func convertBasicType(b *types.Basic) (Type, error) {
	return Type{
		name: b.String(),
		kind: TypeBasic,
	}, nil
}
func convertSliceType(s *types.Slice) (Type, error) {
	elem := s.Elem()
	elemBasic, ok := elem.(*types.Basic)
	if ok == false {
		return Type{}, fmt.Errorf("can not convert slice's element type %s to basic type", elem.String())
	}
	return Type{
		name: elemBasic.String(),
		kind: TypeSlice,
	}, nil
}

func convertNamedType(s *types.Named) (Type, error) {
	return Type{
		name: s.String(),
		kind: TypeNamed,
	}, nil
}
