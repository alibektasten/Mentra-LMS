package usecases

import (
	"errors"
	"net/http"

	apperrors "github.com/ailms/backend/pkg/errors"
)

func isAIBudgetError(err error) bool {
	var appErr *apperrors.AppError
	return errors.As(err, &appErr) && (appErr.HTTPStatus == http.StatusTooManyRequests || appErr.HTTPStatus == http.StatusUnauthorized || appErr.HTTPStatus == http.StatusInternalServerError)
}
