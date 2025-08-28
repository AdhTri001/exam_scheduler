package scheduler

import (
	"testing"
)

const (
	validRegCSV = `student_id,course_id,extra_col
s1,c1,foo
s2,c1,bar
s1,c2,baz
# this is a comment
s3,c2,qux

"Doe, Jane",c3,test
s4,c3,
`
	validHallsCSV = `hall,capacity,group
H1,100,A
H2,50,
"H3, Big",200,B
`
	validAllowedSlotsCSV = `course_id,slot_id
c1,2025-01-07T09:00Z#1
c2,2025-01-07T14:00Z#2
`
)

func TestParseRegistrations(t *testing.T) {
	courses, registrations, err := ParseRegistrations(validRegCSV, nil)
	if err != nil {
		t.Fatalf("ParseRegistrations failed: %v", err)
	}

	if len(courses) != 3 {
		t.Errorf("expected 3 courses, got %d", len(courses))
	}
	if len(registrations) != 6 {
		t.Errorf("expected 6 registrations, got %d", len(registrations))
	}

	if len(courses["c1"].Enrollments) != 2 {
		t.Errorf("expected 2 enrollments for c1, got %d", len(courses["c1"].Enrollments))
	}
	if len(courses["c2"].Enrollments) != 2 {
		t.Errorf("expected 2 enrollments for c2, got %d", len(courses["c2"].Enrollments))
	}
	if len(courses["c3"].Enrollments) != 2 {
		t.Errorf("expected 2 enrollments for c3, got %d", len(courses["c3"].Enrollments))
	}
}

func TestParseHalls(t *testing.T) {
	halls, err := ParseHalls(validHallsCSV, nil)
	if err != nil {
		t.Fatalf("ParseHalls failed: %v", err)
	}

	if len(halls) != 3 {
		t.Errorf("expected 3 halls, got %d", len(halls))
	}

	if halls[0].ID != "H1" || halls[0].Capacity != 100 {
		t.Errorf("unexpected data for hall 1: %+v", halls[0])
	}
	if halls[2].ID != "H3, Big" || halls[2].Capacity != 200 {
		t.Errorf("unexpected data for hall 3: %+v", halls[2])
	}
}

func TestParseAllowedSlots(t *testing.T) {
	allowed, err := ParseAllowedSlots(validAllowedSlotsCSV)
	if err != nil {
		t.Fatalf("ParseAllowedSlots failed: %v", err)
	}

	if !allowed["c1"]["2025-01-07T09:00Z#1"] {
		t.Error("c1 should be allowed in slot 2025-01-07T09:00Z#1")
	}
	if len(allowed["c1"]) != 1 {
		t.Errorf("expected 1 allowed slot for c1, got %d", len(allowed["c1"]))
	}
}

func TestSerializeAssignments(t *testing.T) {
	assignments := []*Assignment{
		{CourseID: "c1", SlotID: "s1", SlotDateTime: "t1", Halls: "h1;h2", EnrolledCount: 10},
		{CourseID: "c2", SlotID: "s2", SlotDateTime: "t2", Halls: "h3", EnrolledCount: 20, Notes: "a note"},
	}

	csv, err := SerializeAssignments(assignments)
	if err != nil {
		t.Fatalf("SerializeAssignments failed: %v", err)
	}

	expected := `course_id,slot_id,slot_datetime,halls,enrolled_count,notes
c1,s1,t1,h1;h2,10,
c2,s2,t2,h3,20,a note
`
	if csv != expected {
		t.Errorf("unexpected CSV output.\nGot:\n%s\nExpected:\n%s", csv, expected)
	}
}
