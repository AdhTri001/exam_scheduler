package scheduler

import (
	"fmt"
	"sort"
	"strings"
)

// AllocateHalls assigns halls to courses in a given slot.
func AllocateHalls(
	assignmentsInSlot []*Assignment,
	allHalls []*Hall,
	usedHalls map[SlotID]map[HallID]bool,
	slotID SlotID,
) (map[CourseID][]HallID, []string, error) {

	allocatedHalls := make(map[CourseID][]HallID)
	var capacityWarnings []string

	// Sort assignments by enrollment, descending, for deterministic packing
	sort.Slice(assignmentsInSlot, func(i, j int) bool {
		return assignmentsInSlot[i].EnrolledCount > assignmentsInSlot[j].EnrolledCount
	})

	// Available halls for this slot
	availableHalls := make([]*Hall, 0, len(allHalls))
	for _, hall := range allHalls {
		if used, ok := usedHalls[slotID]; !ok || !used[hall.ID] {
			availableHalls = append(availableHalls, hall)
		}
	}
	// Sort available halls by capacity, ascending, to find tightest fit
	sort.Slice(availableHalls, func(i, j int) bool {
		return availableHalls[i].Capacity < availableHalls[j].Capacity
	})

	if usedHalls[slotID] == nil {
		usedHalls[slotID] = make(map[HallID]bool)
	}

	for _, assignment := range assignmentsInSlot {
		neededCapacity := assignment.EnrolledCount
		var assignedHallsForCourse []HallID

		// Find the best single hall first
		bestFitIndex := -1
		for i, hall := range availableHalls {
			if hall.Capacity >= neededCapacity {
				if bestFitIndex == -1 || hall.Capacity < availableHalls[bestFitIndex].Capacity {
					bestFitIndex = i
				}
			}
		}

		if bestFitIndex != -1 {
			hall := availableHalls[bestFitIndex]
			assignedHallsForCourse = append(assignedHallsForCourse, hall.ID)
			usedHalls[slotID][hall.ID] = true
			// Remove from available list for this slot
			availableHalls = append(availableHalls[:bestFitIndex], availableHalls[bestFitIndex+1:]...)
		} else {
			// Try to combine multiple smaller halls (greedy approach)
			// Sort remaining halls by capacity descending to fill up faster
			sort.Slice(availableHalls, func(i, j int) bool {
				return availableHalls[i].Capacity > availableHalls[j].Capacity
			})

			currentCapacity := 0
			var combination []*Hall
			for _, hall := range availableHalls {
				if currentCapacity < neededCapacity {
					currentCapacity += hall.Capacity
					combination = append(combination, hall)
				}
			}

			if currentCapacity >= neededCapacity {
				for _, hall := range combination {
					assignedHallsForCourse = append(assignedHallsForCourse, hall.ID)
					usedHalls[slotID][hall.ID] = true
					// Remove from available list
					for i, ah := range availableHalls {
						if ah.ID == hall.ID {
							availableHalls = append(availableHalls[:i], availableHalls[i+1:]...)
							break
						}
					}
				}
			} else {
				// Not enough capacity even with all remaining halls
				msg := fmt.Sprintf("course %s (enrolled: %d) could not be fully allocated. Total available capacity: %d", assignment.CourseID, neededCapacity, currentCapacity)
				capacityWarnings = append(capacityWarnings, msg)
				// Assign what's available anyway
				for _, hall := range availableHalls {
					assignedHallsForCourse = append(assignedHallsForCourse, hall.ID)
					usedHalls[slotID][hall.ID] = true
				}
				availableHalls = nil // All used up
			}
		}
		allocatedHalls[assignment.CourseID] = assignedHallsForCourse
	}

	// Update the main assignment objects
	for _, a := range assignmentsInSlot {
		if halls, ok := allocatedHalls[a.CourseID]; ok {
			var hallIDs []string
			for _, h := range halls {
				hallIDs = append(hallIDs, string(h))
			}
			sort.Strings(hallIDs) // Deterministic output
			a.Halls = strings.Join(hallIDs, ";")
		}
	}

	return allocatedHalls, capacityWarnings, nil
}
