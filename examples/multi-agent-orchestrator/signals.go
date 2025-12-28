package main

import (
	"github.com/taipm/go-signal-agent/signal"
)

// Signal type constants for the orchestration flow
const (
	// SignalUserRequest is the initial user input signal
	SignalUserRequest signal.SignalType = "user_request"
	// SignalTaskAssignment is the coordinator's routing decision
	SignalTaskAssignment signal.SignalType = "task_assignment"
	// SignalWorkerResult is the output from each worker
	SignalWorkerResult signal.SignalType = "worker_result"
	// SignalFinalResponse is the consolidated output to user
	SignalFinalResponse signal.SignalType = "final_response"
)

// UserRequest represents the initial user input
type UserRequest struct {
	SessionID string `json:"session_id"`
	Message   string `json:"message"`
	Language  string `json:"language"` // "vi" or "en"
}

// TaskAssignment represents the coordinator's routing decision
type TaskAssignment struct {
	TaskID          string       `json:"task_id"`
	OriginalRequest *UserRequest `json:"original_request"`
	SelectedWorkers []string     `json:"selected_workers"`
	Context         string       `json:"context"`
}

// WorkerResult represents the output from each worker
type WorkerResult struct {
	TaskID     string  `json:"task_id"`
	WorkerID   string  `json:"worker_id"`
	Content    string  `json:"content"`
	Confidence float64 `json:"confidence"`
}

// FinalResponse represents the consolidated output to user
type FinalResponse struct {
	TaskID       string   `json:"task_id"`
	Content      string   `json:"content"`
	Contributors []string `json:"contributors"`
}

// RoutingDecision represents the LLM's routing decision
type RoutingDecision struct {
	Workers []string `json:"workers"`
	Reason  string   `json:"reason"`
}
