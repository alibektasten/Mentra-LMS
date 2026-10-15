package errors

import (
	"errors"
	"fmt"
	"net/http"
)

type AppError struct {
	Code       string            `json:"code"`
	Message    string            `json:"message"`
	Fields     map[string]string `json:"fields,omitempty"`
	HTTPStatus int               `json:"-"`
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func NotFoundError(resource, id string) *AppError {
	return &AppError{
		Code:       "NOT_FOUND",
		Message:    fmt.Sprintf("%s with id %s not found", resource, id),
		HTTPStatus: http.StatusNotFound,
	}
}

func ValidationError(msg string) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    msg,
		HTTPStatus: http.StatusUnprocessableEntity,
	}
}

func FieldValidationError(fields map[string]string) *AppError {
	return &AppError{
		Code:       "VALIDATION_ERROR",
		Message:    "request validation failed",
		Fields:     fields,
		HTTPStatus: http.StatusUnprocessableEntity,
	}
}

func UnauthorizedError(msg string) *AppError {
	return &AppError{
		Code:       "UNAUTHORIZED",
		Message:    msg,
		HTTPStatus: http.StatusUnauthorized,
	}
}

func ForbiddenError(msg string) *AppError {
	return &AppError{
		Code:       "FORBIDDEN",
		Message:    msg,
		HTTPStatus: http.StatusForbidden,
	}
}

func ConflictError(resource string) *AppError {
	return &AppError{
		Code:       "CONFLICT",
		Message:    fmt.Sprintf("%s already exists", resource),
		HTTPStatus: http.StatusConflict,
	}
}

func TooManyRequestsError(msg string) *AppError {
	return &AppError{Code: "AI_DAILY_QUOTA_EXCEEDED", Message: msg, HTTPStatus: http.StatusTooManyRequests}
}

func InternalError(msg string) *AppError {
	return &AppError{
		Code:       "INTERNAL_ERROR",
		Message:    msg,
		HTTPStatus: http.StatusInternalServerError,
	}
}

func IsNotFound(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.HTTPStatus == http.StatusNotFound
}

func IsValidation(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.HTTPStatus == http.StatusUnprocessableEntity
}

func IsUnauthorized(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.HTTPStatus == http.StatusUnauthorized
}

func IsForbidden(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr) && appErr.HTTPStatus == http.StatusForbidden
}
