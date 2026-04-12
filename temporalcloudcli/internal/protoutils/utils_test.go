package protoutils_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/cloud-sdk/api/cloudservice/v1"
	identityv1 "go.temporal.io/cloud-sdk/api/identity/v1"
	"google.golang.org/protobuf/proto"

	"github.com/temporalio/cloud-cli/temporalcloudcli/internal/protoutils"
)

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
