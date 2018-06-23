package fieldmask_utils

import (
	"fmt"
	"google.golang.org/genproto/protobuf/field_mask"
	"strings"
)

// FieldNameMapper is a function that provides a legitimate Go field name for the given protobuf field name.
// The most common use-case implementation for this type is `generator.CamelCase`.
type FieldNameMapper func(protoFieldName string) (goFieldName string)

// Mask is a tree-based implementation of the protobuf field mask.
type Mask []*MaskField

func (n *Mask) String() string {
	result := make([]string, 0)
	for _, maskNode := range *n {
		maskNodeString := maskNode.String()
		if maskNodeString != "" {
			result = append(result, maskNodeString)
		}
	}
	return strings.Join(result, ",")
}

// IsLeaf returns true if the mask is empty, false otherwise.
func (n *Mask) IsLeaf() bool {
	return len(*n) == 0
}

// GetField gets the field in the mask by its name.
// Returns an error if the field is not found.
func (n *Mask) GetField(protoFieldName string) (*MaskField, error) {
	for _, v := range *n {
		if v.ProtoFieldName == protoFieldName {
			return v, nil
		}
	}
	return nil, fmt.Errorf("field \"%s\" not found in mask %v", protoFieldName, n)
}

// MaskField represents a single field node in a Mask.
type MaskField struct {
	ProtoFieldName string
	GoFieldName    string
	Mask           *Mask
}

func (n *MaskField) String() string {
	subNodes := n.Mask.String()
	if subNodes != "" {
		subNodes = "{" + subNodes + "}"
	}
	return n.ProtoFieldName + subNodes
}

// ParseFieldMask creates a `Mask` (tree-based data structure) from the given FieldMask.
// `mapper` is expected to be a function that maps a field name in protobuf to a legitimate field name in Go.
// In most cases `generator.CamelCase` will work just fine, however see https://github.com/golang/protobuf/issues/457
func ParseFieldMask(fm *field_mask.FieldMask, mapper FieldNameMapper) (*Mask, error) {
	root := &Mask{}
	for _, path := range fm.GetPaths() {
		mask := root
		for _, field := range strings.Split(path, ".") {
			if field == "" {
				return nil, fmt.Errorf("invalid field mask format: \"%s\"", path)
			}
			subNode, err := mask.GetField(field)
			if err != nil {
				subNode = &MaskField{ProtoFieldName: field, GoFieldName: mapper(field), Mask: &Mask{}}
				*mask = append(*mask, subNode)
			}
			mask = subNode.Mask
		}
	}
	return root, nil
}

// MaskFromString creates a `Mask` from a string `s`.
// `s` is supposed to be a valid string representation of a mask like "a,b,c{d,e{f,g}},d".
// This is the same string format as in Mask.String(). MaskFromString should only be used in tests as it does not
// validate the given string and is only convenient to easily create Masks.
func MaskFromString(s string, mapper FieldNameMapper) *Mask {
	mask, _ := maskFromRunes([]rune(s), mapper)
	return mask
}

func maskFromRunes(runes []rune, mapper FieldNameMapper) (*Mask, int) {
	mask := &Mask{}
	fieldName := make([]string, 0)
	runes = append(runes, []rune(",")...)
	pos := 0
	var lastField *MaskField
	for pos < len(runes) {
		char := fmt.Sprintf("%c", runes[pos])
		switch char {
		case " ", "\n", "\t":
			// Ignore white spaces.

		case ",", "{", "}":
			if len(fieldName) == 0 {
				switch char {
				case "}":
					return mask, pos
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
				subMask, jump = maskFromRunes(runes[pos+1:], mapper)
				pos += jump + 1
			} else {
				subMask = &Mask{}
			}
			f := strings.Join(fieldName, "")
			lastField = &MaskField{ProtoFieldName: f, GoFieldName: mapper(f), Mask: subMask}
			*mask = append(*mask, lastField)
			fieldName = make([]string, 0)

			if char == "}" {
				return mask, pos
			}

		default:
			fieldName = append(fieldName, char)
		}
		pos += 1
	}
	return mask, pos
}
