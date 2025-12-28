# Release Notes

## {Project Name} v{version}

**Release Date:** {date}
**Build:** {build_number}

---

## Overview

{Brief description of this release}

---

## Features

### New Features
- **{Feature 1}:** {Description}
- **{Feature 2}:** {Description}

### Improvements
- **{Improvement 1}:** {Description}
- **{Improvement 2}:** {Description}

### Bug Fixes
- **{Fix 1}:** {Description}
- **{Fix 2}:** {Description}

---

## API Changes

### New Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/{resource}/{id}` | Get resource by ID |
| POST | `/api/v1/{resource}` | Create new resource |

### Changed Endpoints

| Method | Endpoint | Change |
|--------|----------|--------|
| {method} | {endpoint} | {description of change} |

### Deprecated Endpoints

| Method | Endpoint | Replacement | Removal Date |
|--------|----------|-------------|--------------|
| {method} | {endpoint} | {new endpoint} | {date} |

---

## Breaking Changes

{List of breaking changes, if any}

- {Breaking change 1}
- {Breaking change 2}

**Migration Guide:**
{Steps to migrate from previous version}

---

## Dependencies

### Added
- `{package}` v{version} - {reason}

### Updated
- `{package}` v{old} → v{new}

### Removed
- `{package}` - {reason}

---

## Configuration

### New Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| {option} | {type} | {default} | {description} |

### Changed Defaults

| Option | Old Default | New Default |
|--------|-------------|-------------|
| {option} | {old} | {new} |

---

## Deployment

### Requirements
- Go 1.22+
- Docker (optional)

### Quick Start

```bash
# Build
make build

# Run
./bin/server

# Or with Docker
docker build -t app:{version} .
docker run -p 8080:8080 app:{version}
```

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| PORT | No | 8080 | Server port |
| LOG_LEVEL | No | info | Log level |
| {VAR} | {yes/no} | {default} | {description} |

---

## Metrics

### Build Metrics

| Metric | Value |
|--------|-------|
| Binary Size | {size} MB |
| Docker Image | {size} MB |
| Build Time | {time} |

### Quality Metrics

| Metric | Value |
|--------|-------|
| Test Coverage | {percentage}% |
| Tests Passing | {count}/{total} |
| Lint Issues | 0 |
| Race Conditions | 0 |

---

## Known Issues

- {Issue 1} - {Workaround if any}
- {Issue 2} - {Workaround if any}

---

## Upgrade Notes

### From v{previous} to v{current}

1. {Step 1}
2. {Step 2}
3. {Step 3}

---

## Contributors

- PM Agent - Requirements
- Architect Agent - Design
- Go Coder Agent - Implementation
- Test Agent - Testing
- Reviewer Agent - Code Review
- Optimizer Agent - Performance
- DevOps Agent - Release

---

## Links

- [Documentation]({link})
- [API Reference]({link})
- [Issue Tracker]({link})
