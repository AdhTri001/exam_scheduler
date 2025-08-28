//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"syscall/js"
	"time"

	"exam-scheduler/pkg/scheduler"
)

// --- Structs for JSON Payloads ---

type RunParams struct {
	ExamStartDate   string                   `json:"examStartDate"`
	ExamEndDate     string                   `json:"examEndDate"`
	SlotsPerDay     int                      `json:"slotsPerDay"`
	SlotTimes       []string                 `json:"slotTimes"`
	SlotDuration    int                      `json:"slotDuration"`
	Holidays        []string                 `json:"holidays"`
	Tries           int                      `json:"tries"`
	Seed            int64                    `json:"seed"`
	MinGap          int                      `json:"minGap"`
	AllowedSlotsCSV string                   `json:"allowedSlotsCSV"`
	Timezone        string                   `json:"timezone"` // IANA TZ string
	ColumnMapping   *scheduler.ColumnMapping `json:"columnMapping,omitempty"`
}

type SuccessResponse struct {
	Success     bool                        `json:"success"`
	ScheduleCSV string                      `json:"scheduleCSV,omitempty"`
	Report      *scheduler.ValidationReport `json:"report,omitempty"`
	Stats       *Stats                      `json:"stats,omitempty"`
}

type ErrorResponse struct {
	Success bool                        `json:"success"`
	Error   string                      `json:"error"`
	Report  *scheduler.ValidationReport `json:"report,omitempty"`
	Stats   *Stats                      `json:"stats,omitempty"`
}

type Stats struct {
	Seed        int64   `json:"seed"`
	TotalTime   float64 `json:"totalTime"` // in ms
	Attempts    int     `json:"attempts"`
	BestPenalty float64 `json:"bestPenalty"`
	SlotsUsed   int     `json:"slotsUsed"`
}

type VersionInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Go      string `json:"go"`
}

func main() {
	c := make(chan struct{}, 0)
	js.Global().Set("version", js.FuncOf(version))
	js.Global().Set("runSchedule", js.FuncOf(runSchedule))
	js.Global().Set("verify", js.FuncOf(verify))
	<-c
}

func version(this js.Value, args []js.Value) interface{} {
	info := VersionInfo{
		Name:    "exam-scheduler-go",
		Version: "1.0.0",
		Go:      runtime.Version(),
	}
	result, err := json.Marshal(info)
	if err != nil {
		return `{"error":"failed to marshal version info"}`
	}
	return string(result)
}

func runSchedule(this js.Value, args []js.Value) interface{} {
	startTime := time.Now()

	regCSV := args[0].String()
	hallsCSV := args[1].String()
	paramsJSON := args[2].String()

	var params RunParams
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return marshalError(fmt.Sprintf("failed to parse params JSON: %v", err), nil, 0, 0)
	}

	if params.Tries == 0 {
		params.Tries = 100
	}
	if params.Timezone == "" {
		params.Timezone = "UTC"
	}

	// Use provided seed or generate a new one
	seed := params.Seed
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	stats := &Stats{Seed: seed, Attempts: params.Tries}

	// 1. Parse Inputs
	courses, registrations, err := scheduler.ParseRegistrations(regCSV, params.ColumnMapping)
	if err != nil {
		return marshalError(fmt.Sprintf("failed to parse registrations CSV: %v", err), nil, seed, time.Since(startTime).Seconds()*1000)
	}
	halls, err := scheduler.ParseHalls(hallsCSV, params.ColumnMapping)
	if err != nil {
		return marshalError(fmt.Sprintf("failed to parse halls CSV: %v", err), nil, seed, time.Since(startTime).Seconds()*1000)
	}
	allowedSlots, err := scheduler.ParseAllowedSlots(params.AllowedSlotsCSV)
	if err != nil {
		return marshalError(fmt.Sprintf("failed to parse allowed slots CSV: %v", err), nil, seed, time.Since(startTime).Seconds()*1000)
	}

	// 2. Generate Slots
	slots, err := scheduler.GenerateSlots(params.ExamStartDate, params.ExamEndDate, params.SlotsPerDay, params.SlotTimes, params.SlotDuration, params.Holidays, params.Timezone)
	if err != nil {
		return marshalError(fmt.Sprintf("failed to generate slots: %v", err), nil, seed, time.Since(startTime).Seconds()*1000)
	}

	// 3. Build Conflict Graph
	graph := scheduler.NewConflictGraph(courses)

	// 4. Run Scheduler
	penaltyConfig := scheduler.PenaltyConfig{StudentProximityWeight: 1.0, MinGapViolationWeight: 10.0}
	result, err := scheduler.RunSchedulingAttempts(params.Tries, seed, courses, halls, slots, allowedSlots, graph, params.MinGap, penaltyConfig)
	if err != nil {
		return marshalError(fmt.Sprintf("scheduling failed: %v", err), nil, seed, time.Since(startTime).Seconds()*1000)
	}

	// 5. Serialize final schedule
	scheduleCSV, err := scheduler.SerializeAssignments(result.Assignments)
	if err != nil {
		return marshalError(fmt.Sprintf("failed to serialize schedule: %v", err), nil, seed, time.Since(startTime).Seconds()*1000)
	}

	// 6. Final verification
	finalReport, _ := scheduler.VerifySchedule(registrations, scheduleCSV, halls)
	finalReport.CapacityWarnings = result.Report.CapacityWarnings // Carry over warnings from allocation

	// 7. Populate stats and response
	stats.TotalTime = time.Since(startTime).Seconds() * 1000
	stats.BestPenalty = result.Penalty

	usedSlots := make(map[scheduler.SlotID]bool)
	for _, a := range result.Assignments {
		usedSlots[a.SlotID] = true
	}
	stats.SlotsUsed = len(usedSlots)

	response := SuccessResponse{
		Success:     true,
		ScheduleCSV: scheduleCSV,
		Report:      finalReport,
		Stats:       stats,
	}

	jsonResponse, _ := json.Marshal(response)
	return string(jsonResponse)
}

func verify(this js.Value, args []js.Value) interface{} {
	regCSV := args[0].String()
	scheduleCSV := args[1].String()
	// For verification, we need halls to check capacity, but it's not in the API contract.
	// Let's assume we need to parse it from somewhere or the contract is flawed.
	// For now, we'll proceed without hall capacity checks if hallsCSV is not provided.
	// Let's add a dummy halls slice. A better approach would be to adjust the contract.
	halls := []*scheduler.Hall{}

	_, registrations, err := scheduler.ParseRegistrations(regCSV, nil)
	if err != nil {
		return marshalError(fmt.Sprintf("failed to parse registrations CSV for verification: %v", err), nil, 0, 0)
	}

	report, err := scheduler.VerifySchedule(registrations, scheduleCSV, halls)
	if err != nil {
		// This error is for catastrophic parsing issues, not validation failures.
		return marshalError(fmt.Sprintf("verification failed with an error: %v", err), report, 0, 0)
	}

	response := SuccessResponse{
		Success: true, // The function succeeded, even if the schedule is invalid
		Report:  report,
	}
	jsonResponse, _ := json.Marshal(response)
	return string(jsonResponse)
}

func marshalError(errMsg string, report *scheduler.ValidationReport, seed int64, totalTime float64) string {
	errResp := ErrorResponse{
		Success: false,
		Error:   errMsg,
		Report:  report,
		Stats:   &Stats{Seed: seed, TotalTime: totalTime},
	}
	jsonResponse, _ := json.Marshal(errResp)
	return string(jsonResponse)
}
