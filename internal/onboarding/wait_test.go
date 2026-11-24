package onboarding_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/temporalio/cloud-cli/internal/onboarding"
	onboardingservice "github.com/temporalio/cloud-cli/internal/onboardingservice/v1"
)

type result struct {
	resp *onboardingservice.GetOrCreateCLIOnboardingStateResponse
	err  error
}

// fakeClient returns scripted results in order, repeating the last one.
type fakeClient struct {
	results []result
	calls   int
}

func (f *fakeClient) GetOrCreateCLIOnboardingState(
	context.Context, *onboardingservice.GetOrCreateCLIOnboardingStateRequest, ...grpc.CallOption,
) (*onboardingservice.GetOrCreateCLIOnboardingStateResponse, error) {
	r := f.results[min(f.calls, len(f.results)-1)]
	f.calls++
	return r.resp, r.err
}

func atStep(step onboardingservice.StepType) result {
	return result{resp: &onboardingservice.GetOrCreateCLIOnboardingStateResponse{
		State: &onboardingservice.OnboardingState{CurrentStep: step},
	}}
}

var fastPoll = onboarding.PollConfig{Interval: time.Millisecond, MaxInterval: 2 * time.Millisecond, FastAttempts: 2}

func TestWait(t *testing.T) {
	alreadyExists := result{resp: &onboardingservice.GetOrCreateCLIOnboardingStateResponse{UserAlreadyExists: true}}
	done := atStep(onboardingservice.StepType_STEP_TYPE_END_OF_FUNNEL)
	creating := atStep(onboardingservice.StepType_STEP_TYPE_WAIT_FOR_ACCOUNT_CREATION)

	tests := []struct {
		name          string
		results       []result
		timeout       time.Duration
		expectedResp  *onboardingservice.GetOrCreateCLIOnboardingStateResponse
		expectedErr   string
		expectedCalls int
		expectedSteps []onboardingservice.StepType
	}{
		{
			name:          "user already exists returns immediately",
			results:       []result{alreadyExists},
			expectedResp:  alreadyExists.resp,
			expectedCalls: 1,
		},
		{
			name:          "polls until end of funnel and reports each step once",
			results:       []result{creating, creating, creating, done},
			expectedResp:  done.resp,
			expectedCalls: 4,
			expectedSteps: []onboardingservice.StepType{onboardingservice.StepType_STEP_TYPE_WAIT_FOR_ACCOUNT_CREATION},
		},
		{
			name: "retries unavailable and rate limited",
			results: []result{
				{err: status.Error(codes.Unavailable, "down")},
				{err: status.Error(codes.ResourceExhausted, "slow down")},
				done,
			},
			expectedResp:  done.resp,
			expectedCalls: 3,
		},
		{
			name:          "stops on permission denied",
			results:       []result{{err: status.Error(codes.PermissionDenied, "user not allowed")}},
			expectedErr:   "failed to get onboarding state: rpc error: code = PermissionDenied desc = user not allowed",
			expectedCalls: 1,
		},
		{
			name:          "times out with the last step",
			results:       []result{creating},
			timeout:       20 * time.Millisecond,
			expectedErr:   "timed out waiting for onboarding to complete (last step: STEP_TYPE_WAIT_FOR_ACCOUNT_CREATION)",
			expectedSteps: []onboardingservice.StepType{onboardingservice.StepType_STEP_TYPE_WAIT_FOR_ACCOUNT_CREATION},
		},
		{
			name:        "times out before any state",
			results:     []result{{err: status.Error(codes.Unavailable, "down")}},
			timeout:     20 * time.Millisecond,
			expectedErr: "timed out waiting for onboarding to start",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, tt.timeout)
				defer cancel()
			}
			client := &fakeClient{results: tt.results}
			var steps []onboardingservice.StepType

			resp, err := onboarding.Wait(ctx, client, &onboardingservice.GetOrCreateCLIOnboardingStateRequest{}, fastPoll,
				func(s onboardingservice.StepType) { steps = append(steps, s) })

			if tt.expectedErr != "" {
				require.EqualError(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResp, resp)
			}
			if tt.expectedCalls > 0 {
				assert.Equal(t, tt.expectedCalls, client.calls)
			}
			assert.Equal(t, tt.expectedSteps, steps)
		})
	}
}

func TestWait_Canceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := onboarding.Wait(ctx, &fakeClient{results: []result{atStep(onboardingservice.StepType_STEP_TYPE_UNSPECIFIED)}},
		&onboardingservice.GetOrCreateCLIOnboardingStateRequest{}, fastPoll, nil)

	require.True(t, errors.Is(err, context.Canceled))
}
