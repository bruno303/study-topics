#!/bin/bash

# =================================================================
# Kubernetes Secret Key Updater
# Requirements: kubectl, yq (https://github.com/mikefarah/yq)
# Usage: ./update_secret.sh <SECRET_NAME> <NEW_KEY> <NEW_VALUE>
# =================================================================

set -e

if [ "$#" -ne 3 ]; then
    echo "Usage: $0 <SECRET_NAME> <NEW_KEY> <NEW_VALUE>"
    exit 1
fi

SECRET_NAME=$1
NEW_KEY=$2
RAW_VALUE=$3
TEMP_FILE="${SECRET_NAME}_temp_secret.yaml"

if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl command not found. Please ensure it is installed and in your PATH."
    exit 1
fi

if ! command -v yq &> /dev/null; then
    echo "Error: yq command not found. Please install yq (e.g., brew install yq or go install github.com/mikefarah/yq@latest)."
    exit 1
fi

CURRENT_NAMESPACE=$(kubectl config view --minify --output 'jsonpath={..namespace}' 2>/dev/null)
if [ -z "$CURRENT_NAMESPACE" ]; then
    CURRENT_NAMESPACE="default"
fi

echo ""
echo "========================================================"
echo "ACTION: Update Secret '$SECRET_NAME'"
echo "TARGET: Namespace '$CURRENT_NAMESPACE'"
echo "NEW KEY: '$NEW_KEY'"
echo "========================================================"

read -r -p "Do you want to continue with this configuration? (y/N): " response

if [[ ! "$response" =~ ^([yY][eE][sS]|[yY])$ ]]; then
    echo "Operation cancelled by user."
    exit 0
fi
echo ""

# Using the '-n' flag to prevent trailing newlines from being encoded
ENCODED_VALUE=$(echo -n "$RAW_VALUE" | base64)

if [ -z "$ENCODED_VALUE" ]; then
    echo "Error: Could not encode value. Exiting."
    exit 1
fi

echo "Encoded value for key '$NEW_KEY' generated successfully."

# --- 4. Retrieve Existing Secret and Clean Metadata ---
set +e
echo "Retrieving existing Secret '$SECRET_NAME'..."
kubectl get secret "$SECRET_NAME" -o yaml > "$TEMP_FILE" 2>/dev/null

if [ $? -ne 0 ]; then
    echo "Failed to retrieve secret '$SECRET_NAME'. Secret will be created."
    rm -f "$TEMP_FILE"

    kubectl create secret generic "$SECRET_NAME" \
      --from-literal="$NEW_KEY"="$RAW_VALUE"

    exit 1
fi
set -e

yq 'del(.metadata.resourceVersion, .metadata.uid, .metadata.annotations, .metadata.creationTimestamp, .status)' -i "$TEMP_FILE"

echo "Patching Secret YAML to add key '$NEW_KEY'..."
yq ".data[\"$NEW_KEY\"] = \"$ENCODED_VALUE\"" -i "$TEMP_FILE"

echo "Applying updated Secret definition..."
kubectl apply -f "$TEMP_FILE"

if [ $? -eq 0 ]; then
    echo "✅ Success! Secret '$SECRET_NAME' updated. Key '$NEW_KEY' added."
else
    echo "❌ Error: Failed to apply updated secret."
fi

rm -f "$TEMP_FILE"
echo "Temporary file '$TEMP_FILE' deleted."
