package scheduler

import (
	"testing"
)

func TestNewConflictGraph(t *testing.T) {
	courses := map[CourseID]*Course{
		"c1": {ID: "c1", Enrollments: []StudentID{"s1", "s2"}},
		"c2": {ID: "c2", Enrollments: []StudentID{"s1", "s3"}},
		"c3": {ID: "c3", Enrollments: []StudentID{"s2", "s4"}},
		"c4": {ID: "c4", Enrollments: []StudentID{"s5"}},
	}

	graph := NewConflictGraph(courses)

	if len(graph.Courses) != 4 {
		t.Fatalf("expected 4 courses in graph, got %d", len(graph.Courses))
	}

	// c1 conflicts with c2 (s1) and c3 (s2)
	// c2 conflicts with c1 (s1)
	// c3 conflicts with c1 (s2)
	// c4 conflicts with no one

	c1_idx := graph.CourseIndex["c1"]
	c2_idx := graph.CourseIndex["c2"]
	c3_idx := graph.CourseIndex["c3"]
	c4_idx := graph.CourseIndex["c4"]

	// Check edge weights
	if graph.AdjMatrix[c1_idx][c2_idx] != 1 {
		t.Errorf("expected edge weight 1 between c1 and c2, got %d", graph.AdjMatrix[c1_idx][c2_idx])
	}
	if graph.AdjMatrix[c1_idx][c3_idx] != 1 {
		t.Errorf("expected edge weight 1 between c1 and c3, got %d", graph.AdjMatrix[c1_idx][c3_idx])
	}
	if graph.AdjMatrix[c2_idx][c3_idx] != 0 {
		t.Errorf("expected edge weight 0 between c2 and c3, got %d", graph.AdjMatrix[c2_idx][c3_idx])
	}
	if graph.AdjMatrix[c1_idx][c4_idx] != 0 {
		t.Errorf("expected edge weight 0 between c1 and c4, got %d", graph.AdjMatrix[c1_idx][c4_idx])
	}

	// Check degrees
	if graph.Degrees[c1_idx] != 2 {
		t.Errorf("expected degree of c1 to be 2, got %d", graph.Degrees[c1_idx])
	}
	if graph.Degrees[c2_idx] != 1 {
		t.Errorf("expected degree of c2 to be 1, got %d", graph.Degrees[c2_idx])
	}
	if graph.Degrees[c3_idx] != 1 {
		t.Errorf("expected degree of c3 to be 1, got %d", graph.Degrees[c3_idx])
	}
	if graph.Degrees[c4_idx] != 0 {
		t.Errorf("expected degree of c4 to be 0, got %d", graph.Degrees[c4_idx])
	}
}
