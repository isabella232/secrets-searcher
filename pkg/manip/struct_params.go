package manip

import (
	"reflect"

	"github.com/pantheon-systems/search-secrets/pkg/dev"
)

type StructParams struct {
	Params []*Param
}

func NewStructParams(structPtr interface{}, tagName string, structFilter structFilterFunc) (result *StructParams) {
	structPtrVal := reflect.ValueOf(structPtr)
	dev.PrintVal(structPtrVal, "P.======= NewSParams structPtrVal")

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
	dev.PrintVal(structPtrVal, "SP.New calling .Elem()")
	structElemVal := structPtrVal.Elem()
	dev.PrintVal(structElemVal, "SP.New called. Now calling P.bParams()")

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
		dev.PrintVal(reflect.ValueOf(fieldPtr), "SP.buildParams fieldPtr from getDesc")
		param := NewParam(structElemPtr, fieldPtr, tagName, structFilter)
		result = append(result, param)
	}

	return
}

func getDescendantFieldPointers(structPtrVal reflect.Value, structFilter structFilterFunc) (result []interface{}) {
	dev.PrintVal(structPtrVal, "STARTING SP.getDesc loop")
	structElemVal := structPtrVal.Elem()
	for i := structElemVal.NumField() - 1; i >= 0; i-- {
		fieldVal := structElemVal.Field(i)
		dev.PrintVal(fieldVal, "SP.getDesc fieldVal in loop")
		if !fieldVal.CanSet() {
			dev.PrintVal(fieldVal, "cannot set")
			continue
		}

		fieldPtrVal := fieldVal
		if fieldPtrVal.Kind() != reflect.Ptr {
			fieldPtrVal = fieldPtrVal.Addr()
			//dev.PrintVal(fieldPtrVal, "SP.getDesc fieldVal in loop - converted to ptr")
		}
		fieldPtr := fieldPtrVal.Interface()
		fieldElemVal := fieldPtrVal.Elem()
		//dev.PrintVal(fieldPtr, "SP.getDesc fieldVal in loop - converted to interface")

		result = append(result, fieldPtr)

		// The rest of this block deals with struct fields
		if fieldPtrVal.Elem().Type().Kind() != reflect.Struct {
			dev.PrintVal(fieldVal, "SP.getDesc fieldVal in loop - not a struct")
			continue
		}
		fieldVal.IsValid()
		if structFilter != nil && !structFilter(fieldElemVal) {
			dev.PrintVal(fieldVal, "SP.getDesc fieldVal in loop - kicked out by filter")
			continue
		}

		dev.PrintVal(fieldPtrVal, "SP.getDesc child, recursing")
		descendantFieldPtrs := getDescendantFieldPointers(fieldPtrVal, structFilter)
		if descendantFieldPtrs != nil {
			result = append(result, descendantFieldPtrs...)
		}
	}

	return
}
