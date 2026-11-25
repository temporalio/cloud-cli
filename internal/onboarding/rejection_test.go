package onboarding_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/temporalio/cloud-cli/internal/onboarding"
)

func rejection(t *testing.T, code codes.Code, domain, reason string) error {
	t.Helper()
	st, err := status.New(code, "refused").WithDetails(&errdetails.ErrorInfo{Domain: domain, Reason: reason})
	require.NoError(t, err)
	return st.Err()
}

func TestRejectionReason(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "signup rejection",
			err:      rejection(t, codes.FailedPrecondition, "cloud.temporal.io", onboarding.ReasonAlreadyOnboarded),
			expected: onboarding.ReasonAlreadyOnboarded,
		},
		{
			name: "wrapped signup rejection",
			err: fmt.Errorf("failed to get onboarding state: %w",
				rejection(t, codes.FailedPrecondition, "cloud.temporal.io", onboarding.ReasonTrialSignupCancelled)),
			expected: onboarding.ReasonTrialSignupCancelled,
		},
		{
			name: "other domain",
			err:  rejection(t, codes.FailedPrecondition, "example.com", onboarding.ReasonAlreadyOnboarded),
		},
		{
			name: "other code",
			err:  rejection(t, codes.PermissionDenied, "cloud.temporal.io", onboarding.ReasonAlreadyOnboarded),
		},
		{
			name: "no details",
			err:  status.Error(codes.FailedPrecondition, "refused"),
		},
		{
			name: "not a grpc error",
			err:  errors.New("boom"),
		},
		{
			name: "nil",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, onboarding.RejectionReason(tt.err))
		})
	}
}
