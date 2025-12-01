# Kubernetes Deployment Generator

This directory contains scripts to generate customized Kubernetes deployment manifests with different feature configurations for the Online Boutique application.

## Overview

The `generate_apply.sh` script creates deployment folders with custom environment variables to control:
- **Reliable Delivery** (ENABLE_RELIABLE)
- **Congestion Control** (ENABLE_CC)
- **Flow Control** (ENABLE_FC)

## Usage

### Basic Usage

```bash
./generate_apply.sh [OPTIONS]
```

### Options

| Option | Description | Default |
|--------|-------------|---------|
| `--reliable=true\|false` | Enable/disable reliable delivery | `true` |
| `--cc=true\|false` | Enable/disable congestion control | `true` |
| `--fc=true\|false` | Enable/disable flow control | `true` |
| `--output-dir=name` | Custom output directory name | Auto-generated |
| `--help` | Show help message | - |

