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
# Clone the repository
git clone https://github.com/taipm/go-signal-agent.git

# Install to your Go project
cd /your/go/project
/path/to/go-signal-agent/.claude/agents/go-dev-portable/install.sh

# Or install globally (available in all projects)
/path/to/go-signal-agent/.claude/agents/go-dev-portable/install.sh --global
```

---

## Detailed Migration Scenarios

### Scenario 1: New Project on Same Machine

```bash
# You have go-signal-agent cloned somewhere
cd ~/my-new-go-project

# Install from existing clone
~/GitHub/go-signal-agent/.claude/agents/go-dev-portable/install.sh

# Verify
cat .claude/agents/go-dev/VERSION  # Should show 1.0.0
```

### Scenario 2: New Machine (Fresh Install)

```bash
# Step 1: Clone the repo
git clone https://github.com/taipm/go-signal-agent.git ~/tools/go-signal-agent

# Step 2: Install globally (recommended for new machine)
~/tools/go-signal-agent/.claude/agents/go-dev-portable/install.sh --global

# Step 3: Verify installation
cat ~/.claude/agents/go-dev/VERSION
ls ~/.claude/agents/go-dev/knowledge/

# Now /go-dev works in ANY project on this machine
```

### Scenario 3: Team Sharing (Multiple Developers)

```bash
# Option A: Each developer clones and installs globally
git clone https://github.com/taipm/go-signal-agent.git
./go-signal-agent/.claude/agents/go-dev-portable/install.sh --global

# Option B: Include in project (version controlled)
cd your-team-project
cp -r /path/to/go-dev-portable .claude/agents/go-dev-portable/
git add .claude/agents/go-dev-portable/
git commit -m "Add go-dev-agent for team"

# Each team member runs:
./.claude/agents/go-dev-portable/install.sh
```

### Scenario 4: CI/CD Pipeline

```yaml
# .github/workflows/setup.yml
- name: Setup go-dev-agent
  run: |
    git clone --depth 1 https://github.com/taipm/go-signal-agent.git /tmp/agent
    /tmp/agent/.claude/agents/go-dev-portable/install.sh
```

### Scenario 5: Offline Installation (Air-gapped Machine)

```bash
# On machine with internet
cd /path/to/go-signal-agent/.claude/agents/go-dev-portable
tar -czvf go-dev-portable.tar.gz .

# Transfer go-dev-portable.tar.gz to offline machine (USB, etc.)

# On offline machine
mkdir -p ~/go-dev-portable
tar -xzvf go-dev-portable.tar.gz -C ~/go-dev-portable/
cd /your/project
~/go-dev-portable/install.sh
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
# On old machine - Export all learnings
tar -czvf go-dev-learnings.tar.gz \
  .claude/agents/go-dev/knowledge/learning/archive/ \
  .claude/agents/go-dev/knowledge/09-learned-patterns.md \
  .claude/agents/go-dev/knowledge/10-learned-anti-patterns.md \
  .claude/agents/go-dev/knowledge/11-project-decisions.md

# Transfer to new machine, then:
cd /your/new/project
tar -xzvf go-dev-learnings.tar.gz
```

### Sync Learnings Across Projects

```bash
# If you have learnings in project A, copy to project B:
cp .claude/agents/go-dev/knowledge/10-learned-anti-patterns.md \
   /path/to/project-b/.claude/agents/go-dev/knowledge/

# Or sync all learning files:
rsync -av .claude/agents/go-dev/knowledge/learning/archive/ \
   /path/to/project-b/.claude/agents/go-dev/knowledge/learning/archive/
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
