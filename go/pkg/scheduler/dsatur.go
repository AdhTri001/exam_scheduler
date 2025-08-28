package scheduler

import (
	"fmt"
	"math/rand"
	"sort"
)

// DSATUR assigns colors (slots) to courses using the DSATUR algorithm.
// It returns a mapping of CourseID to SlotID, or an error if no solution is found.
func DSATUR(graph *ConflictGraph, slots []*Slot, allowedSlots map[CourseID]map[SlotID]bool, seed int64) (map[CourseID]int, error) {
	numCourses := len(graph.Courses)
	numSlots := len(slots)
	coloring := make(map[CourseID]int) // Maps CourseID to slot index

	// Initialize saturation degrees
	saturation := make([]int, numCourses)

	// Create a set of available slots for each course
	availableSlots := make([]map[int]bool, numCourses)
	for i := 0; i < numCourses; i++ {
		availableSlots[i] = make(map[int]bool)
		courseID := graph.Courses[i]

		// If restricted, only add allowed slots
		if allowed, ok := allowedSlots[courseID]; ok && len(allowed) > 0 {
			for slotIdx, slot := range slots {
				if allowed[slot.ID] {
					availableSlots[i][slotIdx] = true
				}
			}
		} else { // Otherwise, all slots are available
			for j := 0; j < numSlots; j++ {
				availableSlots[i][j] = true
			}
		}
	}

	// PRNG for tie-breaking
	rng := rand.New(rand.NewSource(seed))

	for len(coloring) < numCourses {
		// Find the uncolored vertex with the highest saturation degree
		maxSat := -1
		maxDegree := -1
		var nextCourseIdx = -1

		var candidates []int
		for i := 0; i < numCourses; i++ {
			if _, colored := coloring[graph.Courses[i]]; !colored {
				if saturation[i] > maxSat {
					maxSat = saturation[i]
					candidates = []int{i}
				} else if saturation[i] == maxSat {
					candidates = append(candidates, i)
				}
			}
		}

		// Tie-break with degree, then randomly
		if len(candidates) > 1 {
			var degreeCandidates []int
			for _, c := range candidates {
				if graph.Degrees[c] > maxDegree {
					maxDegree = graph.Degrees[c]
					degreeCandidates = []int{c}
				} else if graph.Degrees[c] == maxDegree {
					degreeCandidates = append(degreeCandidates, c)
				}
			}
			nextCourseIdx = degreeCandidates[rng.Intn(len(degreeCandidates))]
		} else if len(candidates) == 1 {
			nextCourseIdx = candidates[0]
		} else {
			// All courses colored
			break
		}

		// Assign the smallest possible color (slot index)
		courseID := graph.Courses[nextCourseIdx]

		// Get a sorted list of available slot indices for deterministic iteration
		possibleSlots := make([]int, 0, len(availableSlots[nextCourseIdx]))
		for slotIdx := range availableSlots[nextCourseIdx] {
			possibleSlots = append(possibleSlots, slotIdx)
		}
		sort.Ints(possibleSlots)

		assignedSlot := -1
		for _, slotIdx := range possibleSlots {
			// Check if this slot is available
			isAvailable := true
			for neighborIdx, weight := range graph.AdjMatrix[nextCourseIdx] {
				if weight > 0 {
					if neighborSlot, colored := coloring[graph.Courses[neighborIdx]]; colored && neighborSlot == slotIdx {
						isAvailable = false
						break
					}
				}
			}
			if isAvailable {
				assignedSlot = slotIdx
				break
			}
		}

		if assignedSlot == -1 {
			return nil, fmt.Errorf("infeasible schedule: cannot assign a slot to course %s", courseID)
		}

		coloring[courseID] = assignedSlot

		// Update saturation degrees of neighbors
		for neighborIdx, weight := range graph.AdjMatrix[nextCourseIdx] {
			if weight > 0 {
				if _, colored := coloring[graph.Courses[neighborIdx]]; !colored {
					// If neighbor had this slot as an option, its saturation increases
					if availableSlots[neighborIdx][assignedSlot] {
						// To be precise, we should check if this color was not already forbidden by another neighbor
						isNewForbiddenColor := true
						for otherNeighborIdx, otherWeight := range graph.AdjMatrix[neighborIdx] {
							if otherWeight > 0 {
								if otherSlot, colored := coloring[graph.Courses[otherNeighborIdx]]; colored && otherSlot == assignedSlot {
									isNewForbiddenColor = false
									break
								}
							}
						}
						if isNewForbiddenColor {
							saturation[neighborIdx]++
						}
					}
				}
			}
		}
	}

	return coloring, nil
}
