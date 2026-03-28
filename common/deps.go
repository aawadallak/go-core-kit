package common

import (
	"reflect"
	"strings"
)

func IsPointer(v reflect.Value) bool {
	return v.Kind() == reflect.Ptr
}

func IsNilPointer(v reflect.Value) bool {
	return IsPointer(v) && v.IsNil()
}

// MustValidateDependencies validates that all non-nil fields (unless tagged as "non_required") are present.
// It can accept either a struct or a pointer to a struct.
func MustValidateDependencies[T any](input T) T {
	v := reflect.ValueOf(input)
	var structValue reflect.Value

	if IsPointer(v) {
		if IsNilPointer(v) {
			panic("MustValidateDependencies received a nil pointer")
		}
		structValue = v.Elem()
	} else {
		structValue = v
	}

	if structValue.Kind() != reflect.Struct {
		panic("MustValidateDependencies expects a struct or a pointer to a struct")
	}

	var missingFields []string
	structType := structValue.Type()

	for i := range structValue.NumField() {
		fieldValue := structValue.Field(i)
		fieldType := structType.Field(i)

		if tag, ok := fieldType.Tag.Lookup("deps"); ok && tag == "non_required" {
			continue
		}

		switch fieldValue.Kind() { //nolint:exhaustive // only nil-able kinds are relevant
		case reflect.Pointer, reflect.Interface, reflect.Chan, reflect.Slice, reflect.Map:
			if fieldValue.IsNil() {
				missingFields = append(missingFields, fieldType.Name)
			}
		}
	}

	if len(missingFields) > 0 {
		panic("missing required dependencies: " + strings.Join(missingFields, ", "))
	}

	return input
}
