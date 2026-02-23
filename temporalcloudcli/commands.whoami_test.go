//go:build integration
// +build integration

package temporalcloudcli_test

import (
	"fmt"
	"io"

	"go.temporal.io/cloud-sdk/api/cloudservice/v1"
	"google.golang.org/protobuf/encoding/protojson"
)

func (s *SharedServerSuite) TestWhoami() {
	res := s.Execute(
		"whoami",
		fmt.Sprintf("--server=%s", s.server),
		"-o=json",
	)
	s.Suite.Require().NoError(res.Err)

	buf, err := io.ReadAll(&res.Stdout)
	s.Suite.Require().NoError(err)

	identity := &cloudservice.GetCurrentIdentityResponse{}
	err = protojson.Unmarshal(buf, identity)
	s.Suite.Require().NoError(err)

	// The response must identify the caller as either a user or a service account.
	s.Suite.True(
		identity.GetUser() != nil || identity.GetServiceAccount() != nil,
		"expected identity to be a user or service account",
	)
	if identity.GetServiceAccount() != nil {
		s.Suite.NotNil(identity.GetPrincipalApiKey(), "expected service account to have a principal API key")
	}
}
