// +build typedef

package stdlib

import (
	"math/rand"

	"github.com/d5/tengo/v2"
)

var randModule = map[string]tengo.Object{
	"int": &tengo.UserFunction{
		Name:  "int",
		Value: ToCallableFunc(rand.Int63),
	},
	"float": &tengo.UserFunction{
		Name:  "float",
		Value: ToCallableFunc(rand.Float64),
	},
	"intn": &tengo.UserFunction{
		Name:  "intn",
		Value: ToCallableFunc(rand.Int63n),
	},
	"exp_float": &tengo.UserFunction{
		Name:  "exp_float",
		Value: ToCallableFunc(rand.ExpFloat64),
	},
	"norm_float": &tengo.UserFunction{
		Name:  "norm_float",
		Value: ToCallableFunc(rand.NormFloat64),
	},
	"perm": &tengo.UserFunction{
		Name:  "perm",
		Value: ToCallableFunc(rand.Perm),
	},
	"seed": &tengo.UserFunction{
		Name:  "seed",
		Value: ToCallableFunc(rand.Seed),
	},
	"read": &tengo.UserFunction{
		Name: "read",
		Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			if len(args) != 1 {
				return nil, tengo.ErrWrongNumArguments
			}
			y1, ok := args[0].(*tengo.Bytes)
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "first",
					Expected: "bytes",
					Found:    args[0].TypeName(),
				}
			}
			res, err := rand.Read(y1.Value)
			if err != nil {
				ret = wrapError(err)
				return
			}
			return &tengo.Int{Value: int64(res)}, nil
		},
	},
	"rand": &tengo.UserFunction{
		Name: "rand",
		Value: func(args ...tengo.Object) (tengo.Object, error) {
			if len(args) != 1 {
				return nil, tengo.ErrWrongNumArguments
			}
			i1, ok := tengo.ToInt64(args[0])
			if !ok {
				return nil, tengo.ErrInvalidArgumentType{
					Name:     "first",
					Expected: "int(compatible)",
					Found:    args[0].TypeName(),
				}
			}
			src := rand.NewSource(i1)
			return randRand(rand.New(src)), nil
		},
	},
}

func randRand(r *rand.Rand) *tengo.ImmutableMap {
	return &tengo.ImmutableMap{
		Value: map[string]tengo.Object{
			"int": &tengo.UserFunction{
				Name:  "int",
				Value: ToCallableFunc(r.Int63),
			},
			"float": &tengo.UserFunction{
				Name:  "float",
				Value: ToCallableFunc(r.Float64),
			},
			"intn": &tengo.UserFunction{
				Name:  "intn",
				Value: ToCallableFunc(r.Int63n),
			},
			"exp_float": &tengo.UserFunction{
				Name:  "exp_float",
				Value: ToCallableFunc(r.ExpFloat64),
			},
			"norm_float": &tengo.UserFunction{
				Name:  "norm_float",
				Value: ToCallableFunc(r.NormFloat64),
			},
			"perm": &tengo.UserFunction{
				Name:  "perm",
				Value: ToCallableFunc(r.Perm),
			},
			"seed": &tengo.UserFunction{
				Name:  "seed",
				Value: ToCallableFunc(r.Seed),
			},
			"read": &tengo.UserFunction{
				Name: "read",
				Value: func(args ...tengo.Object) (
					ret tengo.Object,
					err error,
				) {
					if len(args) != 1 {
						return nil, tengo.ErrWrongNumArguments
					}
					y1, ok := args[0].(*tengo.Bytes)
					if !ok {
						return nil, tengo.ErrInvalidArgumentType{
							Name:     "first",
							Expected: "bytes",
							Found:    args[0].TypeName(),
						}
					}
					res, err := r.Read(y1.Value)
					if err != nil {
						ret = wrapError(err)
						return
					}
					return &tengo.Int{Value: int64(res)}, nil
				},
			},
		},
	}
}
