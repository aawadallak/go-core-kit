package common

import "errors"

type FailureMode string

const (
	FailureModeRecoverable    FailureMode = "recoverable"     // retry
	FailureModeNonRecoverable FailureMode = "non_recoverable" // DLQ/alert
	FailureModeDrop           FailureMode = "drop"            // ack and discard
	FailureModeUnknown        FailureMode = "unknown"         // fail-safe fallback
)

// FailureModeError is an optional contract for module-specific errors to explicitly define handling mode.
type FailureModeError interface {
	FailureMode() FailureMode
}

// ClassifyFailureMode resolves how a worker should handle an error.
// Priority:
// 1) Explicit mode from error implementation (FailureModeError)
// 2) Legacy fallback via existing error sentinels/types
// 3) Unknown (caller should treat as non_recoverable by default)
func ClassifyFailureMode(err error) FailureMode {
	if err == nil {
		return FailureModeDrop
	}

	// 1) Explicit mode
	var fm FailureModeError
	if errors.As(err, &fm) {
		mode := fm.FailureMode()
		if mode != "" {
			return mode
		}
		return FailureModeUnknown
	}

	// 2) Legacy fallback
	switch {
	case errors.Is(err, ErrParseRequest),
		errors.Is(err, ErrRequiredField),
		errors.Is(err, ErrRequiredFields):
		return FailureModeDrop
	case errors.Is(err, ErrInternalServer):
		return FailureModeRecoverable
	default:
		return FailureModeUnknown
	}
}
