package perfperiod

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDays(t *testing.T) {

	// 2025-12-15
	baseTime := time.Date(2025, 12, 15, 13, 0, 0, 0, time.UTC)

	daysBefore, daysAfter, err := Days(Yesterday, baseTime)
	assert.Nil(t, err, "Yesterday err")
	assert.Equal(t, daysBefore, -1, "Yesterday daysBefore") // 2025-12-14
	assert.Equal(t, daysAfter, -1, "Yesterday daysAfter")   // 2025-12-14

	daysBefore, daysAfter, err = Days(Last3Days, baseTime)
	assert.Nil(t, err, "Last3Days err")
	assert.Equal(t, daysBefore, -2, "Last3Days daysBefore") // 2025-12-13
	assert.Equal(t, daysAfter, 0, "Last3Days daysAfter")    // 2025-12-15

	daysBefore, daysAfter, err = Days(ThisMonth, baseTime)
	assert.Nil(t, err, "ThisMonth err")
	assert.Equal(t, daysBefore, -14, "ThisMonth daysBefore") // 2025-12-01
	assert.Equal(t, daysAfter, 0, "ThisMonth daysAfter")     // 2025-12-15

	daysBefore, daysAfter, err = Days(LastMonth, baseTime)
	assert.Nil(t, err, "LastMonth err")
	assert.Equal(t, daysBefore, -44, "LastMonth daysBefore") // 2025-11-01
	assert.Equal(t, daysAfter, -15, "LastMonth daysAfter")   // 2025-11-30
}
