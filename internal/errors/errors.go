package errors

import (
	"fmt"
	"time"
)

// ErrorCode represents different types of errors
type ErrorCode int

const (
	ErrUnknown ErrorCode = iota
	ErrBotConnection
	ErrAuthFailed
	ErrCommandParse
	ErrAutomation
	ErrIDENotReady
	ErrTimeout
	ErrScreenshot
)

// AppError is the application error type
type AppError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates a new AppError
func New(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

// Wrap wraps an error with an AppError
func Wrap(code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// Is checks if the error has a specific code
func Is(err error, code ErrorCode) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Code == code
	}
	return false
}

// ErrorName returns the human-readable name for an error code
func ErrorName(code ErrorCode) string {
	names := map[ErrorCode]string{
		ErrUnknown:       "未知錯誤",
		ErrBotConnection: "Bot 連接失敗",
		ErrAuthFailed:    "驗證失敗",
		ErrCommandParse:  "指令解析錯誤",
		ErrAutomation:    "自動化操作失敗",
		ErrIDENotReady:   "IDE 未就緒",
		ErrTimeout:       "操作超時",
		ErrScreenshot:    "截圖失敗",
	}
	if name, ok := names[code]; ok {
		return name
	}
	return "未知錯誤"
}

// Retry retries a function up to maxAttempts times with delay between attempts
func Retry(maxAttempts int, delay time.Duration, fn func() error) error {
	var lastErr error
	for i := 0; i < maxAttempts; i++ {
		if err := fn(); err != nil {
			lastErr = err
			if i < maxAttempts-1 {
				time.Sleep(delay)
			}
			continue
		}
		return nil
	}
	return Wrap(ErrUnknown, fmt.Sprintf("failed after %d attempts", maxAttempts), lastErr)
}

// RetryWithBackoff retries with exponential backoff
func RetryWithBackoff(maxAttempts int, initialDelay time.Duration, fn func() error) error {
	var lastErr error
	delay := initialDelay
	for i := 0; i < maxAttempts; i++ {
		if err := fn(); err != nil {
			lastErr = err
			if i < maxAttempts-1 {
				time.Sleep(delay)
				delay *= 2 // Exponential backoff
			}
			continue
		}
		return nil
	}
	return Wrap(ErrUnknown, fmt.Sprintf("failed after %d attempts with backoff", maxAttempts), lastErr)
}
