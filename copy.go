package fieldmask_utils

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

// StructToStruct copies `src` struct to `dst` struct using the given FieldFilter.
// Only the fields where FieldFilter returns true will be copied to `dst`.
// `src` and `dst` must be coherent in terms of the field names, but it is not required for them to be of the same type.
func StructToStruct(filter FieldFilter, src, dst interface{}) error {
	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr {
		return errors.Errorf("dst must be a pointer, %s given", dstVal.Kind())
	}
	srcVal := reflect.ValueOf(src).Elem()
	dstVal = dstVal.Elem()
	return copyWithFilter(filter, &srcVal, &dstVal)
}

func copyWithFilter(filter FieldFilter, src, dst *reflect.Value) error {
	if src.Kind() != dst.Kind() {
		return errors.Errorf("src kind %s differs from dst kind %s", src.Kind(), dst.Kind())
	}

	switch src.Kind() {
	case reflect.Struct:
		for i := 0; i < src.NumField(); i++ {
			fieldName := src.Type().Field(i).Name
			srcField := src.FieldByName(fieldName)
			dstField := dst.FieldByName(fieldName)

			subFilter, ok := filter.Filter(fieldName)
			if !ok {
				// Skip this field.
				continue
			}
			if !dstField.CanSet() {
				return errors.Errorf("Can't set a value on a destination field %s", fieldName)
			}
			if err := copyWithFilter(subFilter, &srcField, &dstField); err != nil {
				return err
			}
		}

	case reflect.Ptr:
		if src.IsNil() {
			// If src is nil set dst to nil too.
			dst.Set(reflect.Zero(dst.Type()))
			break
		}
		if dst.IsNil() {
			// If dst is nil create a new instance of the underlying type and set dst to the pointer of that instance.
			dst.Set(reflect.New(dst.Type().Elem()))
		}

		srcElem, dstElem := src.Elem(), dst.Elem()
		if err := copyWithFilter(filter, &srcElem, &dstElem); err != nil {
			return err
		}

	case reflect.Interface:
		if src.IsNil() {
			// If src is nil set dst to nil too.
			dst.Set(reflect.Zero(dst.Type()))
			break
		}
		if !dst.Type().Implements(src.Type()) {
			return errors.Errorf("dst %s does not implement src %s", dst.Type(), src.Type())
		}
		if dst.IsNil() {
			if src.Elem().Kind() != reflect.Ptr {
				// Non-pointer interface implementations are not addressable.
				return errors.Errorf("expected a pointer for an interface value, got %s instead", src.Elem().Kind())
			}
			dst.Set(reflect.New(src.Elem().Elem().Type()))
		}

		srcElem, dstElem := src.Elem(), dst.Elem()
		if err := copyWithFilter(filter, &srcElem, &dstElem); err != nil {
			return err
		}

	case reflect.Slice:
		dstLen := dst.Len()
		for i := 0; i < src.Len(); i++ {
			srcItem := src.Index(i)
			var dstItem reflect.Value
			if i < dstLen {
				// Use an existing item.
				dstItem = dst.Index(i)
			} else {
				// Create a new item if needed.
				dstItem = reflect.New(dst.Type().Elem()).Elem()
			}

			if err := copyWithFilter(filter, &srcItem, &dstItem); err != nil {
				return err
			}

			if i >= dstLen {
				// Append newly created items to the slice.
				dst.Set(reflect.Append(*dst, dstItem))
			}
		}

	case reflect.Array:
		dstLen := dst.Len()
		if dstLen != src.Len() {
			return errors.Errorf("dst array size %d differs from src size %d", dstLen, src.Len())
		}
		for i := 0; i < src.Len(); i++ {
			srcItem := src.Index(i)
			dstItem := dst.Index(i)
			if err := copyWithFilter(filter, &srcItem, &dstItem); err != nil {
				return errors.WithStack(err)
			}
		}

	default:
		if !dst.CanSet() {
			return errors.Errorf("dst %s, %s is not settable", dst, dst.Type())
		}
		dst.Set(*src)
	}

	return nil
}

// StructToMap copies `src` struct to the `dst` map.
// Behavior is similar to `StructToStruct`.
func StructToMap(filter FieldFilter, src interface{}, dst map[string]interface{}) error {
	srcVal := indirect(reflect.ValueOf(src))
	srcType := srcVal.Type()
	for i := 0; i < srcVal.NumField(); i++ {
		fieldName := srcType.Field(i).Name
		subFilter, ok := filter.Filter(fieldName)
		if !ok {
			// Skip this field.
			continue
		}
		srcField, err := getField(src, fieldName)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("failed to get the field %s from %T", fieldName, src))
		}
		switch srcField.Kind() {
		case reflect.Ptr, reflect.Interface:
			if srcField.IsNil() {
				dst[fieldName] = nil
				continue
			}

			var newValue map[string]interface{}
			existingValue, ok := dst[fieldName]
			if ok {
				newValue = existingValue.(map[string]interface{})
			} else {
				newValue = make(map[string]interface{})
			}
			if err := StructToMap(subFilter, srcField.Interface(), newValue); err != nil {
				return err
			}
			dst[fieldName] = newValue

		case reflect.Array, reflect.Slice:
			// Check if it is an array of values (non-pointers).
			if srcField.Type().Elem().Kind() != reflect.Ptr {
				// Handle this array/slice as a regular non-nested data structure: copy it entirely to dst.
				if srcField.Len() > 0 {
					dst[fieldName] = srcField.Interface()
				} else {
					dst[fieldName] = []interface{}(nil)
				}
				continue
			}
			v := make([]map[string]interface{}, 0)
			// Iterate over items of the slice/array.
			for i := 0; i < srcField.Len(); i++ {
				subValue := srcField.Index(i)
				newDst := make(map[string]interface{})
				if err := StructToMap(subFilter, subValue.Interface(), newDst); err != nil {
					return err
				}
				v = append(v, newDst)
			}
			dst[fieldName] = v

		case reflect.Struct:
			var newValue map[string]interface{}
			existingValue, ok := dst[fieldName]
			if ok {
				newValue = existingValue.(map[string]interface{})
			} else {
				newValue = make(map[string]interface{})
			}
			if err := StructToMap(subFilter, srcField.Interface(), newValue); err != nil {
				return err
			}
			dst[fieldName] = newValue

		default:
			// Set a value on a map.
			dst[fieldName] = srcField.Interface()
		}
	}
	return nil
}

func getField(obj interface{}, name string) (reflect.Value, error) {
	objValue := reflectValue(obj)
	field := objValue.FieldByName(name)
	if !field.IsValid() {
		return reflect.ValueOf(nil), errors.Errorf("no such field: %s in obj %T", name, obj)
	}
	return field, nil
}

func reflectValue(obj interface{}) reflect.Value {
	var val reflect.Value

	if reflect.TypeOf(obj).Kind() == reflect.Ptr {
		val = reflect.ValueOf(obj).Elem()
	} else {
		val = reflect.ValueOf(obj)
	}

	return val
}

func indirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}
