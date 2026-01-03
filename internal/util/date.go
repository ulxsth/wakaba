package util

import (
	"fmt"
	"strconv"
	"time"
)

// 日付文字列を解析して、開始日時と終了日時を返す
// MMDD 形式で渡された場合、現在年の日付を返す
// YYYYMMDD 形式で渡された場合、指定年の日付を返す
func ParseDateInput(input string, now time.Time) (time.Time, time.Time, error) {
	var year, month, day int
	var err error

	// JST location
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)

	if len(input) == 4 {
		// MMDD format
		year = now.In(jst).Year()
		month, err = strconv.Atoi(input[:2])
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid month: %w", err)
		}
		day, err = strconv.Atoi(input[2:])
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid day: %w", err)
		}
	} else if len(input) == 8 {
		// YYYYMMDD format
		year, err = strconv.Atoi(input[:4])
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid year: %w", err)
		}
		month, err = strconv.Atoi(input[4:6])
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid month: %w", err)
		}
		day, err = strconv.Atoi(input[6:])
		if err != nil {
			return time.Time{}, time.Time{}, fmt.Errorf("invalid day: %w", err)
		}

	} else {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid format, expected MMDD or YYYYMMDD")
	}

	startOfDay := time.Date(year, time.Month(month), day, 0, 0, 0, 0, jst)
	if startOfDay.Year() != year || startOfDay.Month() != time.Month(month) || startOfDay.Day() != day {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid date: %04d-%02d-%02d", year, month, day)
	}

	endOfDay := time.Date(year, time.Month(month), day, 23, 59, 59, 999999999, jst)

	return startOfDay, endOfDay, nil
}
