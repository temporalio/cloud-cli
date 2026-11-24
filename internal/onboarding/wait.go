// Package onboarding polls the Temporal Cloud onboarding service until a
// CLI-started sign-up is complete.
package onboarding

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	onboardingservice "github.com/temporalio/cloud-cli/internal/onboardingservice/v1"
)

// Client is the subset of the onboarding service the CLI calls.
type Client interface {
	GetOrCreateCLIOnboardingState(
		ctx context.Context,
		in *onboardingservice.GetOrCreateCLIOnboardingStateRequest,
		opts ...grpc.CallOption,
	) (*onboardingservice.GetOrCreateCLIOnboardingStateResponse, error)
}

// PollConfig controls the polling backoff. It mirrors the get-cloud website:
// a fixed interval for the first FastAttempts polls, then doubling up to
// MaxInterval.
type PollConfig struct {
	Interval     time.Duration
	MaxInterval  time.Duration
	FastAttempts int
}

// DefaultPollConfig matches the website's waitForStep: 1s for 10 polls, then
// doubling to a 30s cap.
var DefaultPollConfig = PollConfig{
	Interval:     time.Second,
	MaxInterval:  30 * time.Second,
	FastAttempts: 10,
}

// IsComplete reports whether sign-up has finished, either just now or earlier.
func IsComplete(resp *onboardingservice.GetOrCreateCLIOnboardingStateResponse) bool {
	return resp.GetUserAlreadyExists() ||
		resp.GetState().GetCurrentStep() == onboardingservice.StepType_STEP_TYPE_END_OF_FUNNEL
}

// Wait calls GetOrCreateCLIOnboardingState until sign-up is complete, ctx is
// done, or the service returns an error that is not worth retrying. The first
// call creates the onboarding session; the rest observe it. onStep, if set, is
// called each time the reported step changes.
func Wait(
	ctx context.Context,
	client Client,
	req *onboardingservice.GetOrCreateCLIOnboardingStateRequest,
	cfg PollConfig,
	onStep func(onboardingservice.StepType),
) (*onboardingservice.GetOrCreateCLIOnboardingStateResponse, error) {
	interval := cfg.Interval
	lastStep := onboardingservice.StepType(-1)

	for attempt := 1; ; attempt++ {
		resp, err := client.GetOrCreateCLIOnboardingState(ctx, req)
		switch {
		case ctx.Err() != nil:
			return nil, waitError(ctx, lastStep)
		case err == nil:
			if IsComplete(resp) {
				return resp, nil
			}
			if step := resp.GetState().GetCurrentStep(); step != lastStep {
				lastStep = step
				if onStep != nil {
					onStep(step)
				}
			}
		case !isRetryable(err):
			return nil, fmt.Errorf("failed to get onboarding state: %w", err)
		}

		if attempt > cfg.FastAttempts {
			interval = min(interval*2, cfg.MaxInterval)
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, waitError(ctx, lastStep)
		case <-timer.C:
		}
	}
}

// isRetryable covers transient transport failures and the service's per-email
// rate limit. Anything else (auth, allow list, invalid idp) will not change on
// retry.
func isRetryable(err error) bool {
	switch status.Code(err) {
	case codes.Unavailable, codes.ResourceExhausted:
		return true
	default:
		return false
	}
}

func waitError(ctx context.Context, lastStep onboardingservice.StepType) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		if lastStep < 0 {
			return errors.New("timed out waiting for onboarding to start")
		}
		return fmt.Errorf("timed out waiting for onboarding to complete (last step: %s)", lastStep)
	}
	return ctx.Err()
}
