package fieldmask_utils_test

import (
	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/mennanov/fieldmask-utils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/genproto/protobuf/field_mask"
	"testing"
)

func TestParseFieldMaskSuccess(t *testing.T) {
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
		mask, err := fieldmask_utils.ParseFieldMask(testCase.mask, generator.CamelCase)
		assert.Nil(t, err)
		assert.Equal(t, testCase.expectedTree, mask.String())
	}
}

func TestParseFieldMaskFailure(t *testing.T) {
	testCases := []*field_mask.FieldMask{
		{Paths: []string{"a", ".a"}},
		{Paths: []string{"."}},
		{Paths: []string{"a.b.c.d.e", "a.b."}},
	}

	for _, fieldMask := range testCases {
		_, err := fieldmask_utils.ParseFieldMask(fieldMask, generator.CamelCase)
		assert.NotNil(t, err)
	}
}

func TestMaskFromString(t *testing.T) {
	testCases := []struct {
		Input  string
		Output string
		Length int
	}{
		{"foo,bar{c{d,e{f,g,h}}}", "foo,bar{c{d,e{f,g,h}}}", 2},
		{"foo, bar{c {d,e{f,\ng,h}}},t", "foo,bar{c{d,e{f,g,h}}},t", 3},
		{"foo", "foo", 1},
		{"foo,bar", "foo,bar", 2},
		{"foo,bar{c},d,e", "foo,bar{c},d,e", 4},
		{"", "", 0},
	}
	for _, testCase := range testCases {
		mask := fieldmask_utils.MaskFromString(testCase.Input, generator.CamelCase)
		assert.Equal(t, testCase.Output, mask.String())
		assert.Equal(t, testCase.Length, len(*mask))
	}
}
