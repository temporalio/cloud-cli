package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

var (
	testCallerRateLimitDefault        = defaultCallerRateLimit(100)
	testCallerRateLimitDefaultRpsZero = defaultCallerRateLimit(0)
	testCallerRateLimitCaller         = &namespacev1.CallerRateLimit{
		Id:        "crl-caller",
		Namespace: "my-ns.acct",
		Spec:      callerRuleSpec("caller-ns.acct", 10),
	}
)

func defaultCallerRateLimit(rps float64) *namespacev1.CallerRateLimit {
	return &namespacev1.CallerRateLimit{
		Id:        "crl-default",
		Namespace: "my-ns.acct",
		Spec:      defaultRuleSpec(rps),
	}
}

func defaultRuleSpec(rps float64) *namespacev1.CallerRateLimitSpec {
	return &namespacev1.CallerRateLimitSpec{
		Rule: &namespacev1.CallerRateLimitSpec_NamespaceDefault{
			NamespaceDefault: &namespacev1.NamespaceDefaultCallerRateLimit{Rps: rps},
		},
	}
}

func callerRuleSpec(caller string, rps float64) *namespacev1.CallerRateLimitSpec {
	return &namespacev1.CallerRateLimitSpec{
		Rule: &namespacev1.CallerRateLimitSpec_NamespaceLimit{
			NamespaceLimit: &namespacev1.NamespaceCallerRateLimit{NamespaceId: caller, Rps: rps},
		},
	}
}

func newCallerRateLimitGetCommand(
	isDefault bool,
	caller string,
) temporalcloudcli.CloudNamespaceCallerRateLimitGetCommand {
	return temporalcloudcli.CloudNamespaceCallerRateLimitGetCommand{
		Namespace:              "my-ns.acct",
		CallerNamespaceDefault: isDefault,
		CallerNamespace:        caller,
	}
}

func callerRateLimitsRequest(isDefault bool, caller string) *cloudservice.GetCallerRateLimitsRequest {
	return &cloudservice.GetCallerRateLimitsRequest{
		Namespace: "my-ns.acct",
		CallerFilter: &cloudservice.GetCallerRateLimitsRequest_CallerFilter{
			CallerType: "namespace",
			Default:    isDefault,
			Caller:     caller,
		},
	}
}

func TestGetCallerRateLimit(t *testing.T) {
	tests := []struct {
		name               string
		isDefault          bool
		caller             string
		expectRequest      bool
		response           []*namespacev1.CallerRateLimit
		responseErr        error
		expectedErr        string
		expectedJSONOutput any
	}{
		{
			name:               "DefaultFound",
			isDefault:          true,
			expectRequest:      true,
			response:           []*namespacev1.CallerRateLimit{testCallerRateLimitDefault},
			expectedJSONOutput: testCallerRateLimitDefault,
		},
		{
			name:               "CallerFound",
			caller:             "caller-ns.acct",
			expectRequest:      true,
			response:           []*namespacev1.CallerRateLimit{testCallerRateLimitCaller},
			expectedJSONOutput: testCallerRateLimitCaller,
		},
		{
			name:          "DefaultFoundRpsZero",
			isDefault:     true,
			expectRequest: true,
			response:      []*namespacev1.CallerRateLimit{testCallerRateLimitDefaultRpsZero},
			// Emit defaults on the expected side too, so the JSON must contain "rps": 0.
			expectedJSONOutput: testCallerRateLimitDefaultRpsZero,
		},
		{
			name:          "DefaultNotFound",
			isDefault:     true,
			expectRequest: true,
			expectedErr:   `default caller rate limit not found on namespace "my-ns.acct"`,
		},
		{
			name:          "CallerNotFound",
			caller:        "caller-ns.acct",
			expectRequest: true,
			expectedErr: `caller rate limit for caller namespace "caller-ns.acct" not found ` +
				`on namespace "my-ns.acct"`,
		},
		{
			name:          "MultipleMatches",
			isDefault:     true,
			expectRequest: true,
			response:      []*namespacev1.CallerRateLimit{testCallerRateLimitDefault, testCallerRateLimitCaller},
			expectedErr: `expected at most one caller rate limit on namespace "my-ns.acct", ` +
				`found 2 (ids: crl-default, crl-caller)`,
		},
		{
			name:          "ListError",
			isDefault:     true,
			expectRequest: true,
			responseErr:   errors.New("api error"),
			expectedErr:   "api error",
		},
		{
			name:        "BothFlagsSet",
			isDefault:   true,
			caller:      "caller-ns.acct",
			expectedErr: "--caller-namespace-default and --caller-namespace are mutually exclusive",
		},
		{
			name:        "NeitherFlagSet",
			expectedErr: "either --caller-namespace-default or --caller-namespace is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := newCallerRateLimitGetCommand(tt.isDefault, tt.caller)
			var expectations func(*cloudmock.MockCloudServiceClient)
			if tt.expectRequest {
				expectations = func(c *cloudmock.MockCloudServiceClient) {
					var resp *cloudservice.GetCallerRateLimitsResponse
					if tt.responseErr == nil {
						resp = &cloudservice.GetCallerRateLimitsResponse{CallerRateLimits: tt.response}
					}
					c.EXPECT().
						GetCallerRateLimits(mock.Anything, callerRateLimitsRequest(tt.isDefault, tt.caller), mock.Anything).
						Return(resp, tt.responseErr)
				}
			}
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations:        expectations,
				JSONOutput:                     true,
				ExpectedError:                  tt.expectedErr,
				ExpectedOutputJson:             tt.expectedJSONOutput,
				ExpectedOutputJsonEmitDefaults: true,
			})
		})
	}
}
