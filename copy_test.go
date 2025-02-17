package fieldmask_utils_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
	"time"

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

func TestStructToStruct_StructToPointer(t *testing.T) {
	v15 := 15
	v42 := 42

	type N struct {
		Field1 int
	}
	type S struct {
		Field1 N
		Field2 int
	}
	src := &S{
		Field1: N{
			Field1: v15,
		},
		Field2: v42,
	}
	type SN struct {
		Field1 *int
	}
	type D struct {
		Field1 *SN
		Field2 *int
	}
	dst := new(D)

	mask := fieldmask_utils.MaskFromString("Field1,Field2")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &D{
		Field1: &SN{
			Field1: &v15,
		},
		Field2: &v42,
	}, dst)
}

func TestStructToStruct_IntToPointer(t *testing.T) {
	v := 42

	type S struct {
		Field2 int
	}
	src := &S{
		Field2: v,
	}
	type D struct {
		Field2 *int
	}
	dst := new(D)

	mask := fieldmask_utils.MaskFromString("Field2")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &D{
		Field2: &v,
	}, dst)
}

func TestStructToStruct_PointerToInt(t *testing.T) {
	v := 42

	type S struct {
		Field2 *int
	}
	src := &S{
		Field2: &v,
	}
	type D struct {
		Field2 int
	}
	dst := new(D)

	mask := fieldmask_utils.MaskFromString("Field2")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, &D{
		Field2: 42,
	}, dst)
}

func TestStructToStruct_Incompatible(t *testing.T) {
	type S struct {
		Field2 int
	}
	src := &S{
		Field2: 42,
	}
	type D struct {
		Field2 string
	}
	dst := new(D)

	mask := fieldmask_utils.MaskFromString("Field2")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.EqualError(t, err, "src kind int differs from dst kind string")
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

func TestStructToStruct_NilSrcSlice_NonEmptyDst(t *testing.T) {
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
		A:      nil,
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

	assert.NoError(t, err)
	assert.Equal(t, "Jessica", src.Stringer.String())
	assert.Equal(t, "Adam", dst.Stringer.String())
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

func TestStructToMap_NilSrcSlice_NonEmptyDst(t *testing.T) {
	type A struct {
		Field1 string
		Field2 string
	}
	type B struct {
		FieldA []A
	}
	src := &B{FieldA: nil}
	dst := map[string]interface{}{
		"FieldA": []map[string]interface{}{
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
	mask := fieldmask_utils.MaskFromString("FieldA")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"FieldA": []map[string]interface{}{},
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
		"Field1": [5]int{1, 2, 4, 8, 10},
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

func TestStructToStruct_CopyStructSlice_WithMaxCopyListSize(t *testing.T) {
	type AA struct {
		Field int
	}
	type A struct {
		Field1 []AA
	}

	src := &A{
		Field1: []AA{{1}, {2}, {3}},
	}
	dst := &A{}

	const copySize int = 2
	mask := fieldmask_utils.MaskFromString("Field1")
	err := fieldmask_utils.StructToStruct(mask, src, dst, fieldmask_utils.WithCopyListSize(func(src *reflect.Value) int {
		return copySize
	}))
	require.NoError(t, err)
	assert.Equal(t, &A{
		Field1: src.Field1[:copySize],
	}, dst)
}

func TestStructToStruct_CopyIntSlice_WithMaxCopyListSize(t *testing.T) {
	type A struct {
		Field1 []int
	}

	src := &A{
		Field1: []int{1, 2, 3},
	}

	const copySize int = 2
	dst := &A{}
	mask := fieldmask_utils.MaskFromString("Field1")
	err := fieldmask_utils.StructToStruct(mask, src, dst, fieldmask_utils.WithCopyListSize(func(src *reflect.Value) int {
		return copySize
	}))
	require.NoError(t, err)
	assert.Equal(t, &A{
		Field1: src.Field1[:copySize],
	}, dst)
}

func TestStructToStruct_CopyIntArray_WithMaxCopyListSize(t *testing.T) {
	const arraySize int = 3
	type A struct {
		Field1 [arraySize]int
	}
	src := &A{
		Field1: [arraySize]int{1, 2, 3},
	}
	const copySize int = arraySize - 1
	dst := &A{}
	mask := fieldmask_utils.MaskFromString("Field1")
	err := fieldmask_utils.StructToStruct(mask, src, dst, fieldmask_utils.WithCopyListSize(func(src *reflect.Value) int {
		return copySize
	}))
	require.NoError(t, err)
	assert.Equal(t, &A{
		Field1: [3]int{1, 2},
	}, dst)
}

func TestStructToStruct_CopyStructArray_WithMaxCopyListSize(t *testing.T) {
	const arraySize int = 3
	type AA struct {
		Field int
	}
	type A struct {
		Field1 [3]AA
	}

	src := &A{
		Field1: [3]AA{{1}, {2}, {3}},
	}
	dst := &A{}

	const copySize int = arraySize - 1
	mask := fieldmask_utils.MaskFromString("Field1")
	err := fieldmask_utils.StructToStruct(mask, src, dst, fieldmask_utils.WithCopyListSize(func(src *reflect.Value) int {
		return copySize
	}))
	require.NoError(t, err)
	assert.Equal(t, &A{
		Field1: [3]AA{{1}, {2}},
	}, dst)
}

func TestStructToMap_CopyStructSlice_WithMaxCopyListSize(t *testing.T) {
	type AA struct {
		Field int
	}
	type A struct {
		Field1 []AA
	}

	src := &A{
		Field1: []AA{{1}, {2}, {3}},
	}
	dst := map[string]interface{}{}

	const copySize int = 2
	mask := fieldmask_utils.MaskFromString("Field1")
	err := fieldmask_utils.StructToMap(mask, src, dst, fieldmask_utils.WithCopyListSize(func(src *reflect.Value) int {
		return copySize
	}))
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": []map[string]interface{}{{"Field": 1}, {"Field": 2}},
	}, dst)
}

func TestStructToMap_CopyIntSlice_WithMaxCopyListSize(t *testing.T) {
	type A struct {
		Field1 []int
	}

	src := &A{
		Field1: []int{1, 2, 3},
	}
	dst := map[string]interface{}{}

	const copySize int = 2
	mask := fieldmask_utils.MaskFromString("Field1")
	err := fieldmask_utils.StructToMap(mask, src, dst, fieldmask_utils.WithCopyListSize(func(src *reflect.Value) int {
		return copySize
	}))
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": []int{1, 2},
	}, dst)
}

func TestStructToMap_CopyStructArray_WithMaxCopyListSize(t *testing.T) {
	const arraySize int = 3
	type AA struct {
		Field int
	}
	type A struct {
		Field1 [3]AA
	}

	src := &A{
		Field1: [3]AA{{1}, {2}, {3}},
	}
	dst := map[string]interface{}{}

	const copySize int = arraySize - 1
	mask := fieldmask_utils.MaskFromString("Field1")
	err := fieldmask_utils.StructToMap(mask, src, dst, fieldmask_utils.WithCopyListSize(func(src *reflect.Value) int {
		return copySize
	}))
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": []map[string]interface{}{{"Field": 1}, {"Field": 2}},
	}, dst)
}

func TestStructToMap_CopyIntArray_WithMaxCopyListSize(t *testing.T) {
	const arraySize int = 3
	type A struct {
		Field1 [arraySize]int
	}
	src := &A{
		Field1: [arraySize]int{1, 2, 3},
	}
	const copySize int = arraySize - 1
	dst := map[string]interface{}{}
	mask := fieldmask_utils.MaskFromString("Field1")
	err := fieldmask_utils.StructToMap(mask, src, dst, fieldmask_utils.WithCopyListSize(func(src *reflect.Value) int {
		return copySize
	}))
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": src.Field1[:copySize],
	}, dst)
}

func TestStructToMap_CopyStructWithPrivateFields_WithMapVisitor(t *testing.T) {
	type A struct {
		Time  time.Time
		Other int
	}
	unixTime := time.Unix(10, 10)
	src := &A{Time: unixTime}
	dst := map[string]interface{}{}
	mask := fieldmask_utils.MaskFromString("Time")
	err := fieldmask_utils.StructToMap(mask, src, dst, fieldmask_utils.WithMapVisitor(
		func(_ fieldmask_utils.FieldFilter, _, dst reflect.Value,
			srcFieldName, dstFieldName string, srcFieldValue reflect.Value) fieldmask_utils.MapVisitorResult {
			if srcFieldName == "Time" {
				return fieldmask_utils.MapVisitorResult{
					SkipToNext: true,
					UpdatedDst: &srcFieldValue,
				}
			}
			return fieldmask_utils.MapVisitorResult{}
		}))
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Time": unixTime,
	}, dst)
}

func TestStructToMap_MapVisitorVisitsOnlyFilteredFields(t *testing.T) {
	type A struct {
		Field1 int
		Field2 string
		Field3 int
	}
	src := &A{Field1: 42, Field2: "hello", Field3: 44}
	dst := map[string]interface{}{}
	mask := fieldmask_utils.MaskFromString("Field1, Field2")
	var visitedFields []string
	err := fieldmask_utils.StructToMap(mask, src, dst, fieldmask_utils.WithMapVisitor(
		func(_ fieldmask_utils.FieldFilter, _, _ reflect.Value,
			srcFieldName, _ string, _ reflect.Value) fieldmask_utils.MapVisitorResult {
			visitedFields = append(visitedFields, srcFieldName)
			return fieldmask_utils.MapVisitorResult{}
		}))
	require.NoError(t, err)
	assert.Equal(t, visitedFields, []string{"Field1", "Field2"})
}

func TestStructToMap_WithMapVisitor_SkipsToNextField(t *testing.T) {
	type A struct {
		Field1 int
		Field2 string
		Field3 int
	}
	src := &A{Field1: 42, Field2: "hello", Field3: 44}
	dst := map[string]interface{}{}
	mask := fieldmask_utils.MaskFromString("Field1, Field2")
	err := fieldmask_utils.StructToMap(mask, src, dst, fieldmask_utils.WithMapVisitor(
		func(_ fieldmask_utils.FieldFilter, _, _ reflect.Value,
			srcFieldName, dstFieldName string, _ reflect.Value) fieldmask_utils.MapVisitorResult {
			if srcFieldName == "Field1" {
				updatedDst := reflect.ValueOf(33)
				return fieldmask_utils.MapVisitorResult{
					SkipToNext: true,
					UpdatedDst: &updatedDst,
				}
			}
			return fieldmask_utils.MapVisitorResult{}
		}))
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"Field1": 33, "Field2": "hello"}, dst)
}

func TestStructToStruct_CopySlice_WithDiffentItemKind(t *testing.T) {
	type A struct {
		Field1 []int
		Field2 []string
	}
	src := &A{
		Field1: []int{1, 2, 3},
		Field2: []string{"1", "2", "3"},
	}
	dst := &A{}
	const copySize int = 1
	mask := fieldmask_utils.MaskFromString("Field1,Field2")
	err := fieldmask_utils.StructToStruct(mask, src, dst, fieldmask_utils.WithCopyListSize(func(src *reflect.Value) int {
		if itemKind := src.Type().Elem().Kind(); itemKind == reflect.Int {
			return copySize
		} else {
			return src.Len()
		}
	}))
	require.NoError(t, err)
	assert.Equal(t, &A{
		Field1: []int{1},
		Field2: []string{"1", "2", "3"},
	}, dst)
}

func TestStructToMap_CopySlice_WithDiffentItemKind(t *testing.T) {
	type A struct {
		Field1 []int
		Field2 []string
	}
	src := &A{
		Field1: []int{1, 2, 3},
		Field2: []string{"1", "2", "3"},
	}
	dst := map[string]interface{}{}
	const copySize int = 1
	mask := fieldmask_utils.MaskFromString("Field1,Field2")
	err := fieldmask_utils.StructToMap(mask, src, dst, fieldmask_utils.WithCopyListSize(func(src *reflect.Value) int {
		if itemKind := src.Type().Elem().Kind(); itemKind == reflect.Int {
			return copySize
		} else {
			return src.Len()
		}
	}))
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": []int{1},
		"Field2": []string{"1", "2", "3"},
	}, dst)
}

func TestStructToStruct_CopySlice_WithDiffentItemType(t *testing.T) {
	type AA struct {
		Int int
	}
	type A struct {
		Field1 []int
		Field2 []AA
	}
	src := &A{
		Field1: []int{1, 2, 3},
		Field2: []AA{{1}, {2}, {3}},
	}
	dst := &A{}
	const copySize int = 1
	mask := fieldmask_utils.MaskFromString("Field1,Field2")
	err := fieldmask_utils.StructToStruct(mask, src, dst, fieldmask_utils.WithCopyListSize(func(src *reflect.Value) int {
		if itemType := src.Type().Elem().Name(); itemType == "AA" {
			return copySize
		} else {
			return src.Len()
		}
	}))
	require.NoError(t, err)
	assert.Equal(t, &A{
		Field1: []int{1, 2, 3},
		Field2: []AA{{1}},
	}, dst)
}

func TestStructToMap_CopySlice_WithDiffentItemType(t *testing.T) {
	type AA struct {
		Int int
	}
	type A struct {
		Field1 []int
		Field2 []AA
	}
	src := &A{
		Field1: []int{1, 2, 3},
		Field2: []AA{{1}, {2}, {3}},
	}
	dst := map[string]interface{}{}
	const copySize int = 1
	mask := fieldmask_utils.MaskFromString("Field1,Field2")
	err := fieldmask_utils.StructToMap(mask, src, dst, fieldmask_utils.WithCopyListSize(func(src *reflect.Value) int {
		if itemType := src.Type().Elem().Name(); itemType == "AA" {
			return copySize
		} else {
			return src.Len()
		}
	}))
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": []int{1, 2, 3},
		"Field2": []map[string]interface{}{{"Int": 1}},
	}, dst)
}

func TestStructToStruct_WithNonStructSrcError(t *testing.T) {
	type A struct{ Field int }
	var src = 1
	var dst = &A{}
	mask := fieldmask_utils.MaskFromString("Field")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	require.Error(t, err)
}

func TestStructToStruct_WithMultiTagComma(t *testing.T) {
	type A struct {
		Field int `json:"field,omitempty"`
	}
	var src = A{Field: 1}
	var dst = map[string]interface{}{}
	mask := fieldmask_utils.MaskFromString("Field")
	err := fieldmask_utils.StructToMap(mask, src, dst, fieldmask_utils.WithTag("json"))
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"field": 1,
	}, dst)
}

func TestStructToMap_WithInterface(t *testing.T) {
	type user struct {
		A string
		B interface{}
		C interface{}
	}
	type c struct {
		A int
		B interface{}
	}
	mask := fieldmask_utils.MaskFromString("A,B,C")

	src := &user{
		A: "nick",
		B: []int{1, 2, 3, 4},
		C: c{A: 42, B: map[string]interface{}{"hi": 34}},
	}
	dst := make(map[string]interface{})
	err := fieldmask_utils.StructToMap(mask, src, dst, fieldmask_utils.WithTag(`json`))
	assert.Nil(t, err)

	expected := map[string]interface{}{
		"A": "nick",
		"B": []int{1, 2, 3, 4},
		"C": map[string]interface{}{"A": 42, "B": map[string]interface{}{"hi": 34}},
	}
	assert.Equal(t, expected, dst)
}

func TestStructToMap_PtrToInt(t *testing.T) {
	type example struct {
		MyInt    *int64
		WhatEver string
	}
	mask := fieldmask_utils.MaskFromString("MyInt,WhatEver")
	myInt := int64(42)

	src := &example{
		MyInt:    &myInt,
		WhatEver: "hello",
	}
	dst := make(map[string]interface{})
	err := fieldmask_utils.StructToMap(mask, src, dst)
	assert.Nil(t, err)

	expected := map[string]interface{}{
		"MyInt":    int64(42),
		"WhatEver": "hello",
	}
	assert.Equal(t, expected, dst)
}

func TestStructToMap_DifferentTypeWithSameDstKey(t *testing.T) {
	t.Skip("this is a programming error which is expected to panic instead of returning an error")
	type BB struct {
		Field int
	}
	type A1 struct {
		FieldA []int
		FieldB []BB `json:"FieldA"`
	}
	var src1 = A1{FieldA: []int{1, 2}, FieldB: []BB{{1}, {2}}}
	var dst1 = map[string]interface{}{}
	mask := fieldmask_utils.MaskFromString("FieldA,FieldB")
	err := fieldmask_utils.StructToMap(mask, src1, dst1, fieldmask_utils.WithTag("json"))
	require.Error(t, err)

	type A2 struct {
		FieldA [2]int
		FieldB [2]BB `json:"FieldA"`
	}
	var src2 = A2{FieldA: [2]int{1, 2}, FieldB: [2]BB{{1}, {2}}}
	var dst2 = map[string]interface{}{}
	mask = fieldmask_utils.MaskFromString("FieldA,FieldB")
	err = fieldmask_utils.StructToMap(mask, src2, dst2, fieldmask_utils.WithTag("json"))
	require.Error(t, err)
}

func TestStructToMap_EmptySrcSlice_JsonEncode(t *testing.T) {
	type A struct{}
	type B struct {
		As []*A
	}

	src := &B{[]*A{}}
	dst := make(map[string]interface{})

	mask := fieldmask_utils.MaskFromString("As")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)

	jsonStr, _ := json.Marshal(dst)
	assert.Equal(t, "{\"As\":[]}", string(jsonStr))
}

func TestStructToMap_NilSrcSlice_JsonEncode(t *testing.T) {
	t.Skip("the behavior that this test verifies has changed")
	type A struct{}
	type B struct {
		As []*A
	}

	src := &B{}
	dst := make(map[string]interface{})

	mask := fieldmask_utils.MaskFromString("As")
	err := fieldmask_utils.StructToMap(mask, src, dst)
	require.NoError(t, err)

	jsonStr, _ := json.Marshal(dst)
	assert.Equal(t, "{\"As\":null}", string(jsonStr))
}

func TestStructToStruct_CopySlice_WithDiffentAddr_WithDifferentFieldName(t *testing.T) {
	type A struct {
		Field1 []int
		Field2 []int
	}

	var src = &A{
		Field1: []int{1, 2, 3},
		Field2: []int{1, 2, 3},
	}
	var field1 = reflect.ValueOf(src).Elem().FieldByName("Field1")
	var dst = &A{}
	var mask = fieldmask_utils.MaskFromString("Field1,Field2")
	var err = fieldmask_utils.StructToStruct(mask, src, dst, fieldmask_utils.WithCopyListSize(
		func(src *reflect.Value) int {
			if src.Pointer() == (&field1).Pointer() {
				return 2
			} else {
				return src.Len()
			}
		},
	))
	require.NoError(t, err)
	assert.Equal(t, &A{
		Field1: []int{1, 2},
		Field2: []int{1, 2, 3},
	}, dst)

}

func TestStructToStruct_CopySlice_WithSameAddr_WithDifferentFieldName(t *testing.T) {
	t.Skip("Not Address this problem")
	type A struct {
		Field1 []int
		Field2 []int
	}

	var arr = []int{1, 2, 3}

	var src = &A{
		Field1: arr,
		Field2: arr,
	}
	var field1 = reflect.ValueOf(src).Elem().FieldByName("Field1")
	var dst = &A{}
	var mask = fieldmask_utils.MaskFromString("Field1,Field2")
	var err = fieldmask_utils.StructToStruct(mask, src, dst, fieldmask_utils.WithCopyListSize(
		func(src *reflect.Value) int {
			if src.Pointer() == (&field1).Pointer() {
				return 2
			} else {
				return src.Len()
			}
		},
	))
	require.NoError(t, err)
	assert.Equal(t, &A{
		Field1: []int{1, 2},
		Field2: []int{1, 2, 3},
	}, dst)
}

func TestStructToStruct_CopyArraySizeAccordingFieldName(t *testing.T) {
	type A struct {
		Field1 [3]int
		Field2 [3]int
	}

	var src = &A{
		Field1: [3]int{1, 2, 3},
		Field2: [3]int{1, 2, 3},
	}
	var field1 = reflect.ValueOf(src).Elem().FieldByName("Field1")
	var dst = &A{}
	var mask = fieldmask_utils.MaskFromString("Field1,Field2")
	var err = fieldmask_utils.StructToStruct(mask, src, dst, fieldmask_utils.WithCopyListSize(
		func(src *reflect.Value) int {
			if src.Addr() == (&field1).Addr() {
				return 2
			} else {
				return src.Len()
			}
		},
	))
	require.NoError(t, err)
	assert.Equal(t, &A{
		Field1: [3]int{1, 2},
		Field2: [3]int{1, 2, 3},
	}, dst)
}

func TestStructToStruct_WithSrcTag(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int `db:"some_field"`
	}
	type B struct {
		Field1 string `struct:"a_name"`
		A      A      `db:"another_name,omitempty"`
	}
	src := &B{
		Field1: "B Field1",
		A: A{
			Field1: "A Field 1",
			Field2: 1,
		},
	}
	dst := &B{}
	mask := fieldmask_utils.MaskFromString("Field1,A{Field2}")
	err := fieldmask_utils.StructToStruct(mask, src, dst, fieldmask_utils.WithSrcTag("db"))
	require.NoError(t, err)
	assert.Equal(t, &B{Field1: src.Field1}, dst)

	mask, _ = fieldmask_utils.MaskFromPaths([]string{"Field1", "another_name.some_field"}, func(s string) string { return s })
	dst = &B{}
	err = fieldmask_utils.StructToStruct(mask, src, dst, fieldmask_utils.WithSrcTag("db"))
	require.NoError(t, err)
	assert.Equal(t, &B{Field1: src.Field1, A: A{Field2: src.A.Field2}}, dst)
}

func TestStructToMap_WithSrcTag(t *testing.T) {
	type A struct {
		Field1 string
		Field2 int  `db:"some_field1" json:"some_field1_json"`
		Field3 bool `db:"some_field2" json:"some_field2_json"`
	}
	type B struct {
		Field1 string `struct:"a_name"`
		A      A      `db:"another_name,omitempty" json:"another_name_json"`
	}
	src := &B{
		Field1: "B Field1",
		A: A{
			Field1: "A Field 1",
			Field2: 1,
		},
	}
	mask := fieldmask_utils.MaskFromString("Field1,A{Field2}")
	dst := make(map[string]interface{})
	err := fieldmask_utils.StructToMap(mask, src, dst, fieldmask_utils.WithTag("json"), fieldmask_utils.WithSrcTag("db"))
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": src.Field1,
	}, dst)

	mask, _ = fieldmask_utils.MaskFromPaths([]string{"Field1", "another_name.some_field1", "another_name.some_field2"}, func(s string) string { return s })
	dst = make(map[string]interface{})
	err = fieldmask_utils.StructToMap(mask, src, dst, fieldmask_utils.WithTag("json"), fieldmask_utils.WithSrcTag("db"))
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{
		"Field1": src.Field1,
		"another_name_json": map[string]interface{}{
			"some_field1_json": src.A.Field2,
			"some_field2_json": false,
		},
	}, dst)
}
