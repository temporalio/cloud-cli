package temporalcloudcli_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	cliext "github.com/temporalio/cli/cliext"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	resourcev1 "go.temporal.io/cloud-sdk/api/resource/v1"

	cloudmock "github.com/temporalio/cloud-cli/internal/cloudservice/mock"
	"github.com/temporalio/cloud-cli/temporalcloudcli"
)

type (
	callerRateLimitListOutput struct {
		CallerRateLimits []*namespacev1.CallerRateLimit
		NextPageToken    string
	}

	callerRateLimitGetFlags struct {
		isDefault bool
		caller    string
	}

	callerRateLimitListFlags struct {
		callerType string
		isDefault  bool
		caller     string
		pageSize   int
		pageToken  string
	}
)

var (
	testCallerRateLimitDefault        = defaultCallerRateLimit(100)
	testCallerRateLimitDefaultRpsZero = defaultCallerRateLimit(0)
	testCallerRateLimitCaller         = &namespacev1.CallerRateLimit{
		Id:        "crl-caller",
		Namespace: "my-ns.acct",
		State:     resourcev1.ResourceState_RESOURCE_STATE_ACTIVE,
		Spec:      callerRuleSpec("caller-ns.acct", 10),
	}
	testCallerRateLimitCallerRpsZero = &namespacev1.CallerRateLimit{
		Id:        "crl-caller-rps-zero",
		Namespace: "my-ns.acct",
		State:     resourcev1.ResourceState_RESOURCE_STATE_ACTIVE,
		Spec:      callerRuleSpec("blocked-ns.acct", 0),
	}
)

func defaultCallerRateLimit(rps float64) *namespacev1.CallerRateLimit {
	return &namespacev1.CallerRateLimit{
		Id:        "crl-default",
		Namespace: "my-ns.acct",
		State:     resourcev1.ResourceState_RESOURCE_STATE_ACTIVE,
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

func (f callerRateLimitGetFlags) command() temporalcloudcli.CloudNamespaceCallerRateLimitGetCommand {
	return temporalcloudcli.CloudNamespaceCallerRateLimitGetCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
		CallerRateLimitCallerOptions: temporalcloudcli.CallerRateLimitCallerOptions{
			CallerNamespaceDefault: f.isDefault,
			CallerNamespace:        f.caller,
		},
	}
}

func TestGetCallerRateLimit(t *testing.T) {
	tests := []struct {
		name               string
		flags              callerRateLimitGetFlags
		found              []*namespacev1.CallerRateLimit
		lookupErr          error
		expectedErr        string
		expectedJSONOutput any
	}{
		{
			name:               "DefaultFound",
			flags:              callerRateLimitGetFlags{isDefault: true},
			found:              []*namespacev1.CallerRateLimit{testCallerRateLimitDefault},
			expectedJSONOutput: testCallerRateLimitDefault,
		},
		{
			name:               "CallerFound",
			flags:              callerRateLimitGetFlags{caller: "caller-ns.acct"},
			found:              []*namespacev1.CallerRateLimit{testCallerRateLimitCaller},
			expectedJSONOutput: testCallerRateLimitCaller,
		},
		{
			name:  "DefaultFoundRpsZero",
			flags: callerRateLimitGetFlags{isDefault: true},
			found: []*namespacev1.CallerRateLimit{testCallerRateLimitDefaultRpsZero},
			// Emit defaults on the expected side too, so the JSON must contain "rps": 0.
			expectedJSONOutput: testCallerRateLimitDefaultRpsZero,
		},
		{
			name:        "DefaultNotFound",
			flags:       callerRateLimitGetFlags{isDefault: true},
			expectedErr: `default caller rate limit not found on namespace "my-ns.acct"`,
		},
		{
			name:  "CallerNotFound",
			flags: callerRateLimitGetFlags{caller: "caller-ns.acct"},
			expectedErr: `caller rate limit for caller namespace "caller-ns.acct" not found ` +
				`on namespace "my-ns.acct"`,
		},
		{
			name:  "MultipleMatches",
			flags: callerRateLimitGetFlags{isDefault: true},
			found: []*namespacev1.CallerRateLimit{testCallerRateLimitDefault, testCallerRateLimitCaller},
			expectedErr: `expected at most one caller rate limit on namespace "my-ns.acct", ` +
				`found 2 (ids: crl-default, crl-caller)`,
		},
		{
			name:        "LookupError",
			flags:       callerRateLimitGetFlags{isDefault: true},
			lookupErr:   errors.New("api error"),
			expectedErr: "api error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.flags.command()
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					c.EXPECT().
						GetCallerRateLimits(mock.Anything, callerRateLimitsRequest(tt.flags.isDefault, tt.flags.caller),
							mock.Anything).
						Return(&cloudservice.GetCallerRateLimitsResponse{CallerRateLimits: tt.found}, tt.lookupErr)
				},
				JSONOutput:                     true,
				ExpectedError:                  tt.expectedErr,
				ExpectedOutputJson:             tt.expectedJSONOutput,
				ExpectedOutputJsonEmitDefaults: true,
			})
		})
	}
}

func TestGetCallerRateLimit_FlagValidation(t *testing.T) {
	tests := []struct {
		name        string
		flags       callerRateLimitGetFlags
		expectedErr string
	}{
		{
			name:        "BothFlagsSet",
			flags:       callerRateLimitGetFlags{isDefault: true, caller: "caller-ns.acct"},
			expectedErr: "--caller-namespace-default and --caller-namespace are mutually exclusive",
		},
		{
			name:        "NeitherFlagSet",
			expectedErr: "either --caller-namespace-default or --caller-namespace is required",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.flags.command()
			// No CloudClientExpectations: any RPC fails the test, proving validation runs first.
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				ExpectedError: tt.expectedErr,
			})
		})
	}
}

func (f callerRateLimitListFlags) command() temporalcloudcli.CloudNamespaceCallerRateLimitListCommand {
	return temporalcloudcli.CloudNamespaceCallerRateLimitListCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
		CallerRateLimitCallerOptions: temporalcloudcli.CallerRateLimitCallerOptions{
			CallerNamespaceDefault: f.isDefault,
			CallerNamespace:        f.caller,
		},
		Type:      cliext.FlagStringEnum{Value: f.callerType},
		PageSize:  f.pageSize,
		PageToken: f.pageToken,
	}
}

func TestListCallerRateLimits(t *testing.T) {
	// The mock echoes this response for every case; the cases differ only in the request run builds.
	// The rps 0 rule checks the JSON keeps "rps": 0.
	allLimits := []*namespacev1.CallerRateLimit{
		testCallerRateLimitDefault,
		testCallerRateLimitCaller,
		testCallerRateLimitCallerRpsZero,
	}
	expectedOutput := callerRateLimitListOutput{CallerRateLimits: allLimits, NextPageToken: "page-3"}
	tests := []struct {
		name               string
		flags              callerRateLimitListFlags
		expectedRequest    *cloudservice.GetCallerRateLimitsRequest
		lookupErr          error
		expectedErr        string
		expectedJSONOutput any
	}{
		{
			name:  "NoFilterWithPagination",
			flags: callerRateLimitListFlags{pageSize: 5, pageToken: "page-2"},
			expectedRequest: &cloudservice.GetCallerRateLimitsRequest{
				Namespace: "my-ns.acct",
				PageSize:  5,
				PageToken: "page-2",
			},
			expectedJSONOutput: expectedOutput,
		},
		{
			name:  "TypeOnly",
			flags: callerRateLimitListFlags{callerType: "namespace"},
			expectedRequest: &cloudservice.GetCallerRateLimitsRequest{
				Namespace: "my-ns.acct",
				CallerFilter: &cloudservice.GetCallerRateLimitsRequest_CallerFilter{
					CallerType: "namespace",
				},
			},
			expectedJSONOutput: expectedOutput,
		},
		{
			name:               "DefaultOnly",
			flags:              callerRateLimitListFlags{isDefault: true},
			expectedRequest:    callerRateLimitsRequest(true, ""),
			expectedJSONOutput: expectedOutput,
		},
		{
			name:               "CallerOnly",
			flags:              callerRateLimitListFlags{caller: "caller-ns.acct"},
			expectedRequest:    callerRateLimitsRequest(false, "caller-ns.acct"),
			expectedJSONOutput: expectedOutput,
		},
		{
			name: "DefaultAndCallerUnion",
			flags: callerRateLimitListFlags{
				callerType: "namespace",
				isDefault:  true,
				caller:     "caller-ns.acct",
			},
			expectedRequest:    callerRateLimitsRequest(true, "caller-ns.acct"),
			expectedJSONOutput: expectedOutput,
		},
		{
			name:            "LookupError",
			expectedRequest: &cloudservice.GetCallerRateLimitsRequest{Namespace: "my-ns.acct"},
			lookupErr:       errors.New("api error"),
			expectedErr:     "api error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.flags.command()
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					c.EXPECT().
						GetCallerRateLimits(mock.Anything, tt.expectedRequest, mock.Anything).
						Return(&cloudservice.GetCallerRateLimitsResponse{
							CallerRateLimits: allLimits,
							NextPageToken:    "page-3",
						}, tt.lookupErr)
				},
				JSONOutput:                     true,
				ExpectedError:                  tt.expectedErr,
				ExpectedOutputJson:             tt.expectedJSONOutput,
				ExpectedOutputJsonEmitDefaults: true,
			})
		})
	}
}

// AIDEV-NOTE: Table (non-JSON) output guard test. Rows are not flattened, so the
// Rule column renders the spec's oneof as JSON.
func TestListCallerRateLimits_TableOutput(t *testing.T) {
	cmd := callerRateLimitListFlags{}.command()
	expectedTable := "           Id                  State                                      Rule                            \n" +
		"  crl-default          RESOURCE_STATE_ACTIVE  {\"NamespaceDefault\":{\"rps\":100}}                            \n" +
		"  crl-caller           RESOURCE_STATE_ACTIVE  {\"NamespaceLimit\":{\"namespaceId\":\"caller-ns.acct\",\"rps\":10}}\n" +
		"  crl-caller-rps-zero  RESOURCE_STATE_ACTIVE  {\"NamespaceLimit\":{\"namespaceId\":\"blocked-ns.acct\",\"rps\":0}}\n"
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetCallerRateLimits(mock.Anything, mock.Anything, mock.Anything).
				Return(&cloudservice.GetCallerRateLimitsResponse{
					CallerRateLimits: []*namespacev1.CallerRateLimit{
						testCallerRateLimitDefault,
						testCallerRateLimitCaller,
						testCallerRateLimitCallerRpsZero,
					},
				}, nil)
		},
		ExpectedOutput: expectedTable,
	})
}
