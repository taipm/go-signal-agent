package main

import (
	"github.com/taipm/go-signal-agent/signal"
)

// Orchestrator manages dynamic routing and result collection
type Orchestrator struct {
	outputAgent *OutputAgent
	workerIDs   []string
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(output *OutputAgent, workerIDs []string) *Orchestrator {
	return &Orchestrator{
		outputAgent: output,
		workerIDs:   workerIDs,
	}
}

// CreateRoutingRules returns routing rules for the engine
func (o *Orchestrator) CreateRoutingRules() []signal.RoutingRule {
	return []signal.RoutingRule{
		// Route user requests to coordinator
		func(sig *signal.Signal) []string {
			if sig.Type == SignalUserRequest {
				return []string{"coordinator"}
			}
			return nil
		},

		// Route task assignments to selected workers
		// Note: The coordinator sets explicit destinations, so this is backup
		func(sig *signal.Signal) []string {
			if sig.Type == SignalTaskAssignment {
				// Explicit destination is set by coordinator
				if sig.Destination != "" {
					return []string{sig.Destination}
				}
				// Fallback: route to all workers
				return o.workerIDs
			}
			return nil
		},

		// Route worker results to output agent
		func(sig *signal.Signal) []string {
			if sig.Type == SignalWorkerResult {
				return []string{"output"}
			}
			return nil
		},

		// Final response is terminal (no further routing)
		func(sig *signal.Signal) []string {
			if sig.Type == SignalFinalResponse {
				return nil
			}
			return nil
		},
	}
}

// RegisterTaskFromAssignment extracts worker count and registers with output agent
func (o *Orchestrator) RegisterTaskFromAssignment(assignment *TaskAssignment) {
	o.outputAgent.RegisterTask(assignment.TaskID, len(assignment.SelectedWorkers))
}
