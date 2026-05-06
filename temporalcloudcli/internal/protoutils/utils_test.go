package protoutils_test

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/protoutils"
)

var updateGolden = flag.Bool("update-golden", false, "update golden files")

func TestClearDeprecatedFields(t *testing.T) {
	tests := []struct {
		name   string
		input  proto.Message
		verify func(t *testing.T, got proto.Message)
	}{
		{
			name: "TopLevelDeprecatedField",
			input: &identityv1.ApiKey{
				Id:              "key-1",
				StateDeprecated: "active",
			},
			verify: func(t *testing.T, got proto.Message) {
				key := got.(*identityv1.ApiKey)
				assert.Equal(t, "key-1", key.Id)
				assert.Empty(t, key.StateDeprecated)
			},
		},
		{
			name: "NoDeprecatedFields",
			input: &identityv1.ApiKey{
				Id: "key-1",
			},
			verify: func(t *testing.T, got proto.Message) {
				key := got.(*identityv1.ApiKey)
				assert.Equal(t, "key-1", key.Id)
			},
		},
		{
			name: "NestedMessage",
			input: &identityv1.ApiKey{
				Id: "key-1",
				Spec: &identityv1.ApiKeySpec{
					DisplayName:         "my-key",
					OwnerTypeDeprecated: "user",
				},
			},
			verify: func(t *testing.T, got proto.Message) {
				key := got.(*identityv1.ApiKey)
				assert.Equal(t, "key-1", key.Id)
				assert.Equal(t, "my-key", key.Spec.DisplayName)
				assert.Empty(t, key.Spec.OwnerTypeDeprecated)
			},
		},
		{
			name: "MapMessageValues",
			input: &identityv1.Access{
				NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
					"ns1": {
						Permission:           identityv1.NamespaceAccess_PERMISSION_READ,
						PermissionDeprecated: "read",
					},
				},
			},
			verify: func(t *testing.T, got proto.Message) {
				access := got.(*identityv1.Access)
				ns1 := access.NamespaceAccesses["ns1"]
				require.NotNil(t, ns1)
				assert.Equal(t, identityv1.NamespaceAccess_PERMISSION_READ, ns1.Permission)
				assert.Empty(t, ns1.PermissionDeprecated)
			},
		},
		{
			name: "RepeatedMessages",
			input: &cloudservice.GetApiKeysResponse{
				ApiKeys: []*identityv1.ApiKey{
					{Id: "key-1", StateDeprecated: "active"},
					{Id: "key-2", StateDeprecated: "deleted"},
				},
			},
			verify: func(t *testing.T, got proto.Message) {
				resp := got.(*cloudservice.GetApiKeysResponse)
				require.Len(t, resp.ApiKeys, 2)
				assert.Equal(t, "key-1", resp.ApiKeys[0].Id)
				assert.Empty(t, resp.ApiKeys[0].StateDeprecated)
				assert.Equal(t, "key-2", resp.ApiKeys[1].Id)
				assert.Empty(t, resp.ApiKeys[1].StateDeprecated)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			protoutils.ClearDeprecatedFields(tt.input)
			tt.verify(t, tt.input)
		})
	}
}

func TestStripDeprecatedJSONFields(t *testing.T) {
	tests := []struct {
		name  string
		input proto.Message
	}{
		{
			name: "removes_suffix_deprecated_top_level_field",
			input: &identityv1.ApiKey{
				Id:              "key-1",
				StateDeprecated: "active",
			},
		},
		{
			name: "removes_suffix_deprecated_nested_field",
			input: &identityv1.ApiKey{
				Id: "key-1",
				Spec: &identityv1.ApiKeySpec{
					DisplayName:         "my-key",
					OwnerTypeDeprecated: "user",
				},
			},
		},
		{
			name: "removes_option_deprecated_field_without_suffix",
			input: &namespacev1.NamespaceSpec{
				Name:                   "my-ns",
				Regions:                []string{"aws-us-west-2"},
				CustomSearchAttributes: map[string]string{"k": "v"},
				RetentionDays:          7,
			},
		},
		{
			name: "removes_deprecated_field_in_repeated_message",
			input: &cloudservice.GetApiKeysResponse{
				ApiKeys: []*identityv1.ApiKey{
					{Id: "key-1", StateDeprecated: "active"},
					{Id: "key-2", StateDeprecated: "deleted"},
				},
			},
		},
		{
			name: "removes_deprecated_field_in_map_value_message",
			input: &identityv1.Access{
				NamespaceAccesses: map[string]*identityv1.NamespaceAccess{
					"ns1": {
						Permission:           identityv1.NamespaceAccess_PERMISSION_READ,
						PermissionDeprecated: "read",
					},
				},
			},
		},
		{
			name: "preserves_non_deprecated_fields",
			input: &identityv1.ApiKey{
				Id: "key-1",
				Spec: &identityv1.ApiKeySpec{
					DisplayName: "my-key",
					OwnerId:     "owner-1",
				},
			},
		},
	}
	marshaler := protojson.MarshalOptions{
		EmitUnpopulated: true,
		Indent:          "    ",
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := marshaler.Marshal(tt.input)
			require.NoError(t, err)
			got, err := protoutils.StripDeprecatedJSONFields(data, tt.input)
			require.NoError(t, err)

			golden := filepath.Join("testdata", "strip_deprecated", tt.name+".golden.json")
			if *updateGolden {
				require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0o755))
				require.NoError(t, os.WriteFile(golden, got, 0o644))
			}
			want, err := os.ReadFile(golden)
			require.NoError(t, err)
			assert.JSONEq(t, string(want), string(got), "output does not match contents of golden file: have you run --update-golden?")
		})
	}
}

func TestStripDeprecatedJSONFields_InvalidJSON(t *testing.T) {
	_, err := protoutils.StripDeprecatedJSONFields([]byte("not-valid-json"), &identityv1.ApiKey{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unable to parse json")
}
