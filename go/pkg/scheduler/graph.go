package scheduler

// ConflictGraph represents the course conflict graph.
// Nodes are courses, and edges represent shared students.
type ConflictGraph struct {
	Courses     []CourseID
	CourseIndex map[CourseID]int
	AdjMatrix   [][]int // Adjacency matrix with edge weights (shared students)
	Degrees     []int
}

// NewConflictGraph creates a new conflict graph from the given courses.
func NewConflictGraph(courses map[CourseID]*Course) *ConflictGraph {
	numCourses := len(courses)
	courseList := make([]CourseID, 0, numCourses)
	courseIndex := make(map[CourseID]int, numCourses)

	i := 0
	for courseID := range courses {
		courseList = append(courseList, courseID)
		courseIndex[courseID] = i
		i++
	}

	adjMatrix := make([][]int, numCourses)
	for i := range adjMatrix {
		adjMatrix[i] = make([]int, numCourses)
	}

	degrees := make([]int, numCourses)

	// Create a map for quick student lookup
	studentCourses := make(map[StudentID]map[CourseID]bool)
	for courseID, course := range courses {
		for _, studentID := range course.Enrollments {
			if _, ok := studentCourses[studentID]; !ok {
				studentCourses[studentID] = make(map[CourseID]bool)
			}
			studentCourses[studentID][courseID] = true
		}
	}

	// Populate adjacency matrix and degrees
	for _, studentCourseSet := range studentCourses {
		// Convert map keys to a slice for iteration
		conflictingCourses := make([]CourseID, 0, len(studentCourseSet))
		for courseID := range studentCourseSet {
			conflictingCourses = append(conflictingCourses, courseID)
		}

		for i := 0; i < len(conflictingCourses); i++ {
			for j := i + 1; j < len(conflictingCourses); j++ {
				c1 := courseIndex[conflictingCourses[i]]
				c2 := courseIndex[conflictingCourses[j]]

				if adjMatrix[c1][c2] == 0 {
					degrees[c1]++
					degrees[c2]++
				}
				adjMatrix[c1][c2]++
				adjMatrix[c2][c1]++
			}
		}
	}

	return &ConflictGraph{
		Courses:     courseList,
		CourseIndex: courseIndex,
		AdjMatrix:   adjMatrix,
		Degrees:     degrees,
	}
}
