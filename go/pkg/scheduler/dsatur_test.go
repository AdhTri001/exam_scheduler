package scheduler

import (
	"testing"
)

func TestDSATUR(t *testing.T) {
	// Graph: c1-c2, c1-c3. c4 is isolated.
	courses := map[CourseID]*Course{
		"c1": {ID: "c1", Enrollments: []StudentID{"s1", "s2"}},
		"c2": {ID: "c2", Enrollments: []StudentID{"s1"}},
		"c3": {ID: "c3", Enrollments: []StudentID{"s2"}},
		"c4": {ID: "c4", Enrollments: []StudentID{"s3"}},
	}
	graph := NewConflictGraph(courses)
	slots, _ := GenerateSlots("2025-01-06", "2025-01-07", 2, []string{"09:00", "14:00"}, 180, nil, "UTC")

	coloring, err := DSATUR(graph, slots, make(map[CourseID]map[SlotID]bool), 123)
	if err != nil {
		t.Fatalf("DSATUR failed: %v", err)
	}

	if len(coloring) != 4 {
		t.Errorf("expected coloring for 4 courses, got %d", len(coloring))
	}

	// c1 must have a different slot from c2 and c3
	if coloring["c1"] == coloring["c2"] {
		t.Error("c1 and c2 have the same slot, but they conflict")
	}
	if coloring["c1"] == coloring["c3"] {
		t.Error("c1 and c3 have the same slot, but they conflict")
	}
	// c2 and c3 can have the same slot
	// c4 can have any slot
}

func TestDSATUR_Infeasible(t *testing.T) {
	// Clique of 3 courses, but only 2 slots available
	courses := map[CourseID]*Course{
		"c1": {ID: "c1", Enrollments: []StudentID{"s1", "s2"}},
		"c2": {ID: "c2", Enrollments: []StudentID{"s1", "s3"}},
		"c3": {ID: "c3", Enrollments: []StudentID{"s2", "s3"}},
	}
	graph := NewConflictGraph(courses)
	slots, _ := GenerateSlots("2025-01-06", "2025-01-06", 2, []string{"09:00", "14:00"}, 180, nil, "UTC")

	_, err := DSATUR(graph, slots, make(map[CourseID]map[SlotID]bool), 123)
	if err == nil {
		t.Fatal("DSATUR should have failed for an infeasible schedule, but it succeeded")
	}
}

func TestDSATUR_WithAllowedSlots(t *testing.T) {
	courses := map[CourseID]*Course{
		"c1": {ID: "c1", Enrollments: []StudentID{"s1"}},
		"c2": {ID: "c2", Enrollments: []StudentID{"s1"}},
	}
	graph := NewConflictGraph(courses)
	slots, _ := GenerateSlots("2025-01-06", "2025-01-06", 2, []string{"09:00", "14:00"}, 180, nil, "UTC")

	// Restrict c1 and c2 to the same slot, which is impossible
	allowedSlots := map[CourseID]map[SlotID]bool{
		"c1": {slots[0].ID: true},
		"c2": {slots[0].ID: true},
	}

	_, err := DSATUR(graph, slots, allowedSlots, 123)
	if err == nil {
		t.Fatal("DSATUR should have failed due to allowed slots constraint, but it succeeded")
	}

	// Now provide a valid restriction
	allowedSlots = map[CourseID]map[SlotID]bool{
		"c1": {slots[0].ID: true},
		"c2": {slots[1].ID: true},
	}
	coloring, err := DSATUR(graph, slots, allowedSlots, 123)
	if err != nil {
		t.Fatalf("DSATUR failed with valid restrictions: %v", err)
	}
	if coloring["c1"] != 0 {
		t.Errorf("c1 was not assigned to its only allowed slot")
	}
	if coloring["c2"] != 1 {
		t.Errorf("c2 was not assigned to its only allowed slot")
	}
}
