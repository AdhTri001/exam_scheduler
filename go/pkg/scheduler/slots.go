package scheduler

import (
	"fmt"
	"time"
)

// GenerateSlots creates a list of exam slots based on the provided parameters.
func GenerateSlots(startDate, endDate string, slotsPerDay int, slotTimes []string, slotDuration int, holidays []string, timezone string) ([]*Slot, error) {
	var slots []*Slot
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		// Fallback to UTC if timezone is invalid
		loc = time.UTC
	}

	start, err := time.ParseInLocation("2006-01-02", startDate, loc)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	end, err := time.ParseInLocation("2006-01-02", endDate, loc)
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	holidayMap := make(map[string]bool)
	for _, h := range holidays {
		holidayMap[h] = true
	}

	dayIndex := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		if holidayMap[dateStr] || d.Weekday() == time.Saturday || d.Weekday() == time.Sunday {
			continue
		}

		var timesToUse []time.Time
		if len(slotTimes) > 0 {
			for _, st := range slotTimes {
				t, err := time.Parse("15:04", st)
				if err != nil {
					return nil, fmt.Errorf("invalid slot time '%s': %w", st, err)
				}
				timesToUse = append(timesToUse, d.Add(time.Hour*time.Duration(t.Hour())+time.Minute*time.Duration(t.Minute())))
			}
		} else {
			// Generate evenly spaced times
			dayStart, _ := time.ParseInLocation("15:04", "09:00", loc) // Default start time
			for i := 0; i < slotsPerDay; i++ {
				timesToUse = append(timesToUse, d.Add(time.Hour*time.Duration(dayStart.Hour())+time.Minute*time.Duration(dayStart.Minute())).Add(time.Duration(i*slotDuration)*time.Minute))
			}
		}

		for i, t := range timesToUse {
			slotID := SlotID(fmt.Sprintf("%sT%sZ#%d", d.Format("2006-01-02"), t.Format("15:04"), i+1))
			slot := &Slot{
				ID:         slotID,
				Start:      t,
				End:        t.Add(time.Duration(slotDuration) * time.Minute),
				DayIndex:   dayIndex,
				IndexInDay: i,
			}
			slots = append(slots, slot)
		}
		dayIndex++
	}

	return slots, nil
}
