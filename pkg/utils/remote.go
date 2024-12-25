package utils

import (
	"context"
	"fmt"
	"reflect"
)

// RemoteCall calls a remote api
// Parameters:
// - fn: remote api.
// - childCtx: the first argument to fn
// - args: the rest arguments
//
// Returns:
// - results: fn return values in an array
// - err: error
func RemoteCall(fn interface{}, childCtx context.Context, args ...interface{}) (results chan []interface{}, err error) {
	results = make(chan []interface{}, 1)

	fnValue := reflect.ValueOf(fn)
	if fnValue.Kind() != reflect.Func {
		panic("wrong fn")
	}

	fnType := fnValue.Type()

	ctx_args := make([]interface{}, 0, len(args)+1)
	ctx_args = append(ctx_args, childCtx)
	ctx_args = append(ctx_args, args...)
	if len(ctx_args) != fnType.NumIn() {
		panic("wrong arg number")
	}
	in := make([]reflect.Value, 0, len(ctx_args))

	for i, arg := range ctx_args {

		if arg == nil || reflect.TypeOf(arg).ConvertibleTo(fnType.In(i)) {
			in = append(in, reflect.ValueOf(arg))
		} else {
			panic(fmt.Sprintf("wrong argument type, expect: %s, got: %s", fnType.In(i), reflect.TypeOf(arg)))
		}
	}

	go func() {
		result := fnValue.Call(in)
		values := make([]interface{}, 0, len(result))
		for _, v := range result {
			values = append(values, v.Interface())
		}
		results <- values
	}()

	return results, nil

}
