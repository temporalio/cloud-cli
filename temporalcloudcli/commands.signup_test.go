package temporalcloudcli

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

func TestConnectionAuthParams(t *testing.T) {
	tests := []struct {
		name           string
		domain         string
		method         string
		wantConnection string
		wantScreenHint string
	}{
		{name: "google", domain: "login.tmprl.cloud", method: "google", wantConnection: "google-oauth2"},
		{name: "microsoft prod", domain: "login.tmprl.cloud", method: "microsoft", wantConnection: "microsoft-v2"},
		{name: "microsoft non-prod", domain: "login.tmprl-test.cloud", method: "microsoft", wantConnection: "microsoft"},
		{
			name:           "password",
			domain:         "login.tmprl.cloud",
			method:         "password",
			wantConnection: "Username-Password-Authentication",
			wantScreenHint: "signup",
		},
		{name: "unset leaves the choice to Auth0", domain: "login.tmprl.cloud", method: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CloudSignupCommand{Domain: tt.domain}
			query := authorizeQuery(t, c.connectionAuthParams(tt.method))
			assert.Equal(t, tt.wantConnection, query.Get("connection"))
			assert.Equal(t, tt.wantScreenHint, query.Get("screen_hint"))
		})
	}
}

// Guards against adding a picker entry without a matching Auth0 connection.
func TestLoginMethodsHaveConnections(t *testing.T) {
	c := &CloudSignupCommand{Domain: "login.tmprl.cloud"}
	for _, m := range loginMethods {
		query := authorizeQuery(t, c.connectionAuthParams(m.value))
		assert.NotEmpty(t, query.Get("connection"), "login method %q has no connection", m.value)
	}
}

func authorizeQuery(t *testing.T, opts []oauth2.AuthCodeOption) url.Values {
	t.Helper()
	cfg := &oauth2.Config{Endpoint: oauth2.Endpoint{AuthURL: "https://example.com/authorize"}}
	u, err := url.Parse(cfg.AuthCodeURL("state", opts...))
	require.NoError(t, err)
	return u.Query()
}
