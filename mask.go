package fieldmask_utils

import (
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/genproto/protobuf/field_mask"
	"strings"
)

// FieldNameMapper is a function that provides a legitimate Go field name for the given protobuf field name.
// The most common use-case implementation for this type is `generator.CamelCase` or an identity function.
type FieldNameMapper func(protoFieldName string) (goFieldName string)

// Mask is a tree-based implementation of the protobuf field mask.
type Mask map[string]*Mask

func (n *Mask) String() string {
	if len(*n) == 0 {
		return ""
	}
	var result []string
	for fieldName, maskNode := range *n {
		r := fieldName
		sub := maskNode.String()
		if sub != "" {
			r += "{" + sub + "}"
		}
		result = append(result, r)
	}
	return strings.Join(result, ",")
}

// IsLeaf returns true if the mask is empty, false otherwise.
func (n *Mask) IsLeaf() bool {
	return len(*n) == 0
}

// MaskFromProtoFieldMask creates a `Mask` (tree-based data structure) from the given FieldMask.
func MaskFromProtoFieldMask(fm *field_mask.FieldMask) (*Mask, error) {
	root := Mask{}
	for _, path := range fm.GetPaths() {
		mask := root
		for _, field := range strings.Split(path, ".") {
			if field == "" {
				return nil, errors.Errorf("invalid field mask format: \"%s\"", path)
			}
			subNode, ok := mask[field]
			if !ok {
				mask[field] = &Mask{}
				subNode = mask[field]
			}
			mask = *subNode
		}
	}
	return &root, nil
}

// MaskFromString creates a `Mask` from a string `s`.
// `s` is supposed to be a valid string representation of a mask like "a,b,c{d,e{f,g}},d".
// This is the same string format as in Mask.String(). MaskFromString should only be used in tests as it does not
// validate the given string and is only convenient to easily create Masks.
func MaskFromString(s string) *Mask {
	mask, _ := maskFromRunes([]rune(s))
	return mask
}

func maskFromRunes(runes []rune) (*Mask, int) {
	mask := Mask{}
	fieldName := make([]string, 0)
	runes = append(runes, []rune(",")...)
	pos := 0
	for pos < len(runes) {
		char := fmt.Sprintf("%c", runes[pos])
		switch char {
		case " ", "\n", "\t":
			// Ignore white spaces.

		case ",", "{", "}":
			if len(fieldName) == 0 {
				switch char {
				case "}":
					return &mask, pos
				case ",":
					pos += 1
					continue
				default:
					panic("invalid mask string format")
				}
			}

			var subMask *Mask
			if char == "{" {
				var jump int
				// Parse nested tree.
				subMask, jump = maskFromRunes(runes[pos+1:])
				pos += jump + 1
			} else {
				subMask = &Mask{}
			}
			f := strings.Join(fieldName, "")
			mask[f] = subMask
			// Reset fieldName.
			fieldName = make([]string, 0)

			if char == "}" {
				return &mask, pos
			}

		default:
			fieldName = append(fieldName, char)
		}
		pos += 1
	}
	return &mask, pos
}
