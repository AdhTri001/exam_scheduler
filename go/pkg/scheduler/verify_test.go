package scheduler

import (
	"testing"
)

func TestVerifySchedule_NoClashes(t *testing.T) {
	_, regs, _ := ParseRegistrations(`student_id,course_id
s1,c1
s1,c2
s2,c1
`, nil)
	scheduleCSV := `course_id,slot_id,slot_datetime,halls,enrolled_count,notes
c1,slot1,2025-01-06T09:00:00Z,H1,2,
c2,slot2,2025-01-06T14:00:00Z,H1,1,
`
	halls, _ := ParseHalls(`hall,capacity
H1,100
`, nil)

	report, err := VerifySchedule(regs, scheduleCSV, halls)
	if err != nil {
		t.Fatalf("VerifySchedule failed: %v", err)
	}

	if !report.Valid {
		t.Errorf("schedule should be valid, but was marked invalid. Report: %+v", report)
	}
	if report.Conflicts != 0 {
		t.Errorf("expected 0 conflicts, got %d", report.Conflicts)
	}
}

func TestVerifySchedule_WithClash(t *testing.T) {
	_, regs, _ := ParseRegistrations(`student_id,course_id
s1,c1
s1,c2
`, nil)
	scheduleCSV := `course_id,slot_id,slot_datetime,halls,enrolled_count,notes
c1,slot1,2025-01-06T09:00:00Z,H1,1,
c2,slot1,2025-01-06T09:00:00Z,H2,1,
`
	halls, _ := ParseHalls(`hall,capacity
H1,50
H2,50
`, nil)

	report, err := VerifySchedule(regs, scheduleCSV, halls)
	if err != nil {
		t.Fatalf("VerifySchedule failed: %v", err)
	}

	if report.Valid {
		t.Error("schedule should be invalid, but was marked valid")
	}
	if report.Conflicts != 1 {
		t.Errorf("expected 1 conflict, got %d", report.Conflicts)
	}
}

func TestVerifySchedule_Unassigned(t *testing.T) {
	_, regs, _ := ParseRegistrations(`student_id,course_id
s1,c1
s1,c2
`, nil)
	scheduleCSV := `course_id,slot_id,slot_datetime,halls,enrolled_count,notes
c1,slot1,2025-01-06T09:00:00Z,H1,1,
`
	halls, _ := ParseHalls(`hall,capacity
H1,50
`, nil)

	report, err := VerifySchedule(regs, scheduleCSV, halls)
	if err != nil {
		t.Fatalf("VerifySchedule failed: %v", err)
	}

	if report.Valid {
		t.Error("schedule should be invalid due to unassigned course")
	}
	if len(report.Unassigned) != 1 || report.Unassigned[0] != "c2" {
		t.Errorf("expected c2 to be unassigned, got %v", report.Unassigned)
	}
}

func TestVerifySchedule_CapacityWarning(t *testing.T) {
	_, regs, _ := ParseRegistrations(`student_id,course_id
s1,c1
s2,c1
`, nil)
	scheduleCSV := `course_id,slot_id,slot_datetime,halls,enrolled_count,notes
c1,slot1,2025-01-06T09:00:00Z,H1,2,
`
	halls, _ := ParseHalls(`hall,capacity
H1,1
`, nil)

	report, err := VerifySchedule(regs, scheduleCSV, halls)
	if err != nil {
		t.Fatalf("VerifySchedule failed: %v", err)
	}

	if len(report.CapacityWarnings) != 1 {
		t.Errorf("expected 1 capacity warning, got %d", len(report.CapacityWarnings))
	}
}
