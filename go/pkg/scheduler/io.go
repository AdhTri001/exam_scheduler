package scheduler

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/gocarina/gocsv"
)

// ParseRegistrations parses the registrations CSV data with custom column mapping.
func ParseRegistrations(csvData string, columnMapping *ColumnMapping) (map[CourseID]*Course, []Registration, error) {
	courses := make(map[CourseID]*Course)
	var registrations []Registration

	reader := strings.NewReader(csvData)

	// Custom CSV reader configuration
	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1 // Allow variable number of columns
	csvReader.Comment = '#'

	// Read header to find column indices
	header, err := csvReader.Read()
	if err != nil {
		return nil, nil, err
	}

	// Use provided column names or defaults
	studentIDCol := "student_id"
	courseIDCol := "course_id"
	if columnMapping != nil {
		if columnMapping.StudentIDColumn != "" {
			studentIDCol = columnMapping.StudentIDColumn
		}
		if columnMapping.CourseIDColumn != "" {
			courseIDCol = columnMapping.CourseIDColumn
		}
	}

	studentIDIndex := -1
	courseIDIndex := -1
	for i, col := range header {
		if col == studentIDCol {
			studentIDIndex = i
		}
		if col == courseIDCol {
			courseIDIndex = i
		}
	}

	if studentIDIndex == -1 || courseIDIndex == -1 {
		return nil, nil, fmt.Errorf("missing required columns: %s or %s", studentIDCol, courseIDCol)
	}

	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Skip records with parsing errors, maybe log them
			continue
		}

		if len(record) <= studentIDIndex || len(record) <= courseIDIndex {
			continue // Skip malformed rows
		}

		studentID := StudentID(record[studentIDIndex])
		courseID := CourseID(record[courseIDIndex])

		if studentID == "" || courseID == "" {
			continue // Skip rows with empty required fields
		}

		reg := Registration{StudentID: studentID, CourseID: courseID}
		registrations = append(registrations, reg)

		if _, ok := courses[courseID]; !ok {
			courses[courseID] = &Course{ID: courseID}
		}
		courses[courseID].Enrollments = append(courses[courseID].Enrollments, studentID)
	}

	return courses, registrations, nil
}

// ParseHalls parses the halls CSV data with custom column mapping.
func ParseHalls(csvData string, columnMapping *ColumnMapping) ([]*Hall, error) {
	reader := strings.NewReader(csvData)

	// Custom CSV reader configuration
	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true
	csvReader.FieldsPerRecord = -1 // Allow variable number of columns
	csvReader.Comment = '#'

	// Read header to find column indices
	header, err := csvReader.Read()
	if err != nil {
		return nil, err
	}

	// Use provided column names or defaults
	hallIDCol := "hall"
	capacityCol := "capacity"
	groupCol := "group"
	if columnMapping != nil {
		if columnMapping.HallIDColumn != "" {
			hallIDCol = columnMapping.HallIDColumn
		}
		if columnMapping.CapacityColumn != "" {
			capacityCol = columnMapping.CapacityColumn
		}
		if columnMapping.GroupColumn != "" {
			groupCol = columnMapping.GroupColumn
		}
	}

	hallIDIndex := -1
	capacityIndex := -1
	groupIndex := -1
	for i, col := range header {
		if col == hallIDCol {
			hallIDIndex = i
		}
		if col == capacityCol {
			capacityIndex = i
		}
		if col == groupCol {
			groupIndex = i
		}
	}

	if hallIDIndex == -1 || capacityIndex == -1 {
		return nil, fmt.Errorf("missing required columns: %s or %s", hallIDCol, capacityCol)
	}

	var halls []*Hall
	for {
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			continue // Skip records with parsing errors
		}

		if len(record) <= hallIDIndex || len(record) <= capacityIndex {
			continue // Skip malformed rows
		}

		hallID := HallID(record[hallIDIndex])
		if hallID == "" {
			continue // Skip rows with empty hall ID
		}

		var capacity int
		if _, err := fmt.Sscanf(record[capacityIndex], "%d", &capacity); err != nil || capacity < 0 {
			continue // Skip rows with invalid capacity
		}

		var group string
		if groupIndex >= 0 && len(record) > groupIndex {
			group = record[groupIndex]
		}

		hall := &Hall{
			ID:       hallID,
			Capacity: capacity,
			Group:    group,
		}
		halls = append(halls, hall)
	}

	return halls, nil
} // ParseAllowedSlots parses the allowed slots CSV data.
func ParseAllowedSlots(csvData string) (map[CourseID]map[SlotID]bool, error) {
	allowed := make(map[CourseID]map[SlotID]bool)
	if csvData == "" {
		return allowed, nil
	}

	var allowedSlots []*AllowedSlot
	if err := gocsv.UnmarshalString(csvData, &allowedSlots); err != nil {
		return nil, err
	}

	for _, as := range allowedSlots {
		if _, ok := allowed[as.CourseID]; !ok {
			allowed[as.CourseID] = make(map[SlotID]bool)
		}
		allowed[as.CourseID][as.SlotID] = true
	}
	return allowed, nil
}

// SerializeAssignments serializes the schedule assignments to a CSV string.
func SerializeAssignments(assignments []*Assignment) (string, error) {
	var sb strings.Builder
	writer := csv.NewWriter(&sb)

	// Write header
	if err := writer.Write([]string{"course_id", "slot_id", "slot_datetime", "halls", "enrolled_count", "notes"}); err != nil {
		return "", err
	}

	for _, a := range assignments {
		record := []string{
			string(a.CourseID),
			string(a.SlotID),
			a.SlotDateTime,
			a.Halls,
			strconv.Itoa(a.EnrolledCount),
			a.Notes,
		}
		if err := writer.Write(record); err != nil {
			return "", err
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return "", err
	}

	return sb.String(), nil
}
