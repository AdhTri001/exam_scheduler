package scheduler

import (
	"testing"
)

func TestParseRegistrations_CustomColumnMapping(t *testing.T) {
	// CSV with custom column names
	csvData := `student_number,subject_code,extra_field
s001,CS101,ignore
s001,MATH201,ignore
s002,CS101,ignore`

	// Custom column mapping
	mapping := &ColumnMapping{
		StudentIDColumn: "student_number",
		CourseIDColumn:  "subject_code",
	}

	courses, registrations, err := ParseRegistrations(csvData, mapping)
	if err != nil {
		t.Fatalf("ParseRegistrations failed: %v", err)
	}

	if len(courses) != 2 {
		t.Errorf("expected 2 courses, got %d", len(courses))
	}

	if len(registrations) != 3 {
		t.Errorf("expected 3 registrations, got %d", len(registrations))
	}

	// Check that courses were parsed correctly
	expectedCourses := map[CourseID]bool{
		"CS101":   false,
		"MATH201": false,
	}

	for _, course := range courses {
		if _, exists := expectedCourses[course.ID]; exists {
			expectedCourses[course.ID] = true
		}
	}

	for courseID, found := range expectedCourses {
		if !found {
			t.Errorf("course %s not found", courseID)
		}
	}
}

func TestParseHalls_CustomColumnMapping(t *testing.T) {
	// CSV with custom column names
	csvData := `room_id,max_students,department,extra
H001,50,CS,ignore
H002,30,MATH,ignore`

	// Custom column mapping
	mapping := &ColumnMapping{
		HallIDColumn:   "room_id",
		CapacityColumn: "max_students",
		GroupColumn:    "department",
	}

	halls, err := ParseHalls(csvData, mapping)
	if err != nil {
		t.Fatalf("ParseHalls failed: %v", err)
	}

	if len(halls) != 2 {
		t.Errorf("expected 2 halls, got %d", len(halls))
	}

	// Check first hall
	if halls[0].ID != "H001" {
		t.Errorf("expected hall ID H001, got %s", halls[0].ID)
	}
	if halls[0].Capacity != 50 {
		t.Errorf("expected capacity 50, got %d", halls[0].Capacity)
	}
	if halls[0].Group != "CS" {
		t.Errorf("expected group CS, got %s", halls[0].Group)
	}

	// Check second hall
	if halls[1].ID != "H002" {
		t.Errorf("expected hall ID H002, got %s", halls[1].ID)
	}
	if halls[1].Capacity != 30 {
		t.Errorf("expected capacity 30, got %d", halls[1].Capacity)
	}
	if halls[1].Group != "MATH" {
		t.Errorf("expected group MATH, got %s", halls[1].Group)
	}
}

func TestParseRegistrations_MissingColumns(t *testing.T) {
	// CSV missing required column when using custom mapping
	csvData := `student_number,extra_field
s001,ignore`

	mapping := &ColumnMapping{
		StudentIDColumn: "student_number",
		CourseIDColumn:  "nonexistent_column",
	}

	_, _, err := ParseRegistrations(csvData, mapping)
	if err == nil {
		t.Error("expected error for missing required column, got nil")
	}
}

func TestParseHalls_MissingColumns(t *testing.T) {
	// CSV missing required column when using custom mapping
	csvData := `room_id,extra_field
H001,ignore`

	mapping := &ColumnMapping{
		HallIDColumn:   "room_id",
		CapacityColumn: "nonexistent_column",
	}

	_, err := ParseHalls(csvData, mapping)
	if err == nil {
		t.Error("expected error for missing required column, got nil")
	}
}

func TestParseRegistrations_DefaultColumns(t *testing.T) {
	// Test that default columns still work when mapping is nil
	csvData := `student_id,course_id
s001,CS101
s002,MATH201`

	courses, registrations, err := ParseRegistrations(csvData, nil)
	if err != nil {
		t.Fatalf("ParseRegistrations failed: %v", err)
	}

	if len(courses) != 2 {
		t.Errorf("expected 2 courses, got %d", len(courses))
	}

	if len(registrations) != 2 {
		t.Errorf("expected 2 registrations, got %d", len(registrations))
	}
}

func TestParseHalls_DefaultColumns(t *testing.T) {
	// Test that default columns still work when mapping is nil
	csvData := `hall,capacity,group
H001,50,CS
H002,30,MATH`

	halls, err := ParseHalls(csvData, nil)
	if err != nil {
		t.Fatalf("ParseHalls failed: %v", err)
	}

	if len(halls) != 2 {
		t.Errorf("expected 2 halls, got %d", len(halls))
	}
}
