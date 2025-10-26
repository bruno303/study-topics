#!/bin/bash

# Backup Verifier Runner Script
# This script loads environment variables and runs the backup verifier

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Load environment variables if .env exists
if [ -f "$SCRIPT_DIR/.env" ]; then
    echo "Loading environment from .env..."
    set -a  # Enable automatic export of variables
    source "$SCRIPT_DIR/.env"
    set +a  # Disable automatic export
fi

# Check if required variables are set
if [ -z "$EMAIL_USER" ] || [ -z "$EMAIL_PASS" ]; then
    echo "❌ Error: EMAIL_USER and EMAIL_PASS must be set"
    echo "   Create .env.local file with your credentials or set environment variables"
    echo "   See .env.example for reference"
    exit 1
fi

# Run the verifier with all passed arguments
echo "🔍 Running backup verifier..."
cd "$SCRIPT_DIR"
node verifier.js "$@"
#node verifier.js --test
#node verifier.js --date=2025-10-25
