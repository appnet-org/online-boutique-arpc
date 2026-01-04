#!/bin/bash

# Script to generate Kubernetes deployment folders with custom feature flags
# Usage: ./generate_apply.sh [--reliable=true|false] [--cc=true|false] [--fc=true|false] [--encryption=true|false] [--output-dir=name]

set -e

# Default values (all features enabled except encryption)
ENABLE_RELIABLE="true"
ENABLE_CC="true"
ENABLE_FC="true"
ENABLE_ENCRYPTION="false"
OUTPUT_DIR=""

# Parse command line arguments
for arg in "$@"; do
    case $arg in
        --reliable=*)
            ENABLE_RELIABLE="${arg#*=}"
            shift
            ;;
        --cc=*)
            ENABLE_CC="${arg#*=}"
            shift
            ;;
        --fc=*)
            ENABLE_FC="${arg#*=}"
            shift
            ;;
        --encryption=*)
            ENABLE_ENCRYPTION="${arg#*=}"
            shift
            ;;
        --output-dir=*)
            OUTPUT_DIR="${arg#*=}"
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [--reliable=true|false] [--cc=true|false] [--fc=true|false] [--encryption=true|false] [--output-dir=name]"
            echo ""
            echo "Options:"
            echo "  --reliable=true|false   Enable/disable reliable delivery (default: true)"
            echo "  --cc=true|false         Enable/disable congestion control (default: true)"
            echo "  --fc=true|false         Enable/disable flow control (default: true)"
            echo "  --encryption=true|false Enable/disable encryption (default: false)"
            echo "  --output-dir=name       Custom output directory name (default: auto-generated)"
            echo ""
            echo "Examples:"
            echo "  $0 --reliable=false --cc=false --fc=false    # Basic aRPC only"
            echo "  $0 --reliable=true --cc=false --fc=false     # Only reliable delivery"
            echo "  $0 --cc=true --reliable=false --fc=false     # Only congestion control"
            echo "  $0 --output-dir=apply-custom                 # Custom directory name"
            exit 0
            ;;
        *)
            echo "Unknown option: $arg"
            echo "Use --help for usage information"
            exit 1
            ;;
    esac
done

# Determine output directory name if not specified
if [ -z "$OUTPUT_DIR" ]; then
    features=""
    if [ "$ENABLE_RELIABLE" = "true" ]; then
        features="${features}reliable-"
    fi
    if [ "$ENABLE_CC" = "true" ]; then
        features="${features}cc-"
    fi
    if [ "$ENABLE_FC" = "true" ]; then
        features="${features}fc-"
    fi
    if [ "$ENABLE_ENCRYPTION" = "true" ]; then
        features="${features}encryption-"
    fi
    
    if [ -z "$features" ]; then
        OUTPUT_DIR="apply-basic"
    else
        # Remove trailing dash
        features="${features%-}"
        OUTPUT_DIR="apply-${features}"
    fi
fi

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_DIR="${SCRIPT_DIR}/apply"
TARGET_DIR="${SCRIPT_DIR}/${OUTPUT_DIR}"

echo "=========================================="
echo "Generating Kubernetes deployment manifests"
echo "=========================================="
echo "Source directory: ${SOURCE_DIR}"
echo "Target directory: ${TARGET_DIR}"
echo ""
echo "Feature flags:"
echo "  ENABLE_RELIABLE:   ${ENABLE_RELIABLE}"
echo "  ENABLE_CC:         ${ENABLE_CC}"
echo "  ENABLE_FC:         ${ENABLE_FC}"
echo "  ENABLE_ENCRYPTION: ${ENABLE_ENCRYPTION}"
echo "=========================================="
echo ""

# Check if source directory exists
if [ ! -d "$SOURCE_DIR" ]; then
    echo "Error: Source directory ${SOURCE_DIR} does not exist"
    exit 1
fi

# Create target directory
if [ -d "$TARGET_DIR" ]; then
    echo "Warning: Target directory ${TARGET_DIR} already exists"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Aborting"
        exit 1
    fi
    rm -rf "$TARGET_DIR"
fi

mkdir -p "$TARGET_DIR"

# Function to add environment variables to a deployment
add_env_vars() {
    local input_file="$1"
    local output_file="$2"
    
    # Use awk to insert environment variables after the 'args:' line in Deployment resources
    awk -v reliable="$ENABLE_RELIABLE" -v cc="$ENABLE_CC" -v fc="$ENABLE_FC" -v encryption="$ENABLE_ENCRYPTION" '
    /kind: Deployment/ { in_deployment=1 }
    /^---$/ { in_deployment=0 }
    /args:/ && in_deployment {
        print
        # Get the current indentation
        match($0, /^[ \t]*/)
        indent = substr($0, RSTART, RLENGTH)
        # Print environment variables with proper indentation
        print indent "env:"
        print indent "- name: ENABLE_RELIABLE"
        print indent "  value: \"" reliable "\""
        print indent "- name: ENABLE_CC"
        print indent "  value: \"" cc "\""
        print indent "- name: ENABLE_FC"
        print indent "  value: \"" fc "\""
        print indent "- name: ENABLE_ENCRYPTION"
        print indent "  value: \"" encryption "\""
        next
    }
    { print }
    ' "$input_file" > "$output_file"
}

# Process each YAML file
echo "Processing YAML files..."
for yaml_file in "$SOURCE_DIR"/*.yaml; do
    if [ -f "$yaml_file" ]; then
        filename=$(basename "$yaml_file")
        echo "  - ${filename}"
        
        # Check if the file contains a Deployment resource
        if grep -q "kind: Deployment" "$yaml_file"; then
            add_env_vars "$yaml_file" "$TARGET_DIR/$filename"
        else
            # Copy files without Deployments as-is
            cp "$yaml_file" "$TARGET_DIR/$filename"
        fi
    fi
done

echo ""
echo "=========================================="
echo "✓ Successfully generated deployment manifests"
echo "=========================================="
echo ""
echo "Output directory: ${TARGET_DIR}"
echo ""
echo "To deploy:"
echo "  kubectl apply -Rf ${TARGET_DIR}"
echo ""
echo "To verify:"
echo "  kubectl get pods"
echo ""
echo "To clean up:"
echo "  kubectl delete pv,pvc,sa,all --all"
echo ""

