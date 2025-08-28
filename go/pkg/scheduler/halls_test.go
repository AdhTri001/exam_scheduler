package scheduler

import (
	"testing"
)

func TestAllocateHalls_SingleHall(t *testing.T) {
	assignments := []*Assignment{
		{CourseID: "c1", EnrolledCount: 80},
		{CourseID: "c2", EnrolledCount: 40},
	}
	halls := []*Hall{
		{ID: "H1", Capacity: 100},
		{ID: "H2", Capacity: 50},
	}
	usedHalls := make(map[SlotID]map[HallID]bool)
	slotID := SlotID("slot1")

	_, warnings, err := AllocateHalls(assignments, halls, usedHalls, slotID)
	if err != nil {
		t.Fatalf("AllocateHalls failed: %v", err)
	}
	if len(warnings) > 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	// c1 (80) should get H1 (100)
	// c2 (40) should get H2 (50)
	if assignments[0].Halls != "H1" { // c1 is processed first due to higher enrollment
		t.Errorf("expected c1 to be in H1, got %s", assignments[0].Halls)
	}
	if assignments[1].Halls != "H2" {
		t.Errorf("expected c2 to be in H2, got %s", assignments[1].Halls)
	}
}

func TestAllocateHalls_MultiHall(t *testing.T) {
	assignments := []*Assignment{
		{CourseID: "c1", EnrolledCount: 120},
	}
	halls := []*Hall{
		{ID: "H1", Capacity: 100},
		{ID: "H2", Capacity: 50},
		{ID: "H3", Capacity: 30},
	}
	usedHalls := make(map[SlotID]map[HallID]bool)
	slotID := SlotID("slot1")

	_, warnings, err := AllocateHalls(assignments, halls, usedHalls, slotID)
	if err != nil {
		t.Fatalf("AllocateHalls failed: %v", err)
	}
	if len(warnings) > 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	// c1 (120) needs multiple halls. Should pick H1 and H2 (100+50)
	if assignments[0].Halls != "H1;H2" {
		t.Errorf("expected c1 to be in H1;H2, got %s", assignments[0].Halls)
	}
}

func TestAllocateHalls_CapacityWarning(t *testing.T) {
	assignments := []*Assignment{
		{CourseID: "c1", EnrolledCount: 200},
	}
	halls := []*Hall{
		{ID: "H1", Capacity: 100},
		{ID: "H2", Capacity: 50},
	}
	usedHalls := make(map[SlotID]map[HallID]bool)
	slotID := SlotID("slot1")

	_, warnings, err := AllocateHalls(assignments, halls, usedHalls, slotID)
	if err != nil {
		t.Fatalf("AllocateHalls failed: %v", err)
	}
	if len(warnings) != 1 {
		t.Fatalf("expected 1 capacity warning, got %d", len(warnings))
	}
}
