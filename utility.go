package nestor

import (
	"fmt"
	"reflect"
)

func execute(jobFun interface{}, params ...interface{}) (result []reflect.Value, err error) {
	typ := reflect.TypeOf(jobFun)
	if typ.Kind() != reflect.Func {
		return nil, fmt.Errorf("expected function. got %v", typ)
	}
	f := reflect.ValueOf(jobFun)
	if len(params) != f.Type().NumIn() {
		err = fmt.Errorf("expectecd %d parameters. got %d", f.Type().NumIn(), len(params))
		return
	}
	in := make([]reflect.Value, len(params))
	for k, param := range params {
		if param == nil {
			in[k] = reflect.Zero(typ.In(k))
			continue
		}
		in[k] = reflect.ValueOf(param)
	}
	result = f.Call(in)
	return
}
