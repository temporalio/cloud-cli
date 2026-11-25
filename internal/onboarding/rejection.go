package onboarding

import (
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Reasons the onboarding service gives, as a google.rpc.ErrorInfo reason on a
// FailedPrecondition, for refusing a CLI signup.
const (
	// ReasonAlreadyOnboarded: the user already had an account before this signup.
	ReasonAlreadyOnboarded = "ALREADY_ONBOARDED"
	// ReasonTrialSignupUnavailable: CLI signup is turned off.
	ReasonTrialSignupUnavailable = "TRIAL_SIGNUP_UNAVAILABLE"
	// ReasonBrowserSignupInProgress: the user has a signup running in the browser.
	ReasonBrowserSignupInProgress = "BROWSER_SIGNUP_IN_PROGRESS"
	// ReasonTrialSignupCancelled: the user's signup was cancelled.
	ReasonTrialSignupCancelled = "TRIAL_SIGNUP_CANCELLED"
)

const errorInfoDomain = "cloud.temporal.io"

// RejectionReason returns the ErrorInfo reason of a signup refused by the
// onboarding service, or "" when err is not one. err may be wrapped.
func RejectionReason(err error) string {
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.FailedPrecondition {
		return ""
	}
	for _, detail := range st.Details() {
		if info, ok := detail.(*errdetails.ErrorInfo); ok && info.GetDomain() == errorInfoDomain {
			return info.GetReason()
		}
	}
	return ""
}
