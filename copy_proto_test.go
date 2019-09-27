package fieldmask_utils_test

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/golang/protobuf/ptypes"

	fieldmask_utils "github.com/mennanov/fieldmask-utils"
	"github.com/mennanov/fieldmask-utils/testproto"
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

	extraUser, err := ptypes.MarshalAny(testUserFull)
	if err != nil {
		panic(err)
	}

	testUserFull.ExtraUser = extraUser
	testUserPartial = &testproto.User{
		Id:       1,
		Username: "username",
	}
}

func TestStructToStruct_Proto(t *testing.T) {
	userDst := &testproto.User{}
	mask := fieldmask_utils.MaskFromString(
		"Id,Avatar{OriginalUrl},Tags,Images,Permissions,Friends{Images{ResizedUrl}},Name{MaleName},ExtraUser{Id,Avatar{OriginalUrl}}")
	err := fieldmask_utils.StructToStruct(mask, testUserFull, userDst)
	require.NoError(t, err)
	assert.Equal(t, testUserFull.Id, userDst.Id)
	assert.Equal(t, testUserFull.Avatar.OriginalUrl, userDst.Avatar.OriginalUrl)
	assert.Equal(t, "", userDst.Avatar.ResizedUrl)
	assert.Equal(t, testUserFull.Tags, userDst.Tags)
	require.Equal(t, len(testUserFull.Images), len(userDst.Images))
	for i, srcImg := range testUserFull.Images {
		assert.Equal(t, srcImg.OriginalUrl, userDst.Images[i].OriginalUrl)
		assert.Equal(t, srcImg.ResizedUrl, userDst.Images[i].ResizedUrl)
	}
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

	extraUser := &testproto.User{}
	err = ptypes.UnmarshalAny(userDst.ExtraUser, extraUser)
	require.NoError(t, err)
	assert.Equal(t, testUserFull.Id, extraUser.Id)
	assert.Equal(t, testUserFull.Avatar.OriginalUrl, extraUser.Avatar.OriginalUrl)
}

func TestStructToStruct_ExistingAnyPreserved(t *testing.T) {
	existingExtraUser := &testproto.User{
		Id:       42,
		Username: "username",
	}
	existingExtraUserAny, err := ptypes.MarshalAny(existingExtraUser)
	require.NoError(t, err)
	userDst := &testproto.User{
		ExtraUser: existingExtraUserAny,
	}
	mask := fieldmask_utils.MaskFromString("ExtraUser{Id,Avatar{OriginalUrl}}")
	err = fieldmask_utils.StructToStruct(mask, testUserFull, userDst)
	require.NoError(t, err)

	extraUser := &testproto.User{}
	err = ptypes.UnmarshalAny(userDst.ExtraUser, extraUser)
	require.NoError(t, err)
	assert.Equal(t, testUserFull.Id, extraUser.Id)
	assert.Equal(t, testUserFull.Avatar.OriginalUrl, extraUser.Avatar.OriginalUrl)
	assert.Equal(t, "username", extraUser.Username)
}

func TestStructToStruct_PartialProtoSuccess(t *testing.T) {
	userDst := &testproto.User{}
	mask := fieldmask_utils.MaskFromString(
		"Id,Avatar{OriginalUrl},Images,Username,Permissions,Name{MaleName}")
	err := fieldmask_utils.StructToStruct(mask, testUserPartial, userDst)
	assert.Nil(t, err)
	assert.Equal(t, testUserPartial.Id, userDst.Id)
	assert.Equal(t, testUserPartial.Username, userDst.Username)
	assert.Equal(t, testUserPartial.Name, userDst.Name)
}

func TestStructToStruct_MaskInverse(t *testing.T) {
	userSrc := &testproto.User{
		Id:          1,
		Username:    "username",
		Role:        testproto.Role_ADMIN,
		Meta:        map[string]string{"foo": "bar"},
		Deactivated: false,
		Permissions: []testproto.Permission{testproto.Permission_EXECUTE},
		Name:        &testproto.User_FemaleName{FemaleName: "Dana"},
		Friends: []*testproto.User{
			{Id: 2, Username: "friend1"},
			{Id: 3, Username: "friend2"},
		},
	}
	userDst := &testproto.User{}
	mask := fieldmask_utils.MaskInverse{"Id": nil, "Friends": fieldmask_utils.MaskInverse{"Username": nil}}
	err := fieldmask_utils.StructToStruct(mask, userSrc, userDst)
	require.NoError(t, err)
	// Verify that Id is not copied.
	assert.Equal(t, uint32(0), userDst.Id)
	// Verify that Friend Usernames are not copied.
	assert.Equal(t, "", userDst.Friends[0].Username)
	assert.Equal(t, "", userDst.Friends[1].Username)
	// Copy missed fields manually and then compare these structs.
	userDst.Id = userSrc.Id
	userDst.Friends[0].Username = userSrc.Friends[0].Username
	userDst.Friends[1].Username = userSrc.Friends[1].Username
	assert.Equal(t, userSrc, userDst)
}

func TestStructToStruct_NonProtoSuccess(t *testing.T) {
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

	src := &User{
		Id:          1,
		Username:    "johnny",
		Deactivated: true,
		Images: []*Image{
			{"original_url.jpg", "resized_url.jpg"},
			{"original_url.jpg", "resized_url.jpg"},
		},
	}
	dst := &testproto.User{}
	mask := fieldmask_utils.MaskFromString("")
	err := fieldmask_utils.StructToStruct(mask, src, dst)
	assert.NoError(t, err)
	assert.Equal(t, src.Id, dst.Id)
	assert.Equal(t, src.Username, dst.Username)
	assert.Equal(t, len(src.Images), len(dst.Images))
	assert.Equal(t, src.Images[0].OriginalUrl, dst.Images[0].OriginalUrl)
	assert.Equal(t, src.Images[0].ResizedUrl, dst.Images[0].ResizedUrl)
	assert.Equal(t, src.Images[1].OriginalUrl, dst.Images[1].OriginalUrl)
	assert.Equal(t, src.Images[1].ResizedUrl, dst.Images[1].ResizedUrl)
	assert.Equal(t, src.Deactivated, dst.Deactivated)
}

func TestStructToStruct_MaskInverseFromMask(t *testing.T) {
	userSrc := &testproto.User{
		Id:          1,
		Username:    "username",
		Role:        testproto.Role_ADMIN,
		Meta:        map[string]string{"foo": "bar"},
		Deactivated: false,
		Permissions: []testproto.Permission{testproto.Permission_EXECUTE},
		Name:        &testproto.User_FemaleName{FemaleName: "Dana"},
		Friends: []*testproto.User{
			{Id: 2, Username: "friend1"},
			{Id: 3, Username: "friend2"},
		},
	}
	userDst := &testproto.User{}
	// Mask to MaskInverse
	mask := fieldmask_utils.MaskInverse{"Id": fieldmask_utils.Mask{}, "Friends": fieldmask_utils.Mask{"Username": fieldmask_utils.Mask{}}}
	err := fieldmask_utils.StructToStruct(mask, userSrc, userDst)
	require.NoError(t, err)
	assert.Equal(t, &testproto.User{
		Username:    userSrc.Username,
		Role:        userSrc.Role,
		Meta:        userSrc.Meta,
		Deactivated: userSrc.Deactivated,
		Permissions: userSrc.Permissions,
		Name:        userSrc.Name,
		Friends: []*testproto.User{
			{Username: "friend1"},
			{Username: "friend2"},
		},
	}, userDst)
}

func TestStructToStruct_NonProtoFail(t *testing.T) {
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
	err := fieldmask_utils.StructToStruct(mask, userSrc, userDst)
	assert.NotNil(t, err)
}

func TestStructToMap_Success(t *testing.T) {
	userDst := make(map[string]interface{})
	mask := fieldmask_utils.MaskFromString(
		"Id,Avatar{OriginalUrl},Tags,Images,Permissions,Friends{Images{ResizedUrl}}")
	err := fieldmask_utils.StructToMap(mask, testUserFull, userDst)
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

func TestStructToMap_PartialProtoSuccess(t *testing.T) {
	userDst := make(map[string]interface{})
	mask := fieldmask_utils.MaskFromString(
		"Id,Avatar{OriginalUrl},Images,Username,Permissions,Name{MaleName}")
	err := fieldmask_utils.StructToMap(mask, testUserPartial, userDst)
	assert.Nil(t, err)
	expected := map[string]interface{}{
		"Id":          testUserPartial.Id,
		"Avatar":      nil,
		"Images":      []map[string]interface{}(nil),
		"Username":    testUserPartial.Username,
		"Permissions": []testproto.Permission(nil),
		"Name":        nil,
	}
	assert.Equal(t, expected, userDst)
}
