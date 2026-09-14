// Package errcode 定义稳定错误码。前缀按域划分：COMMON_ / AUTH_ / IDEMPOTENCY_ / SM_。
package errcode

import "fmt"

// Code 是对外稳定错误码，写入响应信封的 code 字段。
type Code string

const (
	OK Code = "OK"

	CommonInvalidArgument Code = "COMMON_INVALID_ARGUMENT"
	CommonNotFound        Code = "COMMON_NOT_FOUND"
	CommonInternal        Code = "COMMON_INTERNAL"
	CommonNotImplemented  Code = "COMMON_NOT_IMPLEMENTED"
	CommonUnavailable     Code = "COMMON_UNAVAILABLE"
	CommonPayloadTooLarge Code = "COMMON_PAYLOAD_TOO_LARGE"
	CommonTimeout         Code = "COMMON_TIMEOUT"
	CommonConflict        Code = "COMMON_CONFLICT"

	AuthUnauthorized       Code = "AUTH_UNAUTHORIZED"
	AuthInvalidCredentials Code = "AUTH_INVALID_CREDENTIALS"
	AuthTokenExpired       Code = "AUTH_TOKEN_EXPIRED"
	AuthTokenInvalid       Code = "AUTH_TOKEN_INVALID"

	IdempotencyKeyRequired Code = "IDEMPOTENCY_KEY_REQUIRED"
	IdempotencyKeyConflict Code = "IDEMPOTENCY_KEY_CONFLICT"

	SMCryptoInvalid  Code = "SM_CRYPTO_INVALID"
	SMCryptoDisabled Code = "SM_CRYPTO_DISABLED"
)

// Error 是可映射到 HTTP 状态与信封的业务错误。
type Error struct {
	Code    Code
	HTTP    int
	Message string
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func New(code Code, httpStatus int, msg string) *Error {
	return &Error{Code: code, HTTP: httpStatus, Message: msg}
}

func InvalidArgument(msg string) *Error {
	return New(CommonInvalidArgument, 400, msg)
}

func NotFound(msg string) *Error {
	return New(CommonNotFound, 404, msg)
}

func Internal(msg string) *Error {
	return New(CommonInternal, 500, msg)
}

func NotImplemented(msg string) *Error {
	return New(CommonNotImplemented, 501, msg)
}

func Unavailable(msg string) *Error {
	return New(CommonUnavailable, 503, msg)
}

func Unauthorized(msg string) *Error {
	return New(AuthUnauthorized, 401, msg)
}

func InvalidCredentials() *Error {
	return New(AuthInvalidCredentials, 401, "invalid username or password")
}

func TokenExpired() *Error {
	return New(AuthTokenExpired, 401, "access token expired")
}

func TokenInvalid() *Error {
	return New(AuthTokenInvalid, 401, "access token invalid")
}

func IdempotencyRequired() *Error {
	return New(IdempotencyKeyRequired, 400, "Idempotency-Key header required")
}

func IdempotencyConflict(msg string) *Error {
	if msg == "" {
		msg = "idempotency key reused with different payload"
	}
	return New(IdempotencyKeyConflict, 409, msg)
}
