package scheduler

import "time"

// StudentID is a unique identifier for a student.
type StudentID string

// CourseID is a unique identifier for a course.
type CourseID string

// HallID is a unique identifier for a hall.
type HallID string

// SlotID is a unique identifier for a time slot.
type SlotID string

// Course represents a course with its enrollments.
type Course struct {
	ID          CourseID
	Enrollments []StudentID
}

// Hall represents an examination hall.
type Hall struct {
	ID       HallID `csv:"hall"`
	Capacity int    `csv:"capacity"`
	Group    string `csv:"group,omitempty"`
}

// Slot represents a time slot for an exam.
type Slot struct {
	ID         SlotID
	Start      time.Time
	End        time.Time
	DayIndex   int
	IndexInDay int
}

// Assignment represents a course scheduled in a specific slot and hall(s).
type Assignment struct {
	CourseID      CourseID `csv:"course_id"`
	SlotID        SlotID   `csv:"slot_id"`
	SlotDateTime  string   `csv:"slot_datetime"`
	Halls         string   `csv:"halls"` // Semicolon-separated list of HallIDs
	EnrolledCount int      `csv:"enrolled_count"`
	Notes         string   `csv:"notes,omitempty"`
}

// Registration represents a single student registration for a course.
type Registration struct {
	StudentID StudentID `csv:"student_id"`
	CourseID  CourseID  `csv:"course_id"`
}

// AllowedSlot represents a restriction for a course to be scheduled in a specific slot.
type AllowedSlot struct {
	CourseID CourseID `csv:"course_id"`
	SlotID   SlotID   `csv:"slot_id"`
}

// ColumnMapping defines which columns contain the required data
type ColumnMapping struct {
	StudentIDColumn string `json:"studentIdColumn"`
	CourseIDColumn  string `json:"courseIdColumn"`
	HallIDColumn    string `json:"hallIdColumn"`
	CapacityColumn  string `json:"capacityColumn"`
	GroupColumn     string `json:"groupColumn"`
}
