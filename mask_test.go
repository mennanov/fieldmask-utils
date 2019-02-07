package fieldmask_utils_test

import (
	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/mennanov/fieldmask-utils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/protobuf/field_mask"
	"testing"
)

func TestMask_String(t *testing.T) {
	mask := fieldmask_utils.MaskFromString("a{b{c}}")
	assert.Equal(t, "a{b{c}}", mask.String())
}

func TestMaskFromProtoFieldMaskSuccess(t *testing.T) {
	testCases := []struct {
		mask         *field_mask.FieldMask
		expectedTree string
	}{
		{
			&field_mask.FieldMask{Paths: []string{
				"a", // overwritten by the paths below (a.*)
				"a.b.c",
				"a.b.d",
				"a.c.d",
				"b.c.d",
				"a", // has no effect, since more strict rules are applied above
				"c",
			}},
			"a{b{c,d},c{d}},b{c{d}},c",
		},
		{
			&field_mask.FieldMask{Paths: []string{"a", "b", "b", "a"}},
			"a,b",
		},
		{
			&field_mask.FieldMask{Paths: []string{}},
			"",
		},
	}
	for _, testCase := range testCases {
		mask, err := fieldmask_utils.MaskFromProtoFieldMask(testCase.mask, func(s string) string { return s })
		assert.Nil(t, err)
		assert.Equal(t, fieldmask_utils.MaskFromString(testCase.expectedTree), mask)
	}
}

func TestMaskFromProtoFieldMask_Failure(t *testing.T) {
	testCases := []*field_mask.FieldMask{
		{Paths: []string{"a", ".a"}},
		{Paths: []string{"."}},
		{Paths: []string{"a.b.c.d.e", "a.b."}},
	}

	for _, fieldMask := range testCases {
		_, err := fieldmask_utils.MaskFromProtoFieldMask(fieldMask, generator.CamelCase)
		assert.NotNil(t, err)
	}
}

func TestMaskFromString(t *testing.T) {
	testCases := []struct {
		input        string
		expectedMask fieldmask_utils.Mask
		length       int
	}{
		{
			"foo,bar{c{d,e{f,g,h}}}",
			fieldmask_utils.Mask{
				"foo": fieldmask_utils.Mask{},
				"bar": fieldmask_utils.Mask{
					"c": fieldmask_utils.Mask{
						"d": fieldmask_utils.Mask{},
						"e": fieldmask_utils.Mask{
							"f": fieldmask_utils.Mask{},
							"g": fieldmask_utils.Mask{},
							"h": fieldmask_utils.Mask{},
						},
					},
				},
			}, 2,
		},
		{"foo, bar{c {d,e{f,\ng,h}}},t", fieldmask_utils.Mask{
			"foo": fieldmask_utils.Mask{},
			"bar": fieldmask_utils.Mask{
				"c": fieldmask_utils.Mask{
					"d": fieldmask_utils.Mask{},
					"e": fieldmask_utils.Mask{
						"f": fieldmask_utils.Mask{},
						"g": fieldmask_utils.Mask{},
						"h": fieldmask_utils.Mask{},
					},
				},
			},
			"t": fieldmask_utils.Mask{},
		}, 3},
		{"foo", fieldmask_utils.Mask{"foo": fieldmask_utils.Mask{}}, 1},
		{"foo,bar", fieldmask_utils.Mask{
			"foo": fieldmask_utils.Mask{},
			"bar": fieldmask_utils.Mask{},
		}, 2},
		{"foo,bar{c},d,e", fieldmask_utils.Mask{
			"foo": fieldmask_utils.Mask{},
			"bar": fieldmask_utils.Mask{
				"c": fieldmask_utils.Mask{},
			},
			"d": fieldmask_utils.Mask{},
			"e": fieldmask_utils.Mask{},
		}, 4},
		{"", fieldmask_utils.Mask{}, 0},
	}
	for _, testCase := range testCases {
		mask := fieldmask_utils.MaskFromString(testCase.input)
		assert.Equal(t, testCase.expectedMask, mask)
		assert.Equal(t, testCase.length, len(mask))
	}
}
