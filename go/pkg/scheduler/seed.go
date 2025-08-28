package scheduler

import (
	"math/rand"
	"time"
)

// NewSeededRNG creates a new random number generator with a given seed.
// If the seed is 0, it uses the current time.
func NewSeededRNG(seed int64) *rand.Rand {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	source := rand.NewSource(seed)
	return rand.New(source)
}
