package fieldmask_utils_test

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/mennanov/fieldmask-utils"
	"github.com/mennanov/fieldmask-utils/testproto"
	"github.com/stretchr/testify/assert"
	"testing"
)

var testUserFull *testproto.User
var testUserPartial *testproto.User

func init() {
	ts := &timestamp.Timestamp{
		Seconds: 5, // easy to verify
		Nanos:   6, // easy to verify
	}
	serializedTs, _ := proto.Marshal(ts)

	friend1 := &testproto.User{
		Id:          2,
		Username:    "friend",
		Role:        testproto.Role_REGULAR,
		Meta:        map[string]string{"foo": "bar"},
		Deactivated: true,
		Permissions: []testproto.Permission{testproto.Permission_READ, testproto.Permission_WRITE},
		Avatar: &testproto.Image{
			OriginalUrl: "original.jpg",
			ResizedUrl:  "resized.jpg",
		},
		Images: []*testproto.Image{
			{
				OriginalUrl: "FRIEND original_image1.jpg",
				ResizedUrl:  "FRIEND resized_image1.jpg",
			},
			{
				OriginalUrl: "FRIEND original_image2.jpg",
				ResizedUrl:  "FRIEND resized_image2.jpg",
			},
		},
		Tags: []string{"FRIEND tag1", "FRIEND tag2", "FRIEND tag3"},
		Name: &testproto.User_FemaleName{FemaleName: "Maggy"},
	}
	testUserFull = &testproto.User{
		Id:          1,
		Username:    "username",
		Role:        testproto.Role_ADMIN,
		Meta:        map[string]string{"foo": "bar"},
		Deactivated: true,
		Permissions: []testproto.Permission{testproto.Permission_READ, testproto.Permission_WRITE},
		Avatar: &testproto.Image{
			OriginalUrl: "original.jpg",
			ResizedUrl:  "resized.jpg",
		},
		Images: []*testproto.Image{
			{
				OriginalUrl: "original_image1.jpg",
				ResizedUrl:  "resized_image1.jpg",
			},
			{
				OriginalUrl: "original_image2.jpg",
				ResizedUrl:  "resized_image2.jpg",
			},
		},
		Tags:    []string{"tag1", "tag2", "tag3"},
		Friends: []*testproto.User{friend1},
		Name:    &testproto.User_MaleName{MaleName: "John"},
		Details: []*any.Any{
			{
				TypeUrl: "example.com/example/" + proto.MessageName(ts),
				Value:   serializedTs,
			},
		},
	}
	testUserPartial = &testproto.User{
		Id:       1,
		Username: "username",
	}
}

func TestStructToStructProtoSuccess(t *testing.T) {
	userDst := &testproto.User{}
	mask := fieldmask_utils.MaskFromString(
		"id,avatar{original_url},tags,images,permissions,friends{images{resized_url}},name{male_name}")
	err := fieldmask_utils.StructToStruct(mask, testUserFull, userDst, generator.CamelCase, stringEye)
	assert.Nil(t, err)
	assert.Equal(t, testUserFull.Id, userDst.Id)
	assert.Equal(t, testUserFull.Avatar.OriginalUrl, userDst.Avatar.OriginalUrl)
	assert.Equal(t, "", userDst.Avatar.ResizedUrl)
	assert.Equal(t, testUserFull.Tags, userDst.Tags)
	assert.Equal(t, testUserFull.Images, userDst.Images)
	assert.Equal(t, testUserFull.Name, userDst.Name)
	assert.Equal(t, testUserFull.Permissions, userDst.Permissions)
	assert.Equal(t, len(testUserFull.Friends), len(userDst.Friends))
	assert.Equal(t, len(testUserFull.Friends[0].Images), len(userDst.Friends[0].Images))
	assert.Equal(t, testUserFull.Friends[0].Images[0].ResizedUrl, userDst.Friends[0].Images[0].ResizedUrl)
	assert.Equal(t, "", userDst.Friends[0].Images[0].OriginalUrl)
	// Zero (default) values below.
	assert.Equal(t, testproto.Role_UNKNOWN, userDst.Role)
	assert.Equal(t, false, userDst.Deactivated)
	assert.Equal(t, map[string]string(nil), userDst.Meta)
}

func TestStructToStructEmptyMaskSuccess(t *testing.T) {
	userDst := &testproto.User{}
	mask := fieldmask_utils.MaskFromString("")
	err := fieldmask_utils.StructToStruct(mask, testUserFull, userDst, generator.CamelCase, stringEye)
	assert.Nil(t, err)
	assert.Equal(t, testUserFull, userDst)
}

func TestStructToStructPartialProtoSuccess(t *testing.T) {
	userDst := &testproto.User{}
	mask := fieldmask_utils.MaskFromString(
		"id,avatar{original_url},images,username,permissions,name{male_name}")
	err := fieldmask_utils.StructToStruct(mask, testUserPartial, userDst, generator.CamelCase, stringEye)
	assert.Nil(t, err)
	assert.Equal(t, testUserPartial.Id, userDst.Id)
	assert.Equal(t, testUserPartial.Username, userDst.Username)
	assert.Equal(t, testUserPartial.Name, userDst.Name)
}

func TestStructToStructProtoFail(t *testing.T) {
	userDst := &testproto.User{}
	mask := fieldmask_utils.MaskFromString("id,avatar{unknown_field},tags")
	err := fieldmask_utils.StructToStruct(mask, testUserFull, userDst, generator.CamelCase, stringEye)
	assert.NotNil(t, err)
}

type Name interface {
	someMethod()
}

type FemaleName struct {
	FemaleName string
}

func (*FemaleName) someMethod() {}
func (f *FemaleName) String() string {
	return f.FemaleName
}

type CustomUser struct {
	Name Name
}

func TestStructToStructProtoDifferentInterfacesFail(t *testing.T) {
	userDst := &testproto.User{}
	userSrc := &CustomUser{Name: &FemaleName{FemaleName: "Dana"}}

	mask := fieldmask_utils.MaskFromString("name")
	err := fieldmask_utils.StructToStruct(mask, userSrc, userDst, generator.CamelCase, stringEye)
	assert.NotNil(t, err)
}

func TestStructToStructProtoSameInterfacesSuccess(t *testing.T) {
	type User1 struct {
		Stringer fmt.Stringer
	}

	type User2 struct {
		Stringer fmt.Stringer
	}

	user1 := &User1{
		Stringer: &FemaleName{FemaleName: "Jessica"},
	}

	user2 := &User2{}

	mask := fieldmask_utils.MaskFromString("stringer")
	err := fieldmask_utils.StructToStruct(mask, user1, user2, generator.CamelCase, stringEye)
	assert.Nil(t, err)
	assert.Equal(t, user1.Stringer.String(), user2.Stringer.String())
}

func TestStructToStructNonProtoSuccess(t *testing.T) {
	type Image struct {
		OriginalUrl string
		ResizedUrl  string
	}
	type User struct {
		Id          uint32
		Username    string
		Deactivated bool
		Images      []*Image
	}

	userSrc := &User{
		Id:          1,
		Username:    "johnny",
		Deactivated: true,
		Images: []*Image{
			{"original_url.jpg", "resized_url.jpg"},
			{"original_url.jpg", "resized_url.jpg"},
		},
	}
	userDst := &testproto.User{}
	mask := fieldmask_utils.MaskFromString("")
	err := fieldmask_utils.StructToStruct(mask, userSrc, userDst, generator.CamelCase, stringEye)
	assert.Nil(t, err)
	assert.Equal(t, userSrc.Id, userDst.Id)
	assert.Equal(t, userSrc.Username, userDst.Username)
	assert.Equal(t, len(userSrc.Images), len(userDst.Images))
	assert.Equal(t, userSrc.Images[0].OriginalUrl, userDst.Images[0].OriginalUrl)
	assert.Equal(t, userSrc.Images[0].ResizedUrl, userDst.Images[0].ResizedUrl)
	assert.Equal(t, userSrc.Images[1].OriginalUrl, userDst.Images[1].OriginalUrl)
	assert.Equal(t, userSrc.Images[1].ResizedUrl, userDst.Images[1].ResizedUrl)
	assert.Equal(t, userSrc.Deactivated, userDst.Deactivated)
}

func TestStructToStructNonProtoFail(t *testing.T) {
	type User struct {
		Id           uint32
		UnknownField string
		Deactivated  bool
	}

	userSrc := &User{
		Id:           1,
		UnknownField: "johnny",
		Deactivated:  true,
	}
	userDst := &testproto.User{}
	mask := fieldmask_utils.MaskFromString("")
	err := fieldmask_utils.StructToStruct(mask, userSrc, userDst, generator.CamelCase, stringEye)
	assert.NotNil(t, err)
}

func TestStructToMapSuccess(t *testing.T) {
	userDst := make(map[string]interface{})
	mask := fieldmask_utils.MaskFromString(
		"id,avatar{original_url},tags,images,permissions,friends{images{resized_url}}")
	err := fieldmask_utils.StructToMap(mask, testUserFull, userDst, generator.CamelCase, stringEye)
	assert.Nil(t, err)
	expected := map[string]interface{}{
		"Id": testUserFull.Id,
		"Avatar": map[string]interface{}{
			"OriginalUrl": testUserFull.Avatar.OriginalUrl,
		},
		"Tags": testUserFull.Tags,
		"Images": []map[string]interface{}{
			{"OriginalUrl": testUserFull.Images[0].OriginalUrl, "ResizedUrl": testUserFull.Images[0].ResizedUrl},
			{"OriginalUrl": testUserFull.Images[1].OriginalUrl, "ResizedUrl": testUserFull.Images[1].ResizedUrl},
		},
		"Permissions": testUserFull.Permissions,
		"Friends": []map[string]interface{}{
			{
				"Images": []map[string]interface{}{
					{"ResizedUrl": testUserFull.Friends[0].Images[0].ResizedUrl},
					{"ResizedUrl": testUserFull.Friends[0].Images[1].ResizedUrl},
				},
			},
		},
	}
	assert.Equal(t, expected, userDst)
}

func TestStructToMapPartialProtoSuccess(t *testing.T) {
	userDst := make(map[string]interface{})
	mask := fieldmask_utils.MaskFromString(
		"id,avatar{original_url},images,username,permissions,name{male_name}")
	err := fieldmask_utils.StructToMap(mask, testUserPartial, userDst, generator.CamelCase, stringEye)
	assert.Nil(t, err)
	expected := map[string]interface{}{
		"Id":          testUserPartial.Id,
		"Avatar":      nil,
		"Images":      []map[string]interface{}{},
		"Username":    testUserPartial.Username,
		"Permissions": []interface{}(nil),
		"Name":        nil,
	}
	assert.Equal(t, expected, userDst)
}

func stringEye(s string) string {
	return s
}
