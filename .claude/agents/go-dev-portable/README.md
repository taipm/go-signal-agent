# go-dev-agent Portable Package

A self-contained, portable Go development agent with Linus Torvalds-style quality standards, self-learning capabilities, and optimized knowledge loading.

## Features

- **Linus Torvalds Persona** - Brutally honest code reviews
- **Self-Learning System** - Captures and persists learnings across sessions
- **Relevance-Based Loading** - 50-80% context reduction
- **Parallel Quality Gates** - 50% faster validation
- **Pre-Code Checklist** - Mandatory quality checks before writing code

## Quick Install

```bash
# Clone or download this package
git clone <repo> /tmp/go-dev-portable

# Install to current project
cd /your/go/project
/tmp/go-dev-portable/install.sh

# Or install globally (available in all projects)
/tmp/go-dev-portable/install.sh --global
```

## Usage

After installation, use in Claude Code:

```
/go-dev implement a worker pool with graceful shutdown
/go-dev review this code for race conditions
/go-dev refactor to use context properly
```

## Package Contents

```
go-dev-portable/
├── install.sh              # Installer script
├── agent.md                # Agent definition (1,200+ lines)
├── README.md               # This file
└── knowledge/              # Knowledge base
    ├── 02-graceful-shutdown.md
    ├── 03-interactive-cli.md
    ├── 04-http-patterns.md
    ├── 05-llm-openai-go.md
    ├── 06-concurrency.md
    ├── 07-llm-ollama-local.md
    ├── 08-anti-patterns.md
    ├── 09-learned-patterns.md
    ├── 10-learned-anti-patterns.md
    ├── 11-project-decisions.md
    ├── knowledge-index.yaml
    └── learning/
        ├── config.md
        ├── raw/
        ├── pending/
        ├── archive/
        └── tools/
            ├── common.sh
            ├── select-knowledge.sh
            ├── parallel-validate.sh
            ├── knowledge-keywords.conf
            └── validate.conf
```

## Installation Options

### Project-Level (Recommended)
```bash
./install.sh
# Installs to: .claude/agents/go-dev/
# Creates skill: .claude/commands/go-dev.md
```

### Global
```bash
./install.sh --global
# Installs to: ~/.claude/agents/go-dev/
# Available in all projects
```

### Specific Project
```bash
./install.sh --project /path/to/project
```

### Uninstall
```bash
./install.sh --uninstall
./install.sh --uninstall --global
```

## Tools Included

### Parallel Validation
```bash
# Run quality gates with parallel execution
./knowledge/learning/tools/parallel-validate.sh --tier 2

# Tiers:
# 1 = Quick (build, vet, test)
# 2 = Standard (+ fmt, race)
# 3 = Comprehensive (+ staticcheck, gosec)
```

### Relevance-Based Knowledge Loading
```bash
# See which knowledge files will be loaded for a task
./knowledge/learning/tools/select-knowledge.sh "implement http server"

# Get file list only (for scripting)
./knowledge/learning/tools/select-knowledge.sh --files "worker pool"
```

## Self-Learning System

The agent learns from your project and persists knowledge:

### Capture a Learning
```
*learn-capture
```

### Review Pending Learnings
```
*learn-review
```

### Approve/Reject
```
*learn-approve:L001
*learn-reject:L002
```

Approved learnings are added to:
- `09-learned-patterns.md` - Good patterns
- `10-learned-anti-patterns.md` - Mistakes to avoid

## Migrating Learnings

When moving to a new machine, copy your learnings:

```bash
# On old machine
cp -r .claude/agents/go-dev/knowledge/learning/archive ~/go-dev-learnings-backup

# On new machine (after install)
cp -r ~/go-dev-learnings-backup/* .claude/agents/go-dev/knowledge/learning/archive/
```

## Requirements

- Bash 4.0+ (macOS: `brew install bash`)
- Go 1.21+
- Claude Code CLI

## Version

Current: 1.0.0

Check installed version:
```bash
cat .claude/agents/go-dev/VERSION
```

## License

MIT
