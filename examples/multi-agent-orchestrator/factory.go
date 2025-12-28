package main

import (
	"fmt"
	"time"

	"github.com/taipm/go-signal-agent/examples/multi-agent-orchestrator/config"
	"github.com/taipm/go-signal-agent/examples/multi-agent-orchestrator/memory"
	"github.com/taipm/go-signal-agent/ollama"
	"github.com/taipm/go-signal-agent/signal"
)

// Factory creates and manages all agents
type Factory struct {
	cfg          *config.Config
	memMgr       *memory.Manager
	ollamaClient *ollama.Client
	resultChan   chan *signal.Signal
}

// NewFactory creates a new factory
func NewFactory(cfg *config.Config, memMgr *memory.Manager, resultChan chan *signal.Signal) (*Factory, error) {
	// Create Ollama client
	client := ollama.NewClient(ollama.ClientConfig{
		Endpoint: cfg.Ollama.Host,
		Timeout:  time.Duration(cfg.Ollama.Timeout) * time.Second,
	})

	return &Factory{
		cfg:          cfg,
		memMgr:       memMgr,
		ollamaClient: client,
		resultChan:   resultChan,
	}, nil
}

// CreateCoordinator creates the coordinator agent
func (f *Factory) CreateCoordinator() *CoordinatorAgent {
	return NewCoordinatorAgent(&f.cfg.Coordinator, f.ollamaClient)
}

// CreateWorker creates a worker agent by ID
func (f *Factory) CreateWorker(workerID string) (*WorkerAgent, error) {
	workerCfg, ok := f.cfg.Workers[workerID]
	if !ok {
		return nil, fmt.Errorf("worker config not found: %s", workerID)
	}

	// Resolve memory config
	resolvedMem := f.cfg.ResolveWorkerMemory(workerID)

	var memStore *memory.Store
	if resolvedMem.Enabled && len(resolvedMem.Types) > 0 {
		// Use first memory type
		memStore = f.memMgr.GetStoreByType(workerID, resolvedMem.Types[0].Name)
	}

	return NewWorkerAgent(&workerCfg, memStore, f.ollamaClient), nil
}

// CreateAllWorkers creates all configured worker agents
func (f *Factory) CreateAllWorkers() ([]*WorkerAgent, error) {
	workers := make([]*WorkerAgent, 0)
	for workerID := range f.cfg.Workers {
		worker, err := f.CreateWorker(workerID)
		if err != nil {
			return nil, fmt.Errorf("create worker %s: %w", workerID, err)
		}
		workers = append(workers, worker)
	}
	return workers, nil
}

// CreateOutputAgent creates the output agent
func (f *Factory) CreateOutputAgent() *OutputAgent {
	return NewOutputAgent(&f.cfg.Output, f.ollamaClient, f.resultChan)
}

// CreateAllAgents creates all agents and returns them
func (f *Factory) CreateAllAgents() (*CoordinatorAgent, []*WorkerAgent, *OutputAgent, error) {
	coordinator := f.CreateCoordinator()

	workers, err := f.CreateAllWorkers()
	if err != nil {
		return nil, nil, nil, err
	}

	output := f.CreateOutputAgent()

	return coordinator, workers, output, nil
}

// RegisterAllAgents registers all agents with the router
func (f *Factory) RegisterAllAgents(router *signal.Router) error {
	coordinator, workers, output, err := f.CreateAllAgents()
	if err != nil {
		return err
	}

	router.Register(coordinator)
	for _, worker := range workers {
		router.Register(worker)
	}
	router.Register(output)

	return nil
}

// GetOllamaClient returns the shared Ollama client
func (f *Factory) GetOllamaClient() *ollama.Client {
	return f.ollamaClient
}

// GetWorkerIDs returns all configured worker IDs
func (f *Factory) GetWorkerIDs() []string {
	return f.cfg.GetWorkerIDs()
}
