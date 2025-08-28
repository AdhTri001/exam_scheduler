package scheduler

import (
	"testing"
)

func TestGenerateSlots(t *testing.T) {
	startDate := "2025-01-06" // Monday
	endDate := "2025-01-10"   // Friday
	holidays := []string{"2025-01-08"}
	slotsPerDay := 2
	slotDuration := 180 // 3 hours
	slotTimes := []string{"09:00", "14:00"}

	slots, err := GenerateSlots(startDate, endDate, slotsPerDay, slotTimes, slotDuration, holidays, "UTC")
	if err != nil {
		t.Fatalf("GenerateSlots failed: %v", err)
	}

	// Expected slots: Mon(2), Tue(2), Thu(2), Fri(2) = 8 slots
	if len(slots) != 8 {
		t.Errorf("expected 8 slots, got %d", len(slots))
	}

	// Check if holiday was skipped
	for _, slot := range slots {
		if slot.Start.Day() == 8 {
			t.Errorf("slot generated on a holiday: %v", slot.Start)
		}
	}

	// Check slot ID determinism
	if slots[0].ID != "2025-01-06T09:00Z#1" {
		t.Errorf("unexpected slot ID for first slot: %s", slots[0].ID)
	}
	if slots[2].ID != "2025-01-07T09:00Z#1" {
		t.Errorf("unexpected slot ID for third slot: %s", slots[2].ID)
	}
}
