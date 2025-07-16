package bind

import (
	"fmt"
	"reflect"
)

type item struct {
	target  any
	valueOf reflect.Value
	typeOf  reflect.Type
}

func (bi *item) Startup() error {
	if match, ok := bi.target.(Startup); ok {
		return match.Startup()
	}
	return nil
}

func (bi *item) Shutdown() error {
	if match, ok := bi.target.(Shutdown); ok {
		return match.Shutdown()
	}
	return nil
}

func buildItem(target any) (*item, error) {
	if target == nil {
		return nil, fmt.Errorf("cannot bind nil value")
	}
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("cannot bind non-pointer type %T, all items must be pointers", target)
	}
	if v.IsNil() {
		return nil, fmt.Errorf("cannot bind nil pointer of type %T", target)
	}

	return &item{target: target, valueOf: v, typeOf: reflect.TypeOf(target)}, nil
}
