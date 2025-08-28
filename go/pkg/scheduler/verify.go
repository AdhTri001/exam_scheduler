package scheduler

import (
	"encoding/csv"
	"fmt"
	"strings"
)

// ValidationReport contains the results of a schedule verification.
type ValidationReport struct {
	Valid            bool       `json:"valid"`
	Conflicts        int        `json:"conflicts"`
	Unassigned       []CourseID `json:"unassigned"`
	CapacityWarnings []string   `json:"capacityWarnings"`
	Errors           []string   `json:"errors"`
	StudentClashes   []string   `json:"studentClashes"`
}

// VerifySchedule checks a generated schedule for correctness against the original registrations.
func VerifySchedule(registrations []Registration, scheduleCSV string, halls []*Hall) (*ValidationReport, error) {
	report := &ValidationReport{Valid: true}

	// Parse the schedule CSV
	assignments, err := parseScheduleCSV(scheduleCSV)
	if err != nil {
		report.Valid = false
		report.Errors = append(report.Errors, fmt.Sprintf("error parsing schedule CSV: %v", err))
		return report, err
	}

	// --- Data structures for verification ---
	// Map student to their scheduled slots
	studentSchedules := make(map[StudentID]map[SlotID]bool)
	// Map course to its assignment details
	assignmentMap := make(map[CourseID]*Assignment)
	for _, a := range assignments {
		assignmentMap[a.CourseID] = a
	}
	// Map hall capacities
	hallCapacityMap := make(map[HallID]int)
	for _, h := range halls {
		hallCapacityMap[h.ID] = h.Capacity
	}
	// Map slot to its total allocated capacity
	slotHallUsage := make(map[SlotID]map[HallID]int)

	// --- Check for student clashes ---
	for _, reg := range registrations {
		assignment, ok := assignmentMap[reg.CourseID]
		if !ok {
			// This course was in registrations but not in the schedule.
			// This is handled later as "unassigned".
			continue
		}

		if _, ok := studentSchedules[reg.StudentID]; !ok {
			studentSchedules[reg.StudentID] = make(map[SlotID]bool)
		}

		if studentSchedules[reg.StudentID][assignment.SlotID] {
			clash := fmt.Sprintf("student %s has a clash in slot %s", reg.StudentID, assignment.SlotID)
			report.StudentClashes = append(report.StudentClashes, clash)
			report.Conflicts++
			report.Valid = false
		}
		studentSchedules[reg.StudentID][assignment.SlotID] = true
	}

	// --- Check for hall overbooking and capacity ---
	for _, assignment := range assignments {
		// Check hall capacity
		assignedHalls := strings.Split(assignment.Halls, ";")
		totalCapacity := 0
		for _, hallIDStr := range assignedHalls {
			if hallIDStr == "" {
				continue
			}
			hallID := HallID(hallIDStr)
			capacity, ok := hallCapacityMap[hallID]
			if !ok {
				warning := fmt.Sprintf("course %s assigned to unknown hall %s", assignment.CourseID, hallID)
				report.CapacityWarnings = append(report.CapacityWarnings, warning)
				continue
			}
			totalCapacity += capacity

			// Check for double-booking the same hall in the same slot
			if _, ok := slotHallUsage[assignment.SlotID]; !ok {
				slotHallUsage[assignment.SlotID] = make(map[HallID]int)
			}
			slotHallUsage[assignment.SlotID][hallID]++
			if slotHallUsage[assignment.SlotID][hallID] > 1 {
				errStr := fmt.Sprintf("hall %s is double-booked in slot %s", hallID, assignment.SlotID)
				report.Errors = append(report.Errors, errStr)
				report.Valid = false
			}
		}

		if totalCapacity < assignment.EnrolledCount {
			warning := fmt.Sprintf("course %s has insufficient capacity. Enrolled: %d, Allocated: %d in halls [%s]",
				assignment.CourseID, assignment.EnrolledCount, totalCapacity, assignment.Halls)
			report.CapacityWarnings = append(report.CapacityWarnings, warning)
		}
	}

	// --- Check for unassigned courses ---
	allCoursesInRegs := make(map[CourseID]bool)
	for _, reg := range registrations {
		allCoursesInRegs[reg.CourseID] = true
	}
	for courseID := range allCoursesInRegs {
		if _, scheduled := assignmentMap[courseID]; !scheduled {
			report.Unassigned = append(report.Unassigned, courseID)
		}
	}
	if len(report.Unassigned) > 0 {
		report.Valid = false
	}

	return report, nil
}

// parseScheduleCSV is a helper to parse the schedule CSV for verification.
func parseScheduleCSV(csvData string) ([]*Assignment, error) {
	var assignments []*Assignment
	// gocsv has issues with custom parsing needs here, so we use the standard library
	reader := csv.NewReader(strings.NewReader(csvData))
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}
	// Basic validation of header
	expectedHeaders := []string{"course_id", "slot_id", "slot_datetime", "halls", "enrolled_count"}
	for _, h := range expectedHeaders {
		found := false
		for _, fileH := range header {
			if fileH == h {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("missing required header column: %s", h)
		}
	}

	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	for _, record := range records {
		if len(record) < 6 {
			continue // Skip malformed records
		}
		// Manual mapping to struct
		var enrolledCount int
		fmt.Sscanf(record[4], "%d", &enrolledCount)

		assignment := &Assignment{
			CourseID:      CourseID(record[0]),
			SlotID:        SlotID(record[1]),
			SlotDateTime:  record[2],
			Halls:         record[3],
			EnrolledCount: enrolledCount,
			Notes:         "",
		}
		if len(record) > 5 {
			assignment.Notes = record[5]
		}
		assignments = append(assignments, assignment)
	}

	return assignments, nil
}
