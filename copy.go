package fieldmask_utils

import (
	"fmt"
	"github.com/pkg/errors"
	"reflect"
)

// StructToStruct copies `src` struct to `dst` struct using the given FieldFilter.
// Only the fields where FieldFilter returns true will be copied to `dst`.
// `src` and `dst` must be coherent in terms of the field names, but it is not required for them to be of the same type.
func StructToStruct(filter FieldFilter, src, dst interface{}) error {
	srcVal := indirect(reflect.ValueOf(src))
	srcType := srcVal.Type()
	for i := 0; i < srcVal.NumField(); i++ {
		f := srcVal.Field(i)
		fieldName := srcType.Field(i).Name
		subFilter, ok := filter.Filter(fieldName)
		if !ok {
			// Skip this field.
			continue
		}
		if !f.CanSet() {
			return errors.Errorf("Can't set a value on a field %s", fieldName)
		}

		srcField, err := getField(src, fieldName)
		if err != nil {
			return errors.Wrapf(err, "failed to get the field %s from %T", fieldName, src)
		}
		dstField, err := getField(dst, fieldName)
		if err != nil {
			return errors.Wrapf(err, "failed to get the field %s from %T", fieldName, dst)
		}

		dstFieldType := dstField.Type()

		switch dstFieldType.Kind() {
		case reflect.Interface:
			if srcField.IsNil() {
				dstField.Set(reflect.Zero(dstFieldType))
				continue
			}
			if !srcField.Type().Implements(dstFieldType) {
				return errors.Errorf("src %T does not implement dst %T",
					srcField.Interface(), dstField.Interface())
			}

			v := reflect.New(srcField.Elem().Elem().Type())
			if err := StructToStruct(subFilter, srcField.Interface(), v.Interface()); err != nil {
				return err
			}
			dstField.Set(v)

		case reflect.Ptr:
			if srcField.IsNil() {
				dstField.Set(reflect.Zero(dstFieldType))
				continue
			}
			v := reflect.New(dstFieldType.Elem())
			if err := StructToStruct(subFilter, srcField.Interface(), v.Interface()); err != nil {
				return err
			}
			dstField.Set(v)

		case reflect.Array, reflect.Slice:
			// Check if it is an array of values (non-pointers).
			if dstFieldType.Elem().Kind() != reflect.Ptr {
				// Handle this array/slice as a regular non-nested data structure: copy it entirely to dst.
				dstField.Set(srcField)
				continue
			}
			v := reflect.New(dstFieldType).Elem()
			// Iterate over items of the slice/array.
			for i := 0; i < srcField.Len(); i++ {
				subValue := srcField.Index(i)
				newDst := reflect.New(dstFieldType.Elem().Elem())
				if err := StructToStruct(subFilter, subValue.Interface(), newDst.Interface()); err != nil {
					return err
				}
				v.Set(reflect.Append(v, newDst))
			}
			dstField.Set(v)

		default:
			// For primitive data types just copy them entirely.
			dstField.Set(srcField)
		}
	}
	return nil
}

// StructToMap copies `src` struct to the `dst` map.
// Behavior is similar to `StructToStruct`.
func StructToMap(filter FieldFilter, src interface{}, dst map[string]interface{}) error {
	srcVal := indirect(reflect.ValueOf(src))
	srcType := srcVal.Type()
	for i := 0; i < srcVal.NumField(); i++ {
		f := srcVal.Field(i)
		fieldName := srcType.Field(i).Name
		subFilter, ok := filter.Filter(fieldName)
		if !ok {
			// Skip this field.
			continue
		}
		if !f.CanSet() {
			return errors.Errorf("Can't set a value on a field %s", fieldName)
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
			v := make(map[string]interface{})
			if err := StructToMap(subFilter, srcField.Interface(), v); err != nil {
				return err
			}
			dst[fieldName] = v

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
