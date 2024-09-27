#!/bin/bash

# Set up logging
LOG_FILE="/logs/env_export.log"
mkdir -p "$(dirname "$LOG_FILE")"
exec > >(tee -a "$LOG_FILE") 2>&1

log() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"
}

log_error() {
    echo "[$(date +'%Y-%m-%d %H:%M:%S')] ERROR: $1" >&2
}

# Array to store exported variable names
declare -a exported_vars

# Function to export variables from a file
export_vars() {
    local file=$1
    if [ -f "$file" ]; then
        while IFS= read -r line || [[ -n "$line" ]]; do
            if [[ ! "$line" =~ ^# && -n "$line" ]]; then
                if export "$line"; then
                    # Extract variable name and add to array
                    var_name=$(echo "$line" | cut -d= -f1)
                    exported_vars+=("$var_name")
                else
                    log_error "Failed to export: $line"
                fi
            fi
        done < "$file"
        log "Exported variables from $file"
    else
        log_error "$file not found"
    fi
}

# Process tofu output
if command -v tofu &> /dev/null; then
    log "Processing tofu output..."
    if ! tofu_output=$(tofu -chdir=infrastructure/ output -json 2>&1); then
        log_error "Failed to get tofu output: $tofu_output"
    else
        if ! echo "$tofu_output" | jq -r 'to_entries | .[] | "export \(.key | ascii_upcase)=\(.value.value)"' > .tf-env; then
            log_error "Failed to process tofu output with jq"
        else
            log "Generated .tf-env file from tofu output"
        fi
    fi
else
    log_error "tofu command not found. Skipping tofu output processing."
fi

# Export variables from .env file
export_vars ".env"

# Export variables from .tf-env file
export_vars ".tf-env"

# Write only the exported variables to a file that can be sourced later
{
    for var in "${exported_vars[@]}"; do
        echo "${var}=${!var}"
    done
} | sort > "/logs/.exported_vars.env"

log "Environment variable export process completed. Variables written to /logs/.exported_vars.env"
