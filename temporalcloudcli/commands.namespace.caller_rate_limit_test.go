package temporalcloudcli_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	cliext "github.com/temporalio/cli/cliext"
	cloudservice "go.temporal.io/cloud-sdk/api/cloudservice/v1"
	namespacev1 "go.temporal.io/cloud-sdk/api/namespace/v1"
	"go.temporal.io/cloud-sdk/api/operation/v1"
	resourcev1 "go.temporal.io/cloud-sdk/api/resource/v1"
	"google.golang.org/protobuf/proto"

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

	callerRateLimitDeleteFlags struct {
		isDefault        bool
		resourceVersion  string
		asyncOperationID string
		idempotent       bool
	}

	callerRateLimitSetFlags struct {
		isDefault        bool
		caller           string
		rps              float32
		resourceVersion  string
		asyncOperationID string
	}

	callerRateLimitApplyFlags struct {
		spec             string
		resourceVersion  string
		asyncOperationID string
		verboseDiff      bool
		async            bool
	}

	// mutationExpectation sets one Create or Update expectation and returns it for ordering.
	mutationExpectation func(c *cloudmock.MockCloudServiceClient) *mock.Call
)

var (
	testCallerRateLimitDefault        = defaultCallerRateLimit(100)
	testCallerRateLimitDefaultRpsZero = defaultCallerRateLimit(0)
	testCallerRateLimitCaller         = &namespacev1.CallerRateLimit{
		Id:              "crl-caller",
		Namespace:       "my-ns.acct",
		ResourceVersion: "rv-caller",
		State:           resourcev1.ResourceState_RESOURCE_STATE_ACTIVE,
		Spec:            callerRuleSpec("caller-ns.acct", 10),
	}
	testCallerRateLimitCallerRpsZero = &namespacev1.CallerRateLimit{
		Id:              "crl-caller-rps-zero",
		Namespace:       "my-ns.acct",
		ResourceVersion: "rv-caller-rps-zero",
		State:           resourcev1.ResourceState_RESOURCE_STATE_ACTIVE,
		Spec:            callerRuleSpec("blocked-ns.acct", 0),
	}
)

func defaultCallerRateLimit(rps float64) *namespacev1.CallerRateLimit {
	return &namespacev1.CallerRateLimit{
		Id:              "crl-default",
		Namespace:       "my-ns.acct",
		ResourceVersion: "rv-default",
		State:           resourcev1.ResourceState_RESOURCE_STATE_ACTIVE,
		Spec:            defaultRuleSpec(rps),
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

// promptApply expects PromptApply(existing, modified, verbose) and confirms with result.
func promptApply(existing, modified proto.Message, verbose, result bool) temporalcloudcli.TestPromptOptions {
	return temporalcloudcli.TestPromptOptions{
		ExpectPrompApply:          true,
		ExpectPromptApplyExisting: existing,
		ExpectPromptApplyModified: modified,
		ExpectPromptApplyVerbose:  verbose,
		PromptResult:              result,
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
			name: "DefaultAndCallerBothSent",
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

func (f callerRateLimitDeleteFlags) command() temporalcloudcli.CloudNamespaceCallerRateLimitDeleteCommand {
	return temporalcloudcli.CloudNamespaceCallerRateLimitDeleteCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
		CallerRateLimitCallerOptions: temporalcloudcli.CallerRateLimitCallerOptions{
			CallerNamespaceDefault: f.isDefault,
		},
		AsyncOperationOptions: temporalcloudcli.AsyncOperationOptions{
			Idempotent:       f.idempotent,
			AsyncOperationId: f.asyncOperationID,
		},
		ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: f.resourceVersion},
	}
}

func expectFindDefaultCallerRateLimit(
	c *cloudmock.MockCloudServiceClient,
	found []*namespacev1.CallerRateLimit,
	err error,
) {
	c.EXPECT().
		GetCallerRateLimits(mock.Anything, callerRateLimitsRequest(true, ""), mock.Anything).
		Return(&cloudservice.GetCallerRateLimitsResponse{CallerRateLimits: found}, err)
}

func deleteCallerRateLimitRequest(rv, asyncOpID string) *cloudservice.DeleteCallerRateLimitRequest {
	return &cloudservice.DeleteCallerRateLimitRequest{
		Namespace:         "my-ns.acct",
		CallerRateLimitId: "crl-default",
		ResourceVersion:   rv,
		AsyncOperationId:  asyncOpID,
	}
}

// Cases where the rule is found and the user confirms, so the Delete RPC is sent.
func TestDeleteCallerRateLimit(t *testing.T) {
	tests := []struct {
		name            string
		flags           callerRateLimitDeleteFlags
		expectedRequest *cloudservice.DeleteCallerRateLimitRequest
		deleteErr       error
		pollerOptions   temporalcloudcli.TestAsyncPollerOptions
		expectedErr     string
	}{
		{
			name:            "Success",
			flags:           callerRateLimitDeleteFlags{isDefault: true},
			expectedRequest: deleteCallerRateLimitRequest("rv-default", ""),
			pollerOptions:   temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name: "FlagOverrides",
			flags: callerRateLimitDeleteFlags{
				isDefault:        true,
				resourceVersion:  "rv-override",
				asyncOperationID: "op-custom",
			},
			expectedRequest: deleteCallerRateLimitRequest("rv-override", "op-custom"),
			pollerOptions:   temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: "op-del"},
		},
		{
			name:            "DeleteError",
			flags:           callerRateLimitDeleteFlags{isDefault: true},
			expectedRequest: deleteCallerRateLimitRequest("rv-default", ""),
			deleteErr:       errors.New("api error"),
			expectedErr:     "delete operation failed: api error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.flags.command()
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					expectFindDefaultCallerRateLimit(c, []*namespacev1.CallerRateLimit{testCallerRateLimitDefault}, nil)
					c.EXPECT().
						DeleteCallerRateLimit(mock.Anything, tt.expectedRequest, mock.Anything).
						Return(&cloudservice.DeleteCallerRateLimitResponse{
							AsyncOperation: &operation.AsyncOperation{Id: "op-del"},
						}, tt.deleteErr)
				},
				PromptOptions: temporalcloudcli.TestPromptOptions{
					ExpectPromptYes:        true,
					ExpectPromptYesMessage: "Delete",
					PromptResult:           true,
				},
				AsyncPollerOptions: tt.pollerOptions,
				ExpectedError:      tt.expectedErr,
			})
		})
	}
}

func TestDeleteCallerRateLimit_Declined(t *testing.T) {
	cmd := callerRateLimitDeleteFlags{isDefault: true}.command()
	// No DeleteCallerRateLimit expectation: the RPC must not be sent.
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			expectFindDefaultCallerRateLimit(c, []*namespacev1.CallerRateLimit{testCallerRateLimitDefault}, nil)
		},
		PromptOptions: temporalcloudcli.TestPromptOptions{
			ExpectPromptYes:        true,
			ExpectPromptYesMessage: "Delete",
			PromptResult:           false,
		},
		ExpectedError: "Aborting delete.",
	})
}

// Cases where the lookup does not yield a rule: no prompt and no Delete RPC.
// A missing rule is an error even with --idempotent.
func TestDeleteCallerRateLimit_LookupFails(t *testing.T) {
	tests := []struct {
		name        string
		flags       callerRateLimitDeleteFlags
		lookupErr   error
		expectedErr string
	}{
		{
			name:        "NotFound",
			flags:       callerRateLimitDeleteFlags{isDefault: true},
			expectedErr: `default caller rate limit not found on namespace "my-ns.acct"`,
		},
		{
			name:        "NotFoundIdempotent",
			flags:       callerRateLimitDeleteFlags{isDefault: true, idempotent: true},
			expectedErr: `default caller rate limit not found on namespace "my-ns.acct"`,
		},
		{
			name:        "LookupError",
			flags:       callerRateLimitDeleteFlags{isDefault: true},
			lookupErr:   errors.New("api error"),
			expectedErr: "api error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.flags.command()
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					expectFindDefaultCallerRateLimit(c, nil, tt.lookupErr)
				},
				ExpectedError: tt.expectedErr,
			})
		})
	}
}

func TestDeleteCallerRateLimit_NeitherFlagSet(t *testing.T) {
	cmd := callerRateLimitDeleteFlags{}.command()
	// No CloudClientExpectations or PromptOptions: any RPC or prompt fails the test.
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		ExpectedError: "either --caller-namespace-default or --caller-namespace is required",
	})
}

// createOp expects one Create RPC for spec that returns async operation pollID.
func createOp(spec *namespacev1.CallerRateLimitSpec, asyncOpID, pollID string, err error) mutationExpectation {
	return func(c *cloudmock.MockCloudServiceClient) *mock.Call {
		return c.EXPECT().
			CreateCallerRateLimit(mock.Anything, &cloudservice.CreateCallerRateLimitRequest{
				Namespace:        "my-ns.acct",
				Spec:             spec,
				AsyncOperationId: asyncOpID,
			}, mock.Anything).
			Return(&cloudservice.CreateCallerRateLimitResponse{
				AsyncOperation: &operation.AsyncOperation{Id: pollID},
			}, err).
			Call
	}
}

// updateOp expects one Update RPC of existing to spec that returns async operation pollID.
func updateOp(
	existing *namespacev1.CallerRateLimit,
	spec *namespacev1.CallerRateLimitSpec,
	rv, asyncOpID, pollID string,
	err error,
) mutationExpectation {
	return func(c *cloudmock.MockCloudServiceClient) *mock.Call {
		return c.EXPECT().
			UpdateCallerRateLimit(mock.Anything, &cloudservice.UpdateCallerRateLimitRequest{
				Namespace:         "my-ns.acct",
				CallerRateLimitId: existing.Id,
				Spec:              spec,
				ResourceVersion:   rv,
				AsyncOperationId:  asyncOpID,
			}, mock.Anything).
			Return(&cloudservice.UpdateCallerRateLimitResponse{
				AsyncOperation: &operation.AsyncOperation{Id: pollID},
			}, err).
			Call
	}
}

func (f callerRateLimitSetFlags) command() temporalcloudcli.CloudNamespaceCallerRateLimitSetCommand {
	return temporalcloudcli.CloudNamespaceCallerRateLimitSetCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
		CallerRateLimitCallerOptions: temporalcloudcli.CallerRateLimitCallerOptions{
			CallerNamespaceDefault: f.isDefault,
			CallerNamespace:        f.caller,
		},
		AsyncOperationOptions:  temporalcloudcli.AsyncOperationOptions{AsyncOperationId: f.asyncOperationID},
		ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: f.resourceVersion},
		Rps:                    f.rps,
	}
}

// Cases where no rule exists and the user confirms, so the Create RPC is sent.
func TestSetCallerRateLimit_Create(t *testing.T) {
	// Typed nil: PromptApply receives a nil *CallerRateLimitSpec, not a nil interface.
	noRule := (*namespacev1.CallerRateLimitSpec)(nil)
	tests := []struct {
		name         string
		flags        callerRateLimitSetFlags
		expectedSpec *namespacev1.CallerRateLimitSpec
		mutationErr  error
		pollID       string
		expectedErr  string
	}{
		{
			name:         "DefaultRuleRpsTenth",
			flags:        callerRateLimitSetFlags{isDefault: true, rps: 0.1},
			expectedSpec: defaultRuleSpec(0.1),
			pollID:       "op-create",
		},
		{
			name:         "CallerRule",
			flags:        callerRateLimitSetFlags{caller: "caller-ns.acct", rps: 10},
			expectedSpec: callerRuleSpec("caller-ns.acct", 10),
			pollID:       "op-create",
		},
		{
			name: "AsyncOperationIDSetResourceVersionIgnored",
			flags: callerRateLimitSetFlags{
				isDefault:        true,
				rps:              100,
				resourceVersion:  "rv-ignored",
				asyncOperationID: "op-custom",
			},
			expectedSpec: defaultRuleSpec(100),
			pollID:       "op-create",
		},
		{
			name:         "CreateError",
			flags:        callerRateLimitSetFlags{isDefault: true, rps: 100},
			expectedSpec: defaultRuleSpec(100),
			mutationErr:  errors.New("api error"),
			expectedErr:  "create operation failed: api error",
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
						Return(&cloudservice.GetCallerRateLimitsResponse{}, nil)
					createOp(tt.expectedSpec, tt.flags.asyncOperationID, "op-create", tt.mutationErr)(c)
				},
				PromptOptions:      promptApply(noRule, tt.expectedSpec, false, true),
				AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: tt.pollID},
				ExpectedError:      tt.expectedErr,
			})
		})
	}
}

// Cases where the rule exists and the user confirms, so the Update RPC is sent.
func TestSetCallerRateLimit_Update(t *testing.T) {
	tests := []struct {
		name         string
		flags        callerRateLimitSetFlags
		found        *namespacev1.CallerRateLimit
		expectedSpec *namespacev1.CallerRateLimitSpec
		expectedRV   string
		mutationErr  error
		pollID       string
		expectedErr  string
	}{
		{
			name:         "DefaultRule",
			flags:        callerRateLimitSetFlags{isDefault: true, rps: 200},
			found:        testCallerRateLimitDefault,
			expectedSpec: defaultRuleSpec(200),
			expectedRV:   "rv-default",
			pollID:       "op-update",
		},
		{
			name: "CallerRuleFlagOverrides",
			flags: callerRateLimitSetFlags{
				caller:           "caller-ns.acct",
				rps:              20,
				resourceVersion:  "rv-override",
				asyncOperationID: "op-custom",
			},
			found:        testCallerRateLimitCaller,
			expectedSpec: callerRuleSpec("caller-ns.acct", 20),
			expectedRV:   "rv-override",
			pollID:       "op-update",
		},
		{
			name:         "UpdateError",
			flags:        callerRateLimitSetFlags{isDefault: true, rps: 200},
			found:        testCallerRateLimitDefault,
			expectedSpec: defaultRuleSpec(200),
			expectedRV:   "rv-default",
			mutationErr:  errors.New("api error"),
			expectedErr:  "update operation failed: api error",
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
						Return(&cloudservice.GetCallerRateLimitsResponse{
							CallerRateLimits: []*namespacev1.CallerRateLimit{tt.found},
						}, nil)
					updateOp(tt.found, tt.expectedSpec, tt.expectedRV, tt.flags.asyncOperationID, "op-update",
						tt.mutationErr)(c)
				},
				PromptOptions:      promptApply(tt.found.Spec, tt.expectedSpec, false, true),
				AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: tt.pollID},
				ExpectedError:      tt.expectedErr,
			})
		})
	}
}

func TestSetCallerRateLimit_Declined(t *testing.T) {
	cmd := callerRateLimitSetFlags{isDefault: true, rps: 200}.command()
	// No Create/Update expectation: neither RPC may be sent.
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			expectFindDefaultCallerRateLimit(c, []*namespacev1.CallerRateLimit{testCallerRateLimitDefault}, nil)
		},
		PromptOptions: promptApply(testCallerRateLimitDefault.Spec, defaultRuleSpec(200), false, false),
		ExpectedError: "Aborting set.",
	})
}

func TestSetCallerRateLimit_LookupFails(t *testing.T) {
	cmd := callerRateLimitSetFlags{isDefault: true, rps: 200}.command()
	// No PromptOptions: a prompt fails the test.
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			expectFindDefaultCallerRateLimit(c, nil, errors.New("api error"))
		},
		ExpectedError: "api error",
	})
}

func TestSetCallerRateLimit_NeitherFlagSet(t *testing.T) {
	cmd := callerRateLimitSetFlags{rps: 200}.command()
	// No CloudClientExpectations or PromptOptions: any RPC or prompt fails the test.
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		ExpectedError: "either --caller-namespace-default or --caller-namespace is required",
	})
}

func (f callerRateLimitApplyFlags) command() temporalcloudcli.CloudNamespaceCallerRateLimitApplyCommand {
	return temporalcloudcli.CloudNamespaceCallerRateLimitApplyCommand{
		NamespaceOptions: temporalcloudcli.NamespaceOptions{Namespace: "my-ns.acct"},
		AsyncOperationOptions: temporalcloudcli.AsyncOperationOptions{
			AsyncOperationId: f.asyncOperationID,
			Async:            f.async,
		},
		ResourceVersionOptions: temporalcloudcli.ResourceVersionOptions{ResourceVersion: f.resourceVersion},
		DiffOptions:            temporalcloudcli.DiffOptions{VerboseDiff: f.verboseDiff},
		Spec:                   f.spec,
	}
}

// expectListAllCallerRateLimits expects one unfiltered GetCallerRateLimits call per page,
// chained by page tokens.
func expectListAllCallerRateLimits(c *cloudmock.MockCloudServiceClient, pages ...[]*namespacev1.CallerRateLimit) {
	tokens := make([]string, len(pages)+1) // tokens[0] and tokens[len(pages)] stay ""
	for i := 1; i < len(pages); i++ {
		tokens[i] = fmt.Sprintf("page-%d", i)
	}
	for i, page := range pages {
		c.EXPECT().
			GetCallerRateLimits(mock.Anything, &cloudservice.GetCallerRateLimitsRequest{
				Namespace: "my-ns.acct",
				PageToken: tokens[i],
			}, mock.Anything).
			Return(&cloudservice.GetCallerRateLimitsResponse{CallerRateLimits: page, NextPageToken: tokens[i+1]}, nil)
	}
}

// callerRateLimitDiffList is the PromptApply side for specs: spec-only rules in a response.
func callerRateLimitDiffList(specs ...*namespacev1.CallerRateLimitSpec) *cloudservice.GetCallerRateLimitsResponse {
	res := &cloudservice.GetCallerRateLimitsResponse{}
	for _, spec := range specs {
		res.CallerRateLimits = append(res.CallerRateLimits, &namespacev1.CallerRateLimit{Spec: spec})
	}
	return res
}

// Cases where the user confirms the diff, so the changed rules are sent in key order.
// Ops after a failing op have no expectation, so sending them fails the test.
func TestApplyCallerRateLimits(t *testing.T) {
	existingPage := []*namespacev1.CallerRateLimit{testCallerRateLimitDefault, testCallerRateLimitCaller}
	newCaller := callerRuleSpec("new-ns.acct", 5)
	callerUpdated := callerRuleSpec("caller-ns.acct", 20)
	updateFailedDefault := defaultCallerRateLimit(100)
	updateFailedDefault.State = resourcev1.ResourceState_RESOURCE_STATE_UPDATE_FAILED
	tests := []struct {
		name           string
		flags          callerRateLimitApplyFlags
		pages          [][]*namespacev1.CallerRateLimit
		expectedBefore *cloudservice.GetCallerRateLimitsResponse
		expectedAfter  *cloudservice.GetCallerRateLimitsResponse
		ops            []mutationExpectation
		pollID         string
		expectedErr    string
	}{
		{
			name:           "UpdateVerboseDiff",
			flags:          callerRateLimitApplyFlags{spec: `[{"namespaceDefault":{"rps":200}}]`, verboseDiff: true},
			pages:          [][]*namespacev1.CallerRateLimit{existingPage},
			expectedBefore: callerRateLimitDiffList(testCallerRateLimitDefault.Spec),
			expectedAfter:  callerRateLimitDiffList(defaultRuleSpec(200)),
			ops: []mutationExpectation{
				updateOp(testCallerRateLimitDefault, defaultRuleSpec(200), "rv-default", "", "op-1", nil),
			},
			pollID: "op-1",
		},
		{
			// The unchanged caller rule is on the second page, so it is only skipped if lookup pages.
			name: "UnchangedRuleOnLaterPageSkipped",
			flags: callerRateLimitApplyFlags{
				spec: `[{"namespaceLimit":{"namespaceId":"caller-ns.acct","rps":10}},{"namespaceDefault":{"rps":200}}]`,
			},
			pages: [][]*namespacev1.CallerRateLimit{
				{testCallerRateLimitDefault},
				{testCallerRateLimitCaller},
			},
			expectedBefore: callerRateLimitDiffList(testCallerRateLimitDefault.Spec),
			expectedAfter:  callerRateLimitDiffList(defaultRuleSpec(200)),
			ops: []mutationExpectation{
				updateOp(testCallerRateLimitDefault, defaultRuleSpec(200), "rv-default", "", "op-1", nil),
			},
			pollID: "op-1",
		},
		{
			name: "SingleUpdateFlagOverrides",
			flags: callerRateLimitApplyFlags{
				spec:             `[{"namespaceLimit":{"namespaceId":"caller-ns.acct","rps":20}}]`,
				resourceVersion:  "rv-override",
				asyncOperationID: "op-custom",
			},
			pages:          [][]*namespacev1.CallerRateLimit{existingPage},
			expectedBefore: callerRateLimitDiffList(testCallerRateLimitCaller.Spec),
			expectedAfter:  callerRateLimitDiffList(callerUpdated),
			ops: []mutationExpectation{
				updateOp(testCallerRateLimitCaller, callerUpdated, "rv-override", "op-custom", "op-1", nil),
			},
			pollID: "op-1",
		},
		{
			name: "SingleCreateAsyncOperationIDSetResourceVersionIgnored",
			flags: callerRateLimitApplyFlags{
				spec:             `[{"namespaceLimit":{"namespaceId":"new-ns.acct","rps":5}}]`,
				resourceVersion:  "rv-ignored",
				asyncOperationID: "op-custom",
			},
			pages:          [][]*namespacev1.CallerRateLimit{existingPage},
			expectedBefore: callerRateLimitDiffList(),
			expectedAfter:  callerRateLimitDiffList(newCaller),
			ops:            []mutationExpectation{createOp(newCaller, "op-custom", "op-1", nil)},
			pollID:         "op-1",
		},
		{
			// The existing rule matches the file but its last update failed, so it is updated again.
			name:           "UnchangedButUpdateFailedRetried",
			flags:          callerRateLimitApplyFlags{spec: `[{"namespaceDefault":{"rps":100}}]`},
			pages:          [][]*namespacev1.CallerRateLimit{{updateFailedDefault}},
			expectedBefore: callerRateLimitDiffList(defaultRuleSpec(100)),
			expectedAfter:  callerRateLimitDiffList(defaultRuleSpec(100)),
			ops: []mutationExpectation{
				updateOp(updateFailedDefault, defaultRuleSpec(100), "rv-default", "", "op-1", nil),
			},
			pollID: "op-1",
		},
		{
			// Existing rules with no rule (a caller type this CLI doesn't recognize) are skipped,
			// so two of them are not reported as duplicates.
			name:  "UnrecognizedExistingRulesSkipped",
			flags: callerRateLimitApplyFlags{spec: `[{"namespaceDefault":{"rps":200}}]`},
			pages: [][]*namespacev1.CallerRateLimit{{
				{Id: "crl-unknown-1", Spec: &namespacev1.CallerRateLimitSpec{}},
				{Id: "crl-unknown-2", Spec: &namespacev1.CallerRateLimitSpec{}},
				testCallerRateLimitDefault,
			}},
			expectedBefore: callerRateLimitDiffList(testCallerRateLimitDefault.Spec),
			expectedAfter:  callerRateLimitDiffList(defaultRuleSpec(200)),
			ops: []mutationExpectation{
				updateOp(testCallerRateLimitDefault, defaultRuleSpec(200), "rv-default", "", "op-1", nil),
			},
			pollID: "op-1",
		},
		{
			// File order is the reverse of key order, so ops run default, caller, new-ns.
			// The second op fails: the error names it and the first as applied; the third is never sent.
			name: "FailFastAfterFirstOpInKeyOrder",
			flags: callerRateLimitApplyFlags{spec: `[
				{"namespaceLimit":{"namespaceId":"new-ns.acct","rps":5}},
				{"namespaceLimit":{"namespaceId":"caller-ns.acct","rps":20}},
				{"namespaceDefault":{"rps":200}}
			]`},
			pages:          [][]*namespacev1.CallerRateLimit{existingPage},
			expectedBefore: callerRateLimitDiffList(testCallerRateLimitDefault.Spec, testCallerRateLimitCaller.Spec),
			expectedAfter:  callerRateLimitDiffList(defaultRuleSpec(200), callerUpdated, newCaller),
			ops: []mutationExpectation{
				updateOp(testCallerRateLimitDefault, defaultRuleSpec(200), "rv-default", "", "op-1", nil),
				updateOp(testCallerRateLimitCaller, callerUpdated, "rv-caller", "", "op-2", errors.New("api error")),
			},
			pollID: "op-1",
			expectedErr: `failed to apply caller rate limit for caller namespace "caller-ns.acct": ` +
				`update operation failed: api error (already applied: default caller rate limit)`,
		},
		{
			// With --async, earlier rules were only submitted, so the error says so.
			// The harness poller always waits, so the first op is still polled.
			name: "FailFastAsyncReportsSubmitted",
			flags: callerRateLimitApplyFlags{
				spec:  `[{"namespaceLimit":{"namespaceId":"caller-ns.acct","rps":20}},{"namespaceDefault":{"rps":200}}]`,
				async: true,
			},
			pages:          [][]*namespacev1.CallerRateLimit{existingPage},
			expectedBefore: callerRateLimitDiffList(testCallerRateLimitDefault.Spec, testCallerRateLimitCaller.Spec),
			expectedAfter:  callerRateLimitDiffList(defaultRuleSpec(200), callerUpdated),
			ops: []mutationExpectation{
				updateOp(testCallerRateLimitDefault, defaultRuleSpec(200), "rv-default", "", "op-1", nil),
				updateOp(testCallerRateLimitCaller, callerUpdated, "rv-caller", "", "op-2", errors.New("api error")),
			},
			pollID: "op-1",
			expectedErr: `failed to apply caller rate limit for caller namespace "caller-ns.acct": ` +
				`update operation failed: api error (already submitted: default caller rate limit)`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.flags.command()
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					expectListAllCallerRateLimits(c, tt.pages...)
					calls := make([]*mock.Call, len(tt.ops))
					for i, op := range tt.ops {
						calls[i] = op(c)
					}
					mock.InOrder(calls...)
				},
				PromptOptions:      promptApply(tt.expectedBefore, tt.expectedAfter, tt.flags.verboseDiff, true),
				AsyncPollerOptions: temporalcloudcli.TestAsyncPollerOptions{AsyncOperationID: tt.pollID},
				ExpectedError:      tt.expectedErr,
			})
		})
	}
}

// Cases where nothing changes: a no-op message is printed with no prompt and no mutation.
func TestApplyCallerRateLimits_NoChanges(t *testing.T) {
	tests := []struct {
		name string
		spec string
	}{
		{
			name: "AllUnchanged",
			spec: `[{"namespaceLimit":{"namespaceId":"caller-ns.acct","rps":10}},{"namespaceDefault":{"rps":100}}]`,
		},
		{name: "EmptyArray", spec: `[]`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := callerRateLimitApplyFlags{spec: tt.spec}.command()
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					expectListAllCallerRateLimits(c,
						[]*namespacev1.CallerRateLimit{testCallerRateLimitDefault, testCallerRateLimitCaller})
				},
				ExpectedOutput: "No changes to apply: all rules in --spec already match.\n",
			})
		})
	}
}

// Cases rejected while parsing --spec, before any RPC.
func TestApplyCallerRateLimits_InvalidSpec(t *testing.T) {
	tests := []struct {
		name        string
		spec        string
		expectedErr string
	}{
		{
			name:        "NotAnArray",
			spec:        `{"namespaceDefault":{"rps":1}}`,
			expectedErr: "failed to parse JSON spec: expected an array",
		},
		{
			name:        "InvalidEntry",
			spec:        `[{"namespaceDefault":{"rps":1}},{"namespaceDefault":{"rps":"fast"}}]`,
			expectedErr: "failed to parse JSON spec entry 2",
		},
		{
			name: "DuplicateDefault",
			spec: `[{"namespaceDefault":{"rps":1}},{"namespaceLimit":{"namespaceId":"caller-ns.acct","rps":1}},` +
				`{"namespaceDefault":{"rps":2}}]`,
			expectedErr: "duplicate default caller rate limit in spec (entries 1 and 3)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := callerRateLimitApplyFlags{spec: tt.spec}.command()
			// No CloudClientExpectations or PromptOptions: any RPC or prompt fails the test.
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{ExpectedError: tt.expectedErr})
		})
	}
}

// No prompt and no mutation when the lookup fails.
func TestApplyCallerRateLimits_LookupFails(t *testing.T) {
	cmd := callerRateLimitApplyFlags{spec: `[{"namespaceDefault":{"rps":1}}]`}.command()
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			c.EXPECT().
				GetCallerRateLimits(mock.Anything,
					&cloudservice.GetCallerRateLimitsRequest{Namespace: "my-ns.acct"}, mock.Anything).
				Return(nil, errors.New("api error"))
		},
		ExpectedError: "api error",
	})
}

// With more than one changed rule, --async-operation-id and --resource-version are rejected
// before the prompt.
func TestApplyCallerRateLimits_FlagsWithMultipleChanges(t *testing.T) {
	spec := `[{"namespaceDefault":{"rps":200}},{"namespaceLimit":{"namespaceId":"new-ns.acct","rps":5}}]`
	tests := []struct {
		name  string
		flags callerRateLimitApplyFlags
	}{
		{name: "AsyncOperationID", flags: callerRateLimitApplyFlags{spec: spec, asyncOperationID: "op-custom"}},
		{name: "ResourceVersion", flags: callerRateLimitApplyFlags{spec: spec, resourceVersion: "rv-override"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := tt.flags.command()
			temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
				CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
					expectListAllCallerRateLimits(c, []*namespacev1.CallerRateLimit{testCallerRateLimitDefault})
				},
				ExpectedError: "--async-operation-id and --resource-version can only be used when apply " +
					"changes a single rule",
			})
		})
	}
}

func TestApplyCallerRateLimits_Declined(t *testing.T) {
	cmd := callerRateLimitApplyFlags{spec: `[{"namespaceDefault":{"rps":200}}]`}.command()
	// No Create/Update expectation: neither RPC may be sent.
	temporalcloudcli.TestCommand(t, &cmd, temporalcloudcli.TestCommandOptions{
		CloudClientExpectations: func(c *cloudmock.MockCloudServiceClient) {
			expectListAllCallerRateLimits(c, []*namespacev1.CallerRateLimit{testCallerRateLimitDefault})
		},
		PromptOptions: promptApply(
			callerRateLimitDiffList(testCallerRateLimitDefault.Spec),
			callerRateLimitDiffList(defaultRuleSpec(200)),
			false,
			false,
		),
		ExpectedError: "Aborting apply.",
	})
}
