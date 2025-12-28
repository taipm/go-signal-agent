package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/taipm/go-signal-agent/examples/multi-agent-orchestrator/config"
	"github.com/taipm/go-signal-agent/examples/multi-agent-orchestrator/memory"
	sig "github.com/taipm/go-signal-agent/signal"
)

const (
	defaultConfigPath = "examples/multi-agent-orchestrator/agents.yaml"
	resultChanSize    = 10
)

func main() {
	// Get config path from env or use default
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = defaultConfigPath
	}

	// Load configuration
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize memory manager
	memMgr := memory.NewManager(cfg.Memory)
	if err := memMgr.Initialize(); err != nil {
		log.Fatalf("Failed to initialize memory: %v", err)
	}

	// Create result channel for final responses
	resultChan := make(chan *sig.Signal, resultChanSize)

	// Create factory
	factory, err := NewFactory(cfg, memMgr, resultChan)
	if err != nil {
		log.Fatalf("Failed to create factory: %v", err)
	}

	// Create all agents
	coordinator, workers, outputAgent, err := factory.CreateAllAgents()
	if err != nil {
		log.Fatalf("Failed to create agents: %v", err)
	}

	// Create router and register agents
	router := sig.NewRouter()
	router.Register(coordinator)
	for _, worker := range workers {
		router.Register(worker)
	}
	router.Register(outputAgent)

	// Create orchestrator and add routing rules
	workerIDs := make([]string, len(workers))
	for i, w := range workers {
		workerIDs[i] = w.ID()
	}
	orchestrator := NewOrchestrator(outputAgent, workerIDs)
	for _, rule := range orchestrator.CreateRoutingRules() {
		router.AddRule(rule)
	}

	// Create and start engine
	engine := sig.NewEngine(sig.EngineConfig{
		BufferSize:     50,
		WorkerCount:    5,
		ProcessTimeout: 180 * time.Second,
	}, router)

	// Add hook to register tasks with output agent
	engine.OnSignalProcessed(func(input *sig.Signal, result sig.AgentResult) {
		if input.Type == SignalUserRequest && len(result.Signals) > 0 {
			// Extract task assignment from first output signal
			for _, outSig := range result.Signals {
				if outSig.Type == SignalTaskAssignment {
					if assignment, ok := outSig.Payload.(*TaskAssignment); ok {
						orchestrator.RegisterTaskFromAssignment(assignment)
					}
				}
			}
		}
	})

	engine.Start()

	// Setup graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\n\nShutting down...")
		cancel()
	}()

	// Print welcome message
	printWelcome(workers)

	// Start CLI loop
	runCLI(ctx, engine, resultChan, workers, memMgr)

	// Cleanup
	engine.Stop()
	if err := memMgr.Shutdown(); err != nil {
		log.Printf("Error saving memory: %v", err)
	}
	fmt.Println("Goodbye! / Tam biet!")
}

func printWelcome(workers []*WorkerAgent) {
	fmt.Println("╔═══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║           Multi-Agent Orchestrator with Memory                    ║")
	fmt.Println("║                                                                   ║")
	fmt.Println("║  Agents:                                                          ║")
	fmt.Println("║    - Coordinator: Intelligent request routing (LLM-based)         ║")
	for _, w := range workers {
		name := strings.ToUpper(w.ID()[:1]) + w.ID()[1:]
		desc := w.config.Description
		// Truncate if too long
		maxDescLen := 40
		if len(desc) > maxDescLen {
			desc = desc[:maxDescLen-3] + "..."
		}
		line := fmt.Sprintf("    - %s Worker: %s", name, desc)
		padding := 67 - len(line) - 1 // 67 is inner width, -1 for closing ║
		if padding < 1 {
			padding = 1
		}
		fmt.Printf("║%s%s║\n", line, strings.Repeat(" ", padding))
	}
	fmt.Println("║    - Output Agent: Response synthesis                             ║")
	fmt.Println("║                                                                   ║")
	fmt.Println("║  Commands: /status, /memory, /workers, /clear, /quit              ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func runCLI(ctx context.Context, engine *sig.Engine, resultChan chan *sig.Signal, workers []*WorkerAgent, memMgr *memory.Manager) {
	scanner := bufio.NewScanner(os.Stdin)
	sessionID := fmt.Sprintf("session-%d", time.Now().Unix())

	// Single persistent input goroutine to avoid blocking on stdin
	// This goroutine reads from stdin and sends to inputChan.
	// It exits when stdin is closed, scanner error, or context cancelled.
	inputChan := make(chan string)
	inputDone := make(chan struct{})
	go func() {
		defer close(inputDone)
		defer close(inputChan)
		for scanner.Scan() {
			select {
			case inputChan <- scanner.Text():
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		fmt.Print("[orchestrator] You: ")

		select {
		case <-ctx.Done():
			goto cleanup
		case input, ok := <-inputChan:
			if !ok {
				goto cleanup
			}

			input = strings.TrimSpace(input)
			if input == "" {
				continue
			}

			// Handle commands
			if strings.HasPrefix(input, "/") {
				if handleCommand(input, workers, memMgr) {
					continue
				}
				if input == "/quit" || input == "/exit" || input == "/q" {
					goto cleanup
				}
			}

			// Detect language
			language := detectLanguage(input)

			// Create user request signal
			userReq := &UserRequest{
				SessionID: sessionID,
				Message:   input,
				Language:  language,
			}

			userSignal := sig.NewSignal(SignalUserRequest, userReq).
				WithMetadata("session_id", sessionID).
				WithMetadata("language", language)

			// Submit to engine
			fmt.Println("\nRouting your request...")
			if err := engine.Submit(userSignal); err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}

			// Wait for result with timeout
			select {
			case result := <-resultChan:
				displayResult(result)
			case <-time.After(120 * time.Second):
				fmt.Println("Request timed out. Please try again.")
			case <-ctx.Done():
				goto cleanup
			}
		}
	}

cleanup:
	// Wait for input goroutine to finish (with timeout to avoid blocking forever)
	select {
	case <-inputDone:
	case <-time.After(100 * time.Millisecond):
	}
}

func handleCommand(cmd string, workers []*WorkerAgent, memMgr *memory.Manager) bool {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return false
	}

	switch parts[0] {
	case "/status":
		fmt.Println("\n=== System Status ===")
		fmt.Printf("Workers: %d\n", len(workers))
		for _, w := range workers {
			stats := w.GetMemoryStats()
			fmt.Printf("  - %s: %d entries, %d bytes\n", w.ID(), stats["entries"], stats["size"])
		}
		fmt.Println()
		return true

	case "/memory":
		fmt.Println("\n=== Memory Status ===")
		for _, w := range workers {
			stats := w.GetMemoryStats()
			fmt.Printf("%s: %d entries (%d bytes)\n", w.ID(), stats["entries"], stats["size"])
		}
		fmt.Println()
		return true

	case "/workers":
		fmt.Println("\n=== Available Workers ===")
		for _, w := range workers {
			fmt.Printf("- %s: %s\n", w.ID(), w.config.Description)
			fmt.Printf("  Capabilities: %s\n", strings.Join(w.config.Capabilities, ", "))
		}
		fmt.Println()
		return true

	case "/clear":
		if len(parts) > 1 {
			// Clear specific worker memory
			for _, w := range workers {
				if w.ID() == parts[1] {
					w.ClearMemory()
					fmt.Printf("Cleared memory for %s\n", parts[1])
					return true
				}
			}
			fmt.Printf("Worker not found: %s\n", parts[1])
		} else {
			// Clear all memories
			for _, w := range workers {
				w.ClearMemory()
			}
			fmt.Println("Cleared all worker memories")
		}
		return true

	case "/help":
		fmt.Println("\n=== Available Commands ===")
		fmt.Println("/status  - Show system status")
		fmt.Println("/memory  - Show memory usage")
		fmt.Println("/workers - List available workers")
		fmt.Println("/clear   - Clear all memories")
		fmt.Println("/clear <worker> - Clear specific worker memory")
		fmt.Println("/quit    - Exit the application")
		fmt.Println()
		return true

	case "/quit", "/exit", "/q":
		return false // Signal to exit

	default:
		fmt.Printf("Unknown command: %s\n", parts[0])
		fmt.Println("Type /help for available commands")
		return true
	}
}

func displayResult(result *sig.Signal) {
	if result == nil {
		return
	}

	response, ok := result.Payload.(*FinalResponse)
	if !ok {
		fmt.Println("Invalid response type")
		return
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Contributors: %s\n", strings.Join(response.Contributors, ", "))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()
	fmt.Println(response.Content)
	fmt.Println()
}

func detectLanguage(text string) string {
	// Simple heuristic: check for Vietnamese characters
	vietnameseChars := "àáảãạăằắẳẵặâầấẩẫậèéẻẽẹêềếểễệìíỉĩịòóỏõọôồốổỗộơờớởỡợùúủũụưừứửữựỳýỷỹỵđ"
	for _, r := range text {
		if strings.ContainsRune(vietnameseChars, r) {
			return "vi"
		}
	}
	return "en"
}
