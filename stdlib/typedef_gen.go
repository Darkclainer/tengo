// +build typedef

package stdlib

import (
	"github.com/d5/tengo/v2"
)

func ToCallableFunc(i interface{}) tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		return tengo.UndefinedValue, nil
	}
}
