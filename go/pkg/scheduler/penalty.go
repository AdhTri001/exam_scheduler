package scheduler

import "time"

// PenaltyConfig defines the weights for different penalty components.
type PenaltyConfig struct {
	StudentProximityWeight float64
	MinGapViolationWeight  float64
	// Add other penalty weights here
}

// CalculatePenalty calculates the total penalty for a given schedule.
func CalculatePenalty(
	schedule map[CourseID]int,
	courses map[CourseID]*Course,
	slots []*Slot,
	graph *ConflictGraph,
	minGapMinutes int,
	config PenaltyConfig,
) float64 {
	var totalPenalty float64

	studentSchedules := make(map[StudentID][]time.Time)

	for courseID, slotIdx := range schedule {
		course := courses[courseID]
		slot := slots[slotIdx]
		for _, studentID := range course.Enrollments {
			studentSchedules[studentID] = append(studentSchedules[studentID], slot.Start)
		}
	}

	minGapDuration := time.Duration(minGapMinutes) * time.Minute

	for _, exams := range studentSchedules {
		if len(exams) > 1 {
			for i := 0; i < len(exams); i++ {
				for j := i + 1; j < len(exams); j++ {
					gap := exams[i].Sub(exams[j])
					if gap < 0 {
						gap = -gap
					}

					// Proximity penalty (e.g., exams on the same day)
					if exams[i].Day() == exams[j].Day() {
						totalPenalty += config.StudentProximityWeight
					}

					// Minimum gap violation
					if minGapMinutes > 0 && gap < minGapDuration {
						totalPenalty += config.MinGapViolationWeight
					}
				}
			}
		}
	}

	return totalPenalty
}
