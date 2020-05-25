package fieldmask_utils_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	fieldmask_utils "github.com/mennanov/fieldmask-utils"
)

func TestStructToStruct_SimpleStruct(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	src := &A{
		Field1: "src field1",
		Field2: 1,
	}
	dst := new(A)
	mask := fieldmask_utils.MaskFromString("Field1")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &A{
		Field1: "src field1",
		Field2: 0,
	}, dst)
}

func TestStructToStruct_PtrToInt(t *testing.T) {
	type A struct {
		Field2 *int
	}
	n := 42
	src := &A{
		Field2: &n,
	}
	dst := new(A)

	mask := fieldmask_utils.MaskFromString("Field2")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &A{
		Field2: src.Field2,
	}, dst)
}

func TestStructToStruct_PtrToStruct_EmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	type B struct {
		Field1 string
		Field2 int
		A      *A
	}
	type C struct {
		Field1 string
		B      *B
	}
	src := &C{
		Field1: "C field1",
		B: &B{
			Field1: "StringerB field1",
			Field2: 1,
			A: &A{
				Field1: "StringerA field1",
				Field2: 5,
			},
		},
	}
	dst := new(C)

	mask := fieldmask_utils.MaskFromString("B{Field1,A{Field2}}")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &C{
		Field1: "",
		B: &B{
			Field1: src.B.Field1,
			Field2: 0,
			A: &A{
				Field1: "",
				Field2: src.B.A.Field2,
			},
		},
	}, dst)
}

func TestStructToStruct_PtrToStruct_NonEmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	type B struct {
		Field1 string
		Field2 int
		A      *A
	}
	type C struct {
		Field1 string
		B      *B
	}
	src := &C{
		Field1: "src C field1",
		B: &B{
			Field1: "StringerB field1",
			Field2: 1,
			A: &A{
				Field1: "StringerA field1",
				Field2: 5,
			},
		},
	}
	dst := &C{
		Field1: "dst C field1",
		B: &B{
			Field1: "dst StringerB field1",
			Field2: 2,
			A: &A{
				Field1: "dst StringerA field1",
				Field2: 10,
			},
		},
	}

	mask := fieldmask_utils.MaskFromString("B{Field1,A{Field2}}")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &C{
		Field1: "dst C field1",
		B: &B{
			Field1: src.B.Field1,
			Field2: 2,
			A: &A{
				Field1: "dst StringerA field1",
				Field2: src.B.A.Field2,
			},
		},
	}, dst)
}

func TestStructToStruct_NestedStruct_EmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	type B struct {
		Field1 string
		Field2 int
		A      A
	}
	type C struct {
		Field1 string
		B      B
	}
	src := &C{
		Field1: "C field1",
		B: B{
			Field1: "StringerB field1",
			Field2: 1,
			A: A{
				Field1: "StringerA field1",
				Field2: 5,
			},
		},
	}
	dst := new(C)

	mask := fieldmask_utils.MaskFromString("B{Field1,A{Field2}}")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &C{
		Field1: "",
		B: B{
			Field1: src.B.Field1,
			Field2: 0,
			A: A{
				Field1: "",
				Field2: src.B.A.Field2,
			},
		},
	}, dst)
}

func TestStructToStruct_NestedStruct_EmptyDst_OptionDst(t *testing.T) {
	opts := fieldmask_utils.WithTag("db")
	type ASrc struct {
		Field1 string
		Field2 int `db:"SomeField"`
	}
	type BSrc struct {
		Field1 string `struct:"a_name"`
		A      ASrc   `db:"AnotherName"`
	}
	src := &BSrc{
		Field1: "B Field1",
		A: ASrc{
			Field1: "A Field 1",
			Field2: 1,
		},
	}

	type ADst struct {
		Field1    string
		SomeField int
	}
	type BDst struct {
		Field1      string
		AnotherName ADst
	}
	dst := &BDst{}

	mask := fieldmask_utils.MaskFromString("Field1,A{Field2}")
	err := fieldmask_utils.StructToStruct(mask, src, dst, opts)
	require.NoError(t, err)
	assert.Equal(t, &BDst{
		Field1: src.Field1,
		AnotherName: ADst{
			SomeField: src.A.Field2,
		},
	}, dst)
}

func TestStructToStruct_NestedStruct_NonEmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	type B struct {
		Field1 string
		Field2 int
		A      A
	}
	type C struct {
		Field1 string
		B      B
	}
	src := &C{
		Field1: "src C field1",
		B: B{
			Field1: "src StringerB field1",
			Field2: 1,
			A: A{
				Field1: "src StringerA field1",
				Field2: 5,
			},
		},
	}
	dst := &C{
		Field1: "dst C field1",
		B: B{
			Field1: "dst StringerB field1",
			Field2: 2,
			A: A{
				Field1: "dst StringerA field1",
				Field2: 10,
			},
		},
	}

	mask := fieldmask_utils.MaskFromString("B{Field1,A{Field2}}")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &C{
		Field1: "dst C field1",
		B: B{
			Field1: src.B.Field1,
			Field2: 2,
			A: A{
				Field1: "dst StringerA field1",
				Field2: src.B.A.Field2,
			},
		},
	}, dst)
}

func TestStructToStruct_SliceOfStructs_EmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	type B struct {
		Field1 string
		A      []A
	}
	src := &B{
		Field1: "src StringerB field1",
		A: []A{
			{
				Field1: "StringerA field1 0",
				Field2: 1,
			},
			{
				Field1: "StringerA field1 1",
				Field2: 2,
			},
		},
	}
	dst := new(B)

	mask := fieldmask_utils.MaskFromString("Field1,A{Field2}")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &B{
		Field1: src.Field1,
		A: []A{
			{
				Field1: "",
				Field2: src.A[0].Field2,
			},
			{
				Field1: "",
				Field2: src.A[1].Field2,
			},
		},
	}, dst)
}

func TestStructToStruct_SliceOfStructs_NonEmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	type B struct {
		Field1 string
		A      []A
	}
	src := &B{
		Field1: "src StringerB field1",
		A: []A{
			{
				Field1: "StringerA field1 0",
				Field2: 1,
			},
			{
				Field1: "StringerA field1 1",
				Field2: 2,
			},
			{
				Field1: "StringerA field1 2",
				Field2: 3,
			},
		},
	}
	dst := &B{
		Field1: "dst StringerB field1",
		A: []A{
			{
				Field1: "dst StringerA field1 0",
				Field2: 10,
			},
			{
				Field1: "dst StringerA field1 1",
				Field2: 20,
			},
		},
	}

	mask := fieldmask_utils.MaskFromString("Field1,A{Field2}")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &B{
		Field1: src.Field1,
		A: []A{
			{
				Field1: "dst StringerA field1 0",
				Field2: src.A[0].Field2,
			},
			{
				Field1: "dst StringerA field1 1",
				Field2: src.A[1].Field2,
			},
			{
				Field1: "",
				Field2: src.A[2].Field2,
			},
		},
	}, dst)
}

func TestStructToStruct_EntireSlice_NonEmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	type B struct {
		Field1 string
		A      []A
	}
	src := &B{
		Field1: "src StringerB field1",
		A: []A{
			{
				Field1: "StringerA field1 0",
				Field2: 1,
			},
			{
				Field1: "StringerA field1 1",
				Field2: 2,
			},
		},
	}
	dst := &B{
		Field1: "dst StringerB field1",
		A: []A{
			{
				Field1: "dst StringerA field1 0",
				Field2: 10,
			},
			{
				Field1: "dst StringerA field1 1",
				Field2: 20,
			},
			{
				Field1: "dst StringerA field1 2",
				Field2: 30,
			},
		},
	}

	mask := fieldmask_utils.MaskFromString("Field1,A")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &B{
		Field1: src.Field1,
		A:      src.A,
	}, dst)
}

func TestStructToStruct_SliceOfPtrsToStruct_EmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	type B struct {
		Field1 string
		A      []*A
	}
	src := &B{
		Field1: "src StringerB field1",
		A: []*A{
			{
				Field1: "StringerA field1 0",
				Field2: 1,
			},
			{
				Field1: "StringerA field1 1",
				Field2: 2,
			},
		},
	}
	dst := new(B)

	mask := fieldmask_utils.MaskFromString("Field1,A{Field2}")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &B{
		Field1: src.Field1,
		A: []*A{
			{
				Field1: "",
				Field2: src.A[0].Field2,
			},
			{
				Field1: "",
				Field2: src.A[1].Field2,
			},
		},
	}, dst)
}

func TestStructToStruct_ArrayOfStructs_EmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	type B struct {
		Field1 string
		A      [2]A
	}
	type C struct {
		Field1 string
		A      [3]A
	}
	src := &B{
		Field1: "src StringerB field1",
		A: [2]A{
			{
				Field1: "StringerA field1 0",
				Field2: 1,
			},
			{
				Field1: "StringerA field1 1",
				Field2: 2,
			},
		},
	}
	dst := new(C)

	mask := fieldmask_utils.MaskFromString("Field1,A{Field2}")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &C{
		Field1: src.Field1,
		A: [3]A{
			{
				Field1: "",
				Field2: src.A[0].Field2,
			},
			{
				Field1: "",
				Field2: src.A[1].Field2,
			},
			{
				Field1: "",
				Field2: 0,
			},
		},
	}, dst)
}

func TestStructToStruct_Array_DstLenLessThanSrc(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	type B struct {
		Field1 string
		A      [2]A
	}
	type C struct {
		Field1 string
		A      [1]A
	}
	src := &B{
		Field1: "src StringerB field1",
		A: [2]A{
			{
				Field1: "StringerA field1 0",
				Field2: 1,
			},
			{
				Field1: "StringerA field1 1",
				Field2: 2,
			},
		},
	}
	dst := new(C)

	mask := fieldmask_utils.MaskFromString("Field1,A{Field2}")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	assert.Error(t, err)
}

func TestStructToStruct_DifferentStructTypes(t *testing.T) {
	type A struct {
		Field string
	}

	type B struct {
		Field string
	}

	src := &A{"value"}
	dst := new(B)
	mask := fieldmask_utils.MaskFromString("Field")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &B{src.Field}, dst)
}

func TestStructToStruct_DifferentStructTypesNested(t *testing.T) {
	type A struct {
		Field string
	}

	type AA struct {
		Field string
	}

	type B struct {
		A A
	}

	type C struct {
		A AA
	}

	src := &B{
		A: A{
			Field: "value",
		},
	}
	dst := new(C)
	mask := fieldmask_utils.MaskFromString("A")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &C{
		A: AA{
			Field: src.A.Field,
		},
	}, dst)
}

func TestStructToStruct_DifferentStructTypesPtrNested(t *testing.T) {
	type A struct {
		Field string
	}

	type AA struct {
		Field string
	}

	type B struct {
		A *A
	}

	type C struct {
		A *AA
	}

	src := &B{
		A: &A{
			Field: "value",
		},
	}
	dst := new(C)
	mask := fieldmask_utils.MaskFromString("A")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &C{
		A: &AA{
			Field: src.A.Field,
		},
	}, dst)
}

type StringerA struct {
	Field string
}

func (a *StringerA) String() string {
	return a.Field
}

type StringerB struct {
	Field string
}

func (b *StringerB) String() string {
	return b.Field
}

func TestStructToStruct_Interface_EmptyDst(t *testing.T) {
	type C struct {
		S fmt.Stringer
	}

	src := &C{
		S: &StringerA{
			Field: "StringerA",
		},
	}
	dst := new(C)
	mask := fieldmask_utils.MaskFromString("S")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &C{
		S: &StringerA{
			Field: "StringerA",
		},
	}, dst)
}

func TestStructToStruct_SameInterfaces_NonEmptyDst(t *testing.T) {
	type C struct {
		S fmt.Stringer
	}

	src := &C{
		S: &StringerA{
			Field: "StringerA",
		},
	}
	dst := &C{
		S: &StringerA{
			Field: "StringerB",
		},
	}
	mask := fieldmask_utils.MaskFromString("S")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, src.S.String(), dst.S.String())
	assert.Equal(t, &C{
		S: &StringerA{
			Field: "StringerA",
		},
	}, dst)
}

func TestStructToStruct_DifferentCompatibleInterfaces_NonEmptyDst(t *testing.T) {
	type C struct {
		S fmt.Stringer
	}

	src := &C{
		S: &StringerA{
			Field: "StringerA",
		},
	}
	dst := &C{
		S: &StringerB{
			Field: "StringerB",
		},
	}
	mask := fieldmask_utils.MaskFromString("S")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, src.S.String(), dst.S.String())
}

type Logger interface {
	Log() string
}

type LoggerImpl struct {
	Field string
}

func (d *LoggerImpl) Log() string {
	return d.Field
}

func TestStructToStruct_DifferentIncompatibleInterfaces(t *testing.T) {
	type C struct {
		S fmt.Stringer
	}

	type E struct {
		S Logger
	}

	src := &C{
		S: &StringerA{
			Field: "StringerA",
		},
	}
	dst := &E{
		S: &LoggerImpl{
			Field: "Logger",
		},
	}
	mask := fieldmask_utils.MaskFromString("S")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, src.S.String(), dst.S.Log())
}

func TestStructToStruct_EmptyMask(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	src := &A{
		Field1: "A Field1",
		Field2: 1,
	}
	dst := new(A)
	mask := fieldmask_utils.MaskFromString("")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, src, dst)
}

type StringerImpl struct {
	Name string
}

func (*StringerImpl) someMethod() {}
func (f *StringerImpl) String() string {
	return f.Name
}

type StringerNonPtrImpl struct {
	Name string
}

func (s StringerNonPtrImpl) String() string {
	return s.Name
}

func TestStructToStruct_SameInterfacesPtr_EmptyDst(t *testing.T) {
	type A struct {
		Stringer fmt.Stringer
	}

	type B struct {
		Stringer fmt.Stringer
	}

	src := &A{
		Stringer: &StringerImpl{Name: "Jessica"},
	}

	dst := new(B)

	mask := fieldmask_utils.MaskFromString("Stringer")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	assert.NoError(t, err)
	assert.Equal(t, src.Stringer.String(), dst.Stringer.String())
}

func TestStructToStruct_SameInterfacesPtr_NonEmptyDst(t *testing.T) {
	type A struct {
		Stringer fmt.Stringer
	}

	type B struct {
		Stringer fmt.Stringer
	}

	src := &A{
		Stringer: &StringerImpl{
			Name: "Jessica",
		},
	}
	dst := &B{
		Stringer: &StringerImpl{
			Name: "Dana",
		},
	}

	mask := fieldmask_utils.MaskFromString("Stringer")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	assert.NoError(t, err)
	assert.Equal(t, dst.Stringer.String(), src.Stringer.String())
}

func TestStructToStruct_SameInterfacesNonPtr_EmptyDst(t *testing.T) {
	type A struct {
		Stringer fmt.Stringer
	}
	src := &A{
		Stringer: StringerNonPtrImpl{Name: "Jessica"},
	}

	dst := new(A)

	mask := fieldmask_utils.MaskFromString("Stringer")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	assert.Error(t, err)
}

func TestStructToStruct_SameInterfacesNonPtr_NonEmptyDst(t *testing.T) {
	type A struct {
		Stringer fmt.Stringer
	}
	src := &A{
		Stringer: StringerNonPtrImpl{Name: "Jessica"},
	}

	dst := &A{
		Stringer: StringerNonPtrImpl{Name: "Adam"},
	}

	mask := fieldmask_utils.MaskFromString("Stringer")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	assert.Error(t, err)
}

func TestStructToStruct_NonPtrDst(t *testing.T) {
	type A struct {
		Field int
	}
	src := &A{Field: 1}
	dst := A{}
	mask := fieldmask_utils.MaskFromString("")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	assert.Error(t, err)
}

func TestStructToStruct_DifferentDstKind(t *testing.T) {
	type A struct {
		Field int
	}
	src := &A{Field: 1}
	dst := &map[string]interface{}{}
	mask := fieldmask_utils.MaskFromString("")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	assert.Error(t, err)
}

func TestStructToStruct_UnexportedFieldsPtr(t *testing.T) {
	type A struct {
		foo string
		Bar string
	}

	type B struct {
		A *A
		B string
	}

	src := &B{
		A: &A{
			foo: "foo",
			Bar: "Bar",
		},
		B: "B",
	}
	dst := &B{}

	mask := fieldmask_utils.MaskFromString("A,B")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	assert.NoError(t, err)
	assert.Equal(t, src, dst)
}

func TestStructToStruct_UnexportedFields(t *testing.T) {
	type A struct {
		foo string
		Bar string
	}

	type B struct {
		A A
		B string
	}

	src := &B{
		A: A{
			foo: "foo",
			Bar: "Bar",
		},
		B: "B",
	}
	dst := &B{}

	mask := fieldmask_utils.MaskFromString("A,B")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	assert.NoError(t, err)
	assert.Equal(t, src, dst)
}

func TestStructToStruct_MaskWithInverseMask(t *testing.T) {
	type A struct {
		Foo string
		Bar string
	}

	type B struct {
		A A
		B string
		C string
	}
	src := &B{
		A: A{
			Foo: "foo",
			Bar: "Bar",
		},
		B: "B",
		C: "C",
	}
	for _, mask := range []fieldmask_utils.FieldFilter{
		fieldmask_utils.Mask{"B": nil, "A": &fieldmask_utils.MaskInverse{"Bar": nil}},
		fieldmask_utils.Mask{"B": fieldmask_utils.Mask{}, "A": &fieldmask_utils.MaskInverse{"Bar": fieldmask_utils.Mask{}}},
	} {
		dst := &B{}
		err := fieldmask_utils.StructToStruct(mask, src, dst)
		assert.NoError(t, err)
		assert.Equal(t, &B{
			A: A{
				Foo: src.A.Foo,
			},
			B: "B",
		}, dst)
	}

}

func TestStructToStruct_InverseMaskWithMask(t *testing.T) {
	type A struct {
		Foo string
		Bar string
	}

	type B struct {
		A A
		B string
		C string
	}
	src := &B{
		A: A{
			Foo: "foo",
			Bar: "Bar",
		},
		B: "B",
		C: "C",
	}
	for _, mask := range []fieldmask_utils.FieldFilter{
		fieldmask_utils.MaskInverse{"B": fieldmask_utils.Mask{}, "A": &fieldmask_utils.Mask{"Bar": fieldmask_utils.Mask{}}},
		fieldmask_utils.MaskInverse{"B": nil, "A": &fieldmask_utils.Mask{"Bar": nil}},
	} {
		dst := &B{}
		err := fieldmask_utils.StructToStruct(mask, src, dst)
		assert.NoError(t, err)
		assert.Equal(t, &B{
			A: A{
				Bar: src.A.Bar,
			},
			C: "C",
		}, dst)
	}

}

func TestStructToMap_NestedStruct_EmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	type B struct {
		Field1 string
		A      A
	}
	src := &B{
		Field1: "B Field1",
		A: A{
			Field1: "A Field 1",
			Field2: 1,
		},
	}
	dst := make(map[string]interface{})
	mask := fieldmask_utils.MaskFromString("Field1,A{Field2}")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": src.Field1,
		"A": map[string]interface{}{
			"Field2": src.A.Field2,
		},
	}, dst)
}

func TestStructToMap_NestedStruct_EmptyDst_OptionDst(t *testing.T) {
	opts := fieldmask_utils.WithTag("db")
	type A struct {
		Field1 string
		Field2 int `db:"some_field"`
	}
	type B struct {
		Field1 string `struct:"a_name"`
		A      A      `db:"another_name"`
	}
	src := &B{
		Field1: "B Field1",
		A: A{
			Field1: "A Field 1",
			Field2: 1,
		},
	}
	dst := make(map[string]interface{})
	mask := fieldmask_utils.MaskFromString("Field1,A{Field2}")
	err := fieldmask_utils.StructToMap(mask, src, dst, opts)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": src.Field1,
		"another_name": map[string]interface{}{
			"some_field": src.A.Field2,
		},
	}, dst)
}

func TestStructToMap_NestedStruct_NonEmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	type B struct {
		Field1 string
		A      A
	}
	src := &B{
		Field1: "B Field1",
		A: A{
			Field1: "A Field 1",
			Field2: 1,
		},
	}
	dst := map[string]interface{}{
		"A": map[string]interface{}{
			"Field1": "existing value",
			"Field2": 10,
		},
	}
	mask := fieldmask_utils.MaskFromString("Field1,A{Field2}")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": src.Field1,
		"A": map[string]interface{}{
			"Field1": "existing value",
			"Field2": src.A.Field2,
		},
	}, dst)
}

func TestStructToMap_PtrToStruct_EmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	type B struct {
		Field1 string
		A      *A
	}
	src := &B{
		Field1: "B Field1",
		A: &A{
			Field1: "A Field 1",
			Field2: 1,
		},
	}
	dst := make(map[string]interface{})
	mask := fieldmask_utils.MaskFromString("Field1,A{Field2}")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": src.Field1,
		"A": map[string]interface{}{
			"Field2": src.A.Field2,
		},
	}, dst)
}

func TestStructToMap_PtrToStruct_NonEmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int
	}
	type B struct {
		Field1 string
		A      *A
	}
	src := &B{
		Field1: "B Field1",
		A: &A{
			Field1: "A Field 1",
			Field2: 1,
		},
	}
	dst := map[string]interface{}{
		"A": map[string]interface{}{
			"Field1": "existing value",
			"Field2": 10,
		},
	}
	mask := fieldmask_utils.MaskFromString("Field1,A{Field2}")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": src.Field1,
		"A": map[string]interface{}{
			"Field1": "existing value",
			"Field2": src.A.Field2,
		},
	}, dst)
}

func TestStructToMap_ArrayOfStructs_EmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 string
	}
	type B struct {
		A [1]A
	}
	src := &B{
		A: [1]A{
			{
				Field1: "src field1",
				Field2: "src field2",
			},
		},
	}
	dst := make(map[string]interface{})
	mask := fieldmask_utils.MaskFromString("A{Field2}")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"A": []map[string]interface{}{
			{
				"Field2": src.A[0].Field2,
			},
		},
	}, dst)
}

func TestStructToMap_SliceOfStructs_NonEmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 string
	}
	type B struct {
		A []A
	}
	src := &B{
		A: []A{
			{
				Field1: "src field1",
				Field2: "src field2",
			},
		},
	}
	dst := map[string]interface{}{
		"A": []map[string]interface{}{
			{
				"Field1": "dst field1",
			},
		},
	}
	mask := fieldmask_utils.MaskFromString("A{Field2}")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"A": []map[string]interface{}{
			{
				"Field1": "dst field1",
				"Field2": src.A[0].Field2,
			},
		},
	}, dst)
}

func TestStructToMap_EntireSlicePrimitive_NonEmptyDst(t *testing.T) {
	type A struct {
		Field1 []int
	}
	src := &A{
		Field1: []int{1, 2, 4, 8},
	}

	dst := map[string]interface{}{
		"Field1": []int{16, 32, 64},
	}
	mask := fieldmask_utils.MaskFromString("Field1")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": src.Field1,
	}, dst)
}

func TestStructToMap_EntireSlice_NonEmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 string
	}
	type B struct {
		A []A
	}
	src := &B{
		A: []A{
			{
				Field1: "src ele1 field1",
				Field2: "src ele1 field2",
			},
			{
				Field1: "src ele2 field1",
				Field2: "src ele2 field2",
			},
		},
	}
	dst := map[string]interface{}{
		"A": []map[string]interface{}{
			{
				"Field1": "dst ele1 field1",
			},
			{
				"Field2": "dst ele2 field 2",
			},
			{
				"Field1": "dst ele3 field 3",
			},
		},
	}
	mask := fieldmask_utils.MaskFromString("A")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"A": []map[string]interface{}{
			{
				"Field1": src.A[0].Field1,
				"Field2": src.A[0].Field2,
			},
			{
				"Field1": src.A[1].Field1,
				"Field2": src.A[1].Field2,
			},
		},
	}, dst)
}

func TestStructToMap_EntireSlice_DstSliceLenIsLessThanSource(t *testing.T) {
	type A struct {
		Field1 string
		Field2 string
	}
	type B struct {
		A []A
	}
	src := &B{
		A: []A{
			{
				Field1: "src ele1 field1",
				Field2: "src ele1 field2",
			},
			{
				Field1: "src ele2 field1",
				Field2: "src ele2 field2",
			},
		},
	}
	dst := map[string]interface{}{
		"A": []map[string]interface{}{
			{
				"Field1": "dst ele1 field1",
			},
		},
	}
	mask := fieldmask_utils.MaskFromString("A")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"A": []map[string]interface{}{
			{
				"Field1": "src ele1 field1",
				"Field2": "src ele1 field2",
			},
			{
				"Field1": "src ele2 field1",
				"Field2": "src ele2 field2",
			},
		},
	}, dst)
}

func TestStructToMap_Array_NonEmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 string
	}
	type B struct {
		A [3]A
	}
	src := &B{
		A: [3]A{
			{
				Field1: "src ele1 field1",
				Field2: "src ele1 field2",
			},
			{
				Field1: "src ele2 field1",
				Field2: "src ele2 field2",
			},
			{
				Field1: "src ele3 field1",
				Field2: "src ele3 field2",
			},
		},
	}
	dst := map[string]interface{}{
		"A": [2]map[string]interface{}{
			{
				"Field1": "dst ele1 field1",
			},
			{
				"Field1": "dst ele2 field1",
			},
		},
	}
	mask := fieldmask_utils.MaskFromString("A{Field2}")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"A": []map[string]interface{}{
			{
				"Field1": "dst ele1 field1",
				"Field2": "src ele1 field2",
			},
			{
				"Field1": "dst ele2 field1",
				"Field2": "src ele2 field2",
			},
			{
				"Field2": "src ele3 field2",
			},
		},
	}, dst)
}

func TestStructToMap_ArrayPrimitive_NonEmptyDst(t *testing.T) {
	type A struct {
		Field1 [5]int
	}
	src := &A{
		Field1: [5]int{1, 2, 4, 8, 10},
	}

	dst := map[string]interface{}{
		"Field1": [4]int{16, 32, 64, 0},
	}
	mask := fieldmask_utils.MaskFromString("Field1")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": src.Field1,
	}, dst)
}

func TestStructToMap_EmptySliceSrc_NonEmptyArrayDst(t *testing.T) {
	type A struct {
		Field1 []int
	}
	src := &A{
		Field1: []int{},
	}

	dst := map[string]interface{}{
		"Field1": [4]int{16, 32, 64, 0},
	}
	mask := fieldmask_utils.MaskFromString("Field1")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": src.Field1,
	}, dst)
}
