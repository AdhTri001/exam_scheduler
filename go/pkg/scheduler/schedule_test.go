package scheduler

import (
	"testing"
)

// End-to-end acceptance test
func TestRunSchedulingAttempts_Acceptance(t *testing.T) {
	regCSV := `student_id,course_id
s1,c1
s1,c2
s2,c1
s3,c3
s4,c3
s4,c4
`
	hallsCSV := `hall,capacity
H1,5
H2,2
`
	courses, _, _ := ParseRegistrations(regCSV, nil)
	halls, _ := ParseHalls(hallsCSV, nil)
	slots, _ := GenerateSlots("2025-01-20", "2025-01-21", 2, []string{"09:00", "14:00"}, 180, nil, "UTC")
	graph := NewConflictGraph(courses)

	result, err := RunSchedulingAttempts(10, 12345, courses, halls, slots, make(map[CourseID]map[SlotID]bool), graph, 60, PenaltyConfig{StudentProximityWeight: 1.0})
	if err != nil {
		t.Fatalf("RunSchedulingAttempts failed: %v", err)
	}

	if len(result.Assignments) != 4 {
		t.Errorf("expected 4 assignments, got %d", len(result.Assignments))
	}

	// Verify no conflicts in the result
	assignmentMap := make(map[CourseID]*Assignment)
	for _, a := range result.Assignments {
		assignmentMap[a.CourseID] = a
	}

	if assignmentMap["c1"].SlotID == assignmentMap["c2"].SlotID {
		t.Error("conflicting courses c1 and c2 are in the same slot")
	}
	if assignmentMap["c3"].SlotID == assignmentMap["c4"].SlotID {
		t.Error("conflicting courses c3 and c4 are in the same slot")
	}

	// Check hall allocation - just verify they got assigned to valid halls
	if assignmentMap["c1"].Halls == "" {
		t.Error("c1 was not assigned to any hall")
	}
	if assignmentMap["c2"].Halls == "" {
		t.Error("c2 was not assigned to any hall")
	}
	if assignmentMap["c3"].Halls == "" {
		t.Error("c3 was not assigned to any hall")
	}
	if assignmentMap["c4"].Halls == "" {
		t.Error("c4 was not assigned to any hall")
	}
}
