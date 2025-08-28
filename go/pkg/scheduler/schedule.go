package scheduler

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// ScheduleResult holds the outcome of a scheduling attempt.
type ScheduleResult struct {
	Assignments []*Assignment
	Penalty     float64
	Unassigned  []CourseID
	Report      *ValidationReport
}

// RunSchedulingAttempts runs the scheduling algorithm multiple times and returns the best result.
func RunSchedulingAttempts(
	tries int,
	seed int64,
	courses map[CourseID]*Course,
	halls []*Hall,
	slots []*Slot,
	allowedSlots map[CourseID]map[SlotID]bool,
	graph *ConflictGraph,
	minGapMinutes int,
	penaltyConfig PenaltyConfig,
) (*ScheduleResult, error) {

	var bestResult *ScheduleResult
	bestPenalty := -1.0

	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	rng := rand.New(rand.NewSource(seed))

	for i := 0; i < tries; i++ {
		attemptSeed := rng.Int63()

		coloring, err := DSATUR(graph, slots, allowedSlots, attemptSeed)
		if err != nil {
			// If one attempt is infeasible, it might be due to the random tie-breaking.
			// We can continue trying, but if all fail, we should report it.
			// For now, we'll just log it and continue.
			fmt.Printf("Attempt %d failed: %v\n", i+1, err)
			continue
		}

		// Group assignments by slot
		assignmentsBySlot := make(map[int][]*Assignment)
		allAssignments := make([]*Assignment, 0, len(coloring))

		for courseID, slotIdx := range coloring {
			slot := slots[slotIdx]
			assignment := &Assignment{
				CourseID:      courseID,
				SlotID:        slot.ID,
				SlotDateTime:  slot.Start.Format(time.RFC3339),
				EnrolledCount: len(courses[courseID].Enrollments),
			}
			assignmentsBySlot[slotIdx] = append(assignmentsBySlot[slotIdx], assignment)
			allAssignments = append(allAssignments, assignment)
		}

		// Allocate halls for each slot
		usedHalls := make(map[SlotID]map[HallID]bool)
		var allCapacityWarnings []string
		for slotIdx, assignmentsInSlot := range assignmentsBySlot {
			slotID := slots[slotIdx].ID
			_, warnings, err := AllocateHalls(assignmentsInSlot, halls, usedHalls, slotID)
			if err != nil {
				return nil, fmt.Errorf("hall allocation failed for slot %s: %w", slotID, err)
			}
			allCapacityWarnings = append(allCapacityWarnings, warnings...)
		}

		// Calculate penalty
		penalty := CalculatePenalty(coloring, courses, slots, graph, minGapMinutes, penaltyConfig)

		if bestResult == nil || penalty < bestPenalty {
			bestPenalty = penalty

			// Sort assignments for deterministic output
			sort.Slice(allAssignments, func(i, j int) bool {
				if allAssignments[i].SlotDateTime != allAssignments[j].SlotDateTime {
					return allAssignments[i].SlotDateTime < allAssignments[j].SlotDateTime
				}
				return allAssignments[i].CourseID < allAssignments[j].CourseID
			})

			bestResult = &ScheduleResult{
				Assignments: allAssignments,
				Penalty:     penalty,
				Report: &ValidationReport{
					CapacityWarnings: allCapacityWarnings,
					// Other report fields will be filled by the Verify function
				},
			}
		}
	}

	if bestResult == nil {
		return nil, fmt.Errorf("failed to find a valid schedule after %d attempts", tries)
	}

	return bestResult, nil
}
