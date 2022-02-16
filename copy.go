package fieldmask_utils

import (
	"math"
	"reflect"
	"strings"

	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

// StructToStruct copies `src` struct to `dst` struct using the given FieldFilter.
// Only the fields where FieldFilter returns true will be copied to `dst`.
// `src` and `dst` must be coherent in terms of the field names, but it is not required for them to be of the same type.
// Unexported fields are copied only if the corresponding struct filter is empty and `dst` is assignable to `src`.
func StructToStruct(filter FieldFilter, src, dst interface{}, userOpts ...Option) error {
	opts := newDefaultOptions()
	for _, o := range userOpts {
		o(opts)
	}

	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() != reflect.Ptr {
		return errors.Errorf("dst must be a pointer, %s given", dstVal.Kind())
	}
	srcVal := indirect(reflect.ValueOf(src))
	if srcVal.Kind() != reflect.Struct {
		return errors.Errorf("src kind must be a struct, %s given", srcVal.Kind())
	}
	dstVal = indirect(dstVal)
	if dstVal.Kind() != reflect.Struct {
		return errors.Errorf("dst kind must be a struct, %s given", dstVal.Kind())
	}
	return structToStruct(filter, &srcVal, &dstVal, opts)
}

func ensureCompatible(src, dst *reflect.Value) error {
	srcKind := src.Kind()
	if srcKind == reflect.Ptr {
		srcKind = src.Type().Elem().Kind()
	}
	dstKind := dst.Kind()
	if dstKind == reflect.Ptr {
		dstKind = dst.Type().Elem().Kind()
	}
	if srcKind != dstKind {
		return errors.Errorf("src kind %s differs from dst kind %s", srcKind, dstKind)
	}
	return nil
}

func structToStruct(filter FieldFilter, src, dst *reflect.Value, userOptions *options) error {
	if err := ensureCompatible(src, dst); err != nil {
		return err
	}

	switch src.Kind() {
	case reflect.Struct:
		if dst.CanSet() && dst.Type().AssignableTo(src.Type()) && filter.IsEmpty() {
			dst.Set(*src)
			return nil
		}

		if dst.Kind() == reflect.Ptr {
			if dst.IsNil() {
				dst.Set(reflect.New(dst.Type().Elem()))
			}
			v := dst.Elem()
			dst = &v
		}

		for i := 0; i < src.NumField(); i++ {
			srcType := src.Type()
			fieldName := srcType.Field(i).Name
			dstName := dstKey(userOptions.DstTag, srcType.Field(i))

			subFilter, ok := filter.Filter(fieldName)
			if !ok {
				// Skip this field.
				continue
			}

			srcField := src.FieldByName(fieldName)
			if !srcField.CanInterface() {
				continue
			}

			dstField := dst.FieldByName(dstName)
			if !dstField.CanSet() {
				return errors.Errorf("Can't set a value on a destination field %s", dstName)
			}

			if err := structToStruct(subFilter, &srcField, &dstField, userOptions); err != nil {
				return err
			}
		}

	case reflect.Ptr:
		if src.IsNil() {
			// If src is nil set dst to nil too.
			dst.Set(reflect.Zero(dst.Type()))
			break
		}
		if dst.Kind() == reflect.Ptr && dst.IsNil() {
			// If dst is nil create a new instance of the underlying type and set dst to the pointer of that instance.
			dst.Set(reflect.New(dst.Type().Elem()))
		}

		if srcAny, ok := src.Interface().(*anypb.Any); ok {
			dstAny, ok := dst.Interface().(*anypb.Any)
			if !ok {
				return errors.Errorf("dst type is %s, expected: %s ", dst.Type(), "*any.Any")
			}

			srcProto, err := srcAny.UnmarshalNew()
			if err != nil {
				return errors.WithStack(err)
			}
			srcProtoValue := reflect.ValueOf(srcProto)

			if dstAny.GetTypeUrl() == "" {
				dstAny.TypeUrl = srcAny.GetTypeUrl()
			}
			dstProto, err := dstAny.UnmarshalNew()
			if err != nil {
				return errors.WithStack(err)
			}
			dstProtoValue := reflect.ValueOf(dstProto)

			if err := structToStruct(filter, &srcProtoValue, &dstProtoValue, userOptions); err != nil {
				return err
			}

			newDstAny := new(anypb.Any)
			if err := newDstAny.MarshalFrom(dstProtoValue.Interface().(proto.Message)); err != nil {
				return errors.WithStack(err)
			}

			dst.Set(reflect.ValueOf(newDstAny))
			break
		}

		srcElem, dstElem := src.Elem(), *dst
		if dst.Kind() == reflect.Ptr {
			dstElem = dst.Elem()
		}

		if err := structToStruct(filter, &srcElem, &dstElem, userOptions); err != nil {
			return err
		}

	case reflect.Interface:
		if src.IsNil() {
			// If src is nil set dst to nil too.
			dst.Set(reflect.Zero(dst.Type()))
			break
		}
		if dst.IsNil() {
			if src.Elem().Kind() != reflect.Ptr {
				// Non-pointer interface implementations are not addressable.
				return errors.Errorf("expected a pointer for an interface value, got %s instead", src.Elem().Kind())
			}
			dst.Set(reflect.New(src.Elem().Elem().Type()))
		}

		srcElem, dstElem := src.Elem(), dst.Elem()
		if err := structToStruct(filter, &srcElem, &dstElem, userOptions); err != nil {
			return err
		}

	case reflect.Slice:
		dstLen := dst.Len()
		srcLen := src.Len()

		if srcLen > userOptions.MaxCopyListSize {
			srcLen = userOptions.MaxCopyListSize
		}
		for i := 0; i < srcLen; i++ {
			srcItem := src.Index(i)
			var dstItem reflect.Value
			if i < dstLen {
				// Use an existing item.
				dstItem = dst.Index(i)
			} else {
				// Create a new item if needed.
				dstItem = reflect.New(dst.Type().Elem()).Elem()
			}

			if err := structToStruct(filter, &srcItem, &dstItem, userOptions); err != nil {
				return err
			}

			if i >= dstLen {
				// Append newly created items to the slice.
				dst.Set(reflect.Append(*dst, dstItem))
			}
		}
		if dstLen > srcLen {
			dst.SetLen(srcLen)
		}

	case reflect.Array:
		dstLen := dst.Len()
		srcLen := src.Len()
		if dstLen < srcLen {
			return errors.Errorf("dst array size %d is less than src size %d", dstLen, srcLen)
		}
		if srcLen > userOptions.MaxCopyListSize {
			srcLen = userOptions.MaxCopyListSize
		}
		for i := 0; i < srcLen; i++ {
			srcItem := src.Index(i)
			dstItem := dst.Index(i)
			if err := structToStruct(filter, &srcItem, &dstItem, userOptions); err != nil {
				return errors.WithStack(err)
			}
		}

	default:
		if !dst.CanSet() {
			return errors.Errorf("dst %s, %s is not settable", dst, dst.Type())
		}
		if dst.Kind() == reflect.Ptr {
			if !src.CanAddr() {
				return errors.Errorf("src %s, %s is not addressable", src, src.Type())
			}
			dst.Set(src.Addr())
		} else {
			dst.Set(*src)
		}
	}

	return nil
}

// options are used in StructToStruct and StructToMap functions to modify the copying behavior.
type options struct {
	DstTag string

	// Controls the maximum number of elements in the array/slice that can be copied, if set negative number means can copy all element.
	MaxCopyListSize int
}

// Option function modifies the given options.
type Option func(*options)

// WithTag sets the destination field name
func WithTag(s string) Option {
	return func(o *options) {
		o.DstTag = s
	}
}

// WithMaxCopyListSize sets the max size which can be copied.
func WithMaxCopyListSize(size int) Option {
	return func(o *options) {
		if size < 0 {
			o.MaxCopyListSize = math.MaxInt64
		} else {
			o.MaxCopyListSize = size
		}
	}
}

func newDefaultOptions() *options {
	return &options{MaxCopyListSize: math.MaxInt64}
}

func dstKey(tag string, f reflect.StructField) string {
	if tag == "" {
		return f.Name
	}
	lookupResult, ok := f.Tag.Lookup(tag)
	if !ok {
		return f.Name
	}
	firstComma := strings.Index(lookupResult, ",")
	if firstComma == -1 {
		return lookupResult
	}
	return lookupResult[:firstComma]
}

// StructToMap copies `src` struct to the `dst` map.
// Behavior is similar to `StructToStruct`.
// Arrays in the non-empty dst are converted to slices.
func StructToMap(filter FieldFilter, src interface{}, dst map[string]interface{}, userOpts ...Option) error {
	opts := newDefaultOptions()
	for _, o := range userOpts {
		o(opts)
	}
	return structToMap(filter, src, dst, opts)
}

func structToMap(filter FieldFilter, src interface{}, dst map[string]interface{}, userOptions *options) error {
	srcVal := indirect(reflect.ValueOf(src))
	srcType := srcVal.Type()
	for i := 0; i < srcVal.NumField(); i++ {
		fieldName := srcType.Field(i).Name
		subFilter, ok := filter.Filter(fieldName)
		if !ok {
			// Skip this field.
			continue
		}
		srcField := srcVal.FieldByName(fieldName)
		if !srcField.CanInterface() {
			continue
		}

		dstName := dstKey(userOptions.DstTag, srcType.Field(i))

		switch srcField.Kind() {
		case reflect.Ptr, reflect.Interface:
			if srcField.IsNil() {
				dst[dstName] = nil
				continue
			}

			var newValue map[string]interface{}
			existingValue, ok := dst[dstName]
			if ok {
				newValue = existingValue.(map[string]interface{})
			} else {
				newValue = make(map[string]interface{})
			}
			if err := structToMap(subFilter, srcField.Interface(), newValue, userOptions); err != nil {
				return err
			}
			dst[dstName] = newValue

		case reflect.Array, reflect.Slice:
			// Check if it is a slice of primitive values.
			itemKind := srcField.Type().Elem().Kind()
			if itemKind != reflect.Ptr && itemKind != reflect.Struct && itemKind != reflect.Interface {
				// Handle this array/slice as a regular non-nested data structure: copy it entirely to dst.
				if srcField.Len() < userOptions.MaxCopyListSize {
					dst[dstName] = srcField.Interface()
				} else {
					// copy MaxCopyListSize items to dst
					dst[dstName] = srcField.Slice(0, userOptions.MaxCopyListSize).Interface()
				}
				continue
			}
			srcLen := srcField.Len()
			if srcLen > userOptions.MaxCopyListSize {
				srcLen = userOptions.MaxCopyListSize
			}
			var newValue []map[string]interface{}
			existingValue, ok := dst[dstName]
			if ok {
				v := reflect.ValueOf(existingValue)
				if v.Kind() == reflect.Array {
					// Convert the array to a slice.
					for i := 0; i < v.Len(); i++ {
						itemInterface := v.Index(i).Interface()
						item, k := itemInterface.(map[string]interface{})
						if !k {
							return errors.Errorf("unexpected dst type %T, expected map[string]interface{}", itemInterface)
						}
						newValue = append(newValue, item)
					}
				} else {
					newValue, ok = existingValue.([]map[string]interface{})
					if !ok {
						return errors.Errorf("unexpected dst type %T, expected []map[string]interface{}", newValue)
					}
				}
			}

			// Iterate over items of the slice/array.
			dstLen := len(newValue)
			if dstLen < srcLen {
				// Grow the dst slice to match the src len.
				for i := 0; i < srcLen-dstLen; i++ {
					newValue = append(newValue, make(map[string]interface{}))
				}
				dstLen = srcLen
			}
			for i := 0; i < srcLen; i++ {
				subValue := srcField.Index(i)
				if err := structToMap(subFilter, subValue.Interface(), newValue[i], userOptions); err != nil {
					return err
				}
			}
			// Truncate the dst to the length of src.
			newValue = newValue[:srcLen]
			dst[dstName] = newValue

		case reflect.Struct:
			var newValue map[string]interface{}
			existingValue, ok := dst[dstName]
			if ok {
				newValue = existingValue.(map[string]interface{})
			} else {
				newValue = make(map[string]interface{})
			}
			if err := structToMap(subFilter, srcField.Interface(), newValue, userOptions); err != nil {
				return err
			}
			dst[dstName] = newValue

		default:
			// Set a value on a map.
			dst[dstName] = srcField.Interface()
		}
	}
	return nil
}

func indirect(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}
