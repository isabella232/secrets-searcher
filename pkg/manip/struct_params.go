package manip

import (
	"reflect"
)

type StructParams struct {
	Params []*Param
}

func NewStructParams(structPtr interface{}, tagName string, structFilter structFilterFunc) (result *StructParams) {
	structPtrVal := reflect.ValueOf(structPtr)

	// structElemVal
	if structPtrVal.Kind() != reflect.Ptr {
		panic("must be a pointer (to a struct)")
	}
	if structPtrVal.Elem().Kind() != reflect.Struct {
		panic("must be a pointer to a struct")
	}
	if structPtrVal.IsNil() {
		panic("nil struct")
	}
	structElemVal := structPtrVal.Elem()

	params := buildParams(structElemVal, tagName, structFilter)
	if params == nil {
		panic("no fields found")
	}

	r := &StructParams{
		Params: params,
	}

	return r
}

func buildParams(structElemVal reflect.Value, tagName string, structFilter structFilterFunc) (result []*Param) {
	structPtrVal := structElemVal.Addr()
	fieldPtrs := getDescendantFieldPointers(structPtrVal, structFilter)
	structElemPtr := structElemVal.Addr().Interface()

	for _, fieldPtr := range fieldPtrs {
		param := NewParam(structElemPtr, fieldPtr, tagName, structFilter)
		result = append(result, param)
	}

	return
}

func getDescendantFieldPointers(structPtrVal reflect.Value, structFilter structFilterFunc) (result []interface{}) {
	structElemVal := structPtrVal.Elem()
	for i := structElemVal.NumField() - 1; i >= 0; i-- {
		fieldVal := structElemVal.Field(i)
		if !fieldVal.CanSet() {
			continue
		}

		fieldPtrVal := fieldVal
		if fieldPtrVal.Kind() != reflect.Ptr {
			fieldPtrVal = fieldPtrVal.Addr()
		}
		fieldPtr := fieldPtrVal.Interface()
		fieldElemVal := fieldPtrVal.Elem()

		result = append(result, fieldPtr)

		// The rest of this block deals with struct fields
		if fieldPtrVal.Elem().Type().Kind() != reflect.Struct {
			continue
		}
		fieldVal.IsValid()
		if structFilter != nil && !structFilter(fieldElemVal) {
			continue
		}

		descendantFieldPtrs := getDescendantFieldPointers(fieldPtrVal, structFilter)
		if descendantFieldPtrs != nil {
			result = append(result, descendantFieldPtrs...)
		}
	}

	return
}
