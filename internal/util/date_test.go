package util

import (
	"testing"
	"time"
)

func TestParseDateInput(t *testing.T) {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	// Mock "now" as 2024-01-01
	now := time.Date(2024, 1, 1, 0, 0, 0, 0, jst)

	tests := []struct {
		name      string
		input     string
		wantStart time.Time
		wantEnd   time.Time
		wantErr   bool
	}{
		{
			name:      "MMDD format (current year)",
			input:     "1225",
			wantStart: time.Date(2024, 12, 25, 0, 0, 0, 0, jst),
			wantEnd:   time.Date(2024, 12, 25, 23, 59, 59, 999999999, jst),
			wantErr:   false,
		},
		{
			name:      "YYYYMMDD format",
			input:     "20230101",
			wantStart: time.Date(2023, 1, 1, 0, 0, 0, 0, jst),
			wantEnd:   time.Date(2023, 1, 1, 23, 59, 59, 999999999, jst),
			wantErr:   false,
		},
		{
			name:    "Invalid length",
			input:   "123",
			wantErr: true,
		},
		{
			name:    "Invalid month",
			input:   "1301",
			wantErr: true,
		},
		{
			name:    "Invalid day",
			input:   "0230",
			wantErr: true,
		},
		{
			// 2024 is a leap year
			name:      "Leap year valid",
			input:     "0229",
			wantStart: time.Date(2024, 2, 29, 0, 0, 0, 0, jst),
			wantEnd:   time.Date(2024, 2, 29, 23, 59, 59, 999999999, jst),
			wantErr:   false,
		},
		{
			// 2023 is not a leap year
			name:    "Leap year invalid",
			input:   "20230229",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStart, gotEnd, err := ParseDateInput(tt.input, now)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDateInput() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !gotStart.Equal(tt.wantStart) {
					t.Errorf("ParseDateInput() gotStart = %v, want %v", gotStart, tt.wantStart)
				}
				if !gotEnd.Equal(tt.wantEnd) {
					t.Errorf("ParseDateInput() gotEnd = %v, want %v", gotEnd, tt.wantEnd)
				}
			}
		})
	}
}
