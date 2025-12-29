#!/usr/bin/env bash
# common.sh - Shared utilities for go-dev-agent tools
#
# This file contains common functions used across multiple scripts.
# Source this file in other scripts: source "$(dirname "$0")/common.sh"

# Exit codes (meaningful, not arbitrary)
readonly EXIT_SUCCESS=0
readonly EXIT_USER_ERROR=1
readonly EXIT_SYSTEM_ERROR=2

# Colors (optional, respect NO_COLOR env var)
init_colors() {
    if [[ "${NO_COLOR:-}" == "1" ]] || [[ ! -t 1 ]]; then
        COLOR_GREEN=""
        COLOR_BLUE=""
        COLOR_YELLOW=""
        COLOR_RED=""
        COLOR_RESET=""
    else
        COLOR_GREEN='\033[0;32m'
        COLOR_BLUE='\033[0;34m'
        COLOR_YELLOW='\033[1;33m'
        COLOR_RED='\033[0;31m'
        COLOR_RESET='\033[0m'
    fi
}

# Logging functions - single responsibility, clear intent
log_info() {
    echo -e "${COLOR_BLUE:-}[INFO]${COLOR_RESET:-} $1"
}

log_success() {
    echo -e "${COLOR_GREEN:-}[PASS]${COLOR_RESET:-} $1"
}

log_warn() {
    echo -e "${COLOR_YELLOW:-}[WARN]${COLOR_RESET:-} $1"
}

log_error() {
    echo -e "${COLOR_RED:-}[FAIL]${COLOR_RESET:-} $1" >&2
}

# Box drawing - reusable UI component
print_box_header() {
    local title="$1"
    local width="${2:-64}"

    echo "╔$(printf '═%.0s' $(seq 1 $width))╗"
    printf "║ %-$((width-2))s ║\n" "$title"
    echo "╚$(printf '═%.0s' $(seq 1 $width))╝"
}

# Text normalization
normalize_text() {
    echo "$1" | tr '[:upper:]' '[:lower:]'
}

# Script directory detection
get_script_dir() {
    cd "$(dirname "${BASH_SOURCE[1]}")" && pwd
}

# Configuration file loader
load_config() {
    local config_file="$1"

    if [[ ! -f "$config_file" ]]; then
        log_error "Config file not found: $config_file"
        return $EXIT_USER_ERROR
    fi

    # Read config, skip comments and empty lines
    grep -v '^#' "$config_file" | grep -v '^$'
}

# Initialize colors on source
init_colors
