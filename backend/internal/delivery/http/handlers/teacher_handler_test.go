package handlers

import (
	"testing"
	"time"

	apperrors "github.com/ailms/backend/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalendarRange(t *testing.T) {
	now := time.Date(2026, time.October, 15, 12, 0, 0, 0, time.UTC)

	from, to, err := calendarRange("2026-10-05", "2026-10-11", now)
	require.NoError(t, err)
	assert.Equal(t, "2026-10-05", from.Format("2006-01-02"))
	assert.Equal(t, "2026-10-11", to.Format("2006-01-02"))

	from, to, err = calendarRange("", "", now)
	require.NoError(t, err)
	assert.Equal(t, "2026-10-01", from.Format("2006-01-02"))
	assert.Equal(t, "2026-10-31", to.Format("2006-01-02"))

	_, _, err = calendarRange("2026-10-11", "2026-10-05", now)
	require.Error(t, err)
	assert.True(t, apperrors.IsValidation(err))
}
