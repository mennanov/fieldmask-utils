## Protobuf Field Mask utils for Go

Features:

* Copy from any Go struct to any compatible Go struct with a field mask applied
* Copy from any Go struct to a `map[string]interface{}` with a field mask applied

### Examples

Copy from a protobuf message to a protobuf message:

```
// test.proto

message UpdateUserRequest {
    User user = 1;
    google.protobuf.FieldMask field_mask = 2;
}
```

```
import "github.com/golang/protobuf/protoc-gen-go/generator"

var request UpdateUserRequest
userDst := &testproto.User{} // a struct to copy to
mask, err := fieldmask_utils.ParseFieldMask(request.FieldMask, generator.CamelCase)
// handle err...
fieldmask_utils.StructToStruct(mask, request.User, userDst)
// Only the fields mentioned in the field mask will be copied to userDst, other fields are left intact
```

Copy from a protobuf message to a `map[string]interface{}`:

```
import "github.com/golang/protobuf/protoc-gen-go/generator"

var request UpdateUserRequest
userDst := make(map[string]interface{}) // a map to copy to
mask, err := fieldmask_utils.ParseFieldMask(request.FieldMask, generator.CamelCase)
// handle err...
err := fieldmask_utils.StructToMap(mask, request.User, userDst)
// handle err..
// Only the fields mentioned in the field mask will be copied to userDst, other fields are left intact
```

### Limitations

1.  Larger scope field masks have no effect and are not considered invalid:

    field mask strings `"a", "a.b", "a.b.c"` will result in a mask `a{b{c}}`, which is the same as `"a.b.c"`.

2.  Masks inside a protobuf `Any` and `Map` are not supported.
3.  When copying from a struct to struct the destination struct must have the same fields (or a subset)
    as the source struct. Pointers must also be coherent: if a field is a pointer in the source struct, then
    it also must be a pointer (not a value field) in the destination struct.
