package manip

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/pantheon-systems/search-secrets/pkg/dev"
)

const squashFlag = "squash"

type (
	// A param is a struct, and a field that exists on that struct, or on a nested struct.
	//
	// Usage:
	//
	// type (
	//     base    struct{ inner   inner }
	//     inner   struct{ innerer innerer }
	//     innerer struct{ foo     int }
	// )
	// func main() {
	//     a := base{inner{innerer{foo: 1}}}
	//     fmt.Print(NewBasicParam(&a, &a.inner.innerer.foo).PathName()) // prints "inner.innerer.foo"
	//     fmt.Print(NewBasicParam(&a, &a.foo).PathName())               // prints "inner.innerer.foo"
	// }
	Param struct {
		// structElemVal value (call m.structElemVal.Addr() for the pointer)
		structElemVal reflect.Value

		tagName string

		// Call r.structElemVal.FieldByIndex(r.fieldIndex) to get the nested leaf field that was
		// originally passed into the constructor.
		fieldIndex []int

		// DEV These are so I can see the reflect.Values in my debugger.
		pathNameString     string
		structValString    string
		leafFieldValString string
		structFieldName    string
	}
	structFilterFunc func(structElemVal reflect.Value) bool
)

func NewParam(structPtr, leafFieldPtr interface{}, tagName string, structFilter structFilterFunc) (result *Param) {
	structPtrVal := reflect.ValueOf(structPtr)
	leafFieldPtrVal := reflect.ValueOf(leafFieldPtr)
	dev.PrintVal(structPtrVal, "P.======= NewParam structPtrVal")
	dev.PrintVal(leafFieldPtrVal, "P.                 leafFieldPtrVal")

	// structPtrVal
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

	// Leaf field (use r.LeafField() to get to this field)
	if leafFieldPtrVal.Kind() != reflect.Ptr {
		panic("must be a pointer to the leaf field")
	}

	// This creates a singly linked list to the field
	fieldIndex := buildFieldIndex(structElemVal, leafFieldPtrVal, structFilter)
	if fieldIndex == nil {
		fieldIndex = buildFieldIndex(structElemVal, leafFieldPtrVal, structFilter)
		// The leaf field is not actually a leaf field
		panic(fmt.Sprintf("field not found in struct or any of its descendents: %v", leafFieldPtrVal))
	}

	r := &Param{
		structElemVal: structElemVal,
		fieldIndex:    fieldIndex,
		tagName:       tagName,
	}

	// DEV These are so I can see the reflect.Values in my debugger.
	r.pathNameString = r.PathName()
	r.structValString = structElemVal.String()
	r.leafFieldValString = r.LeafField().String()
	r.structFieldName = r.LeafStructField().Name

	return r
}

func NewBasicParam(structPtr, leafFieldPtr interface{}) (result *Param) {
	return NewParam(structPtr, leafFieldPtr, "", nil)
}

func buildFieldIndex(structElemVal, leafFieldPointerVal reflect.Value, structFilter structFilterFunc) (result []int) {

	// Find direct field
	// TODO Test if this can tell the difference between two struct fields pointing to the same object
	dev.PrintVal(structElemVal, "P.bFIndex PARENT")
	dev.PrintVal(leafFieldPointerVal, fmt.Sprintf("P.bFIndex LOOKING FOR THIS POINTER, of type %s", leafFieldPointerVal.Type().String()))
	for i := structElemVal.NumField() - 1; i >= 0; i-- {
		value := structElemVal.Field(i)

		var valuePtrVal = value
		if value.Kind() != reflect.Ptr {
			valuePtrVal = value.Addr()
		}
		dev.PrintVal(valuePtrVal, "P.bFIndex ?")
		if valuePtrVal.Pointer() == leafFieldPointerVal.Pointer() {
			dev.PrintVal(valuePtrVal, "YES to pointer")

			// Do additional type comparison because it's possible that the address of
			// an embedded struct is the same as the first field of the embedded struct
			if valuePtrVal.Type() != leafFieldPointerVal.Type() {
				dev.PrintVal(valuePtrVal, fmt.Sprintf("NO to type (%s == %s)\n", value.Type(), leafFieldPointerVal.Type()))
				continue
			}
			dev.PrintVal(valuePtrVal, fmt.Sprintf("YES to type (%s == %s)\n", value.Type(), leafFieldPointerVal.Type()))

			// We found the field as a direct child, so our field index only has one element.
			return []int{i}
		}
	}

	// If the field is not on this struct, then it's in a descendent struct
	// so find the create the immediate child, which might cause the child to create the grandchild, etc,
	// until we reach the field that was passed into the constructor.
	for i := structElemVal.NumField() - 1; i >= 0; i-- {
		value := structElemVal.Field(i)

		valuePtrVal := value
		if value.Type().Kind() != reflect.Ptr {
			valuePtrVal = value.Addr()
		}
		valueElemVal := valuePtrVal.Elem()

		if valueElemVal.Kind() != reflect.Struct {
			dev.PrintVal(valuePtrVal, "skipped because not a struct")
			continue
		}

		if structFilter != nil && !structFilter(valueElemVal) {
			dev.PrintVal(valuePtrVal, "skipped because didn't match filter")
			continue
		}

		dev.PrintVal(valuePtrVal, "YES to pointer")
		childFieldIndex := buildFieldIndex(valueElemVal, leafFieldPointerVal, structFilter)

		// If it's nil then the leaf field is under another child (or else we'll panic later)
		if childFieldIndex == nil {
			continue
		}

		return append([]int{i}, childFieldIndex...)
	}

	return
}

// Returns a pointer to the field that was originally passed into the constructor
func (p *Param) LeafField() (result reflect.Value) {
	result = p.structElemVal.FieldByIndex(p.fieldIndex)
	if result.Kind() != reflect.Ptr {
		result = result.Addr()
	}
	return
}

func (p *Param) SetLeafFieldValueFromString(value interface{}) (err error) {
	var ok bool
	var stringValue string
	if stringValue, ok = value.(string); !ok {
		panic("currently only strings are supported")
	}

	fieldPtrVal := p.LeafField()
	var convertedValue interface{}

	kind := fieldPtrVal.Elem().Type().Kind()
	switch kind {
	case reflect.Bool:
		convertedValue, _ = strconv.ParseBool(stringValue)
	case reflect.String:
		convertedValue = stringValue
	case reflect.Int:
		convertedValue, err = strconv.Atoi(stringValue)
	default:
		panic(fmt.Sprintf("unsupported kind: %s; value: %v", kind, value))
	}

	if err != nil {
		return
	}

	fieldPtrVal.Elem().Set(reflect.ValueOf(convertedValue))
	return
}

// Returns the value of the field that was passed into the constructor
func (p *Param) LeafStructField() (result reflect.StructField) {
	result = p.structElemVal.Type().FieldByIndex(p.fieldIndex)
	return
}

func (p *Param) LeafFieldValue() (result interface{}) {
	return p.LeafField().Interface()
}

func (p *Param) tagData(structField reflect.StructField) (keyNameTag string, squash bool) {
	if p.tagName == "" {
		return
	}

	var tagParts []string
	if tagValue, ok := structField.Tag.Lookup(p.tagName); ok {
		tagParts = strings.Split(tagValue, ",")
	} else {
		return
	}

	if tagParts[0] != "" && tagParts[0] != "-" {
		keyNameTag = tagParts[0]
	}

	for _, tag := range tagParts[1:] {
		if tag == squashFlag {
			squash = true
			break
		}
	}

	return
}

func (p *Param) PathNamePieces() (result []string) {
	result = []string{}

	structElemVal := p.structElemVal
	for i, x := range p.fieldIndex {
		structField := structElemVal.Type().Field(x)

		keyNameTag, squash := p.tagData(structField)

		keyName := structField.Name
		if keyNameTag != "" {
			keyName = keyNameTag
		} else {
			keyName = structField.Name
		}

		// The leaf doesn't get squashed
		if !squash || i == len(p.fieldIndex)-1 {
			result = append(result, keyName)
		}

		// Next
		structElemVal = structElemVal.Field(x)
		if structElemVal.Type().Kind() == reflect.Ptr {
			structElemVal = structElemVal.Elem()
		}
	}
	return
}

func (p *Param) PathName() (result string) {
	pieces := p.PathNamePieces()
	return strings.Join(pieces, ".")
}
func (p *Param) String() (result string) {
	result = p.PathName()

	value := fmt.Sprintf("%v", p.LeafFieldValue())
	if value != "" {
		result += fmt.Sprintf(" (%s)", value)
	}

	return
}
