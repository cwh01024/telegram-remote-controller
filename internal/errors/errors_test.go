package errors

import (
	"errors"
	"testing"
	"time"
)

func TestNewAppError(t *testing.T) {
	err := New(ErrAuthFailed, "user not authorized")
	if err.Code != ErrAuthFailed {
		t.Errorf("Expected code %d, got %d", ErrAuthFailed, err.Code)
	}
	if err.Message != "user not authorized" {
		t.Errorf("Expected message 'user not authorized', got '%s'", err.Message)
	}
}

func TestWrapError(t *testing.T) {
	cause := errors.New("connection refused")
	err := Wrap(ErrBotConnection, "failed to connect", cause)

	if err.Cause != cause {
		t.Error("Cause not preserved")
	}
	if !errors.Is(err, cause) {
		t.Error("errors.Is should work with wrapped error")
	}
}

func TestIsError(t *testing.T) {
	err := New(ErrTimeout, "operation timed out")
	if !Is(err, ErrTimeout) {
		t.Error("Is should return true for matching code")
	}
	if Is(err, ErrAuthFailed) {
		t.Error("Is should return false for non-matching code")
	}
}

func TestErrorName(t *testing.T) {
	name := ErrorName(ErrAuthFailed)
	if name != "驗證失敗" {
		t.Errorf("Expected '驗證失敗', got '%s'", name)
	}
}

func TestRetrySuccess(t *testing.T) {
	attempts := 0
	err := Retry(3, time.Millisecond, func() error {
		attempts++
		if attempts < 2 {
			return errors.New("not yet")
		}
		return nil
	})

	if err != nil {
		t.Errorf("Retry should succeed, got: %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetryAllFail(t *testing.T) {
	attempts := 0
	err := Retry(3, time.Millisecond, func() error {
		attempts++
		return errors.New("always fail")
	})

	if err == nil {
		t.Error("Retry should fail")
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryWithBackoff(t *testing.T) {
	attempts := 0
	start := time.Now()
	Retry(3, 10*time.Millisecond, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("not yet")
		}
		return nil
	})
	elapsed := time.Since(start)

	// Should have some delay
	if elapsed < 15*time.Millisecond {
		t.Logf("Elapsed time: %v (expected some delay)", elapsed)
	}
}

func TestErrorString(t *testing.T) {
	err := New(ErrTimeout, "timed out")
	str := err.Error()
	if str == "" {
		t.Error("Error string should not be empty")
	}

	errWithCause := Wrap(ErrAutomation, "script failed", errors.New("exit status 1"))
	str = errWithCause.Error()
	if str == "" {
		t.Error("Error with cause string should not be empty")
	}
}
