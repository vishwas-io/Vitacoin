#!/bin/bash

###############################################################################
###                    VitaCoin Development Environment Setup               ###
###############################################################################

# This script sets up the development environment for VitaCoin
# Run this before working on the project: source ./setup-env.sh

echo "🚀 Setting up VitaCoin development environment..."

# Set project root first
export VITACOIN_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Add local Go to PATH (prioritize project's Go over system Go)
export PATH="$VITACOIN_ROOT/go/bin:$PATH"
export PATH=$PATH:$(go env GOPATH)/bin

# Set Go environment variables
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org,direct

# Verify Go installation
if command -v go &> /dev/null; then
    echo "✅ Go $(go version | awk '{print $3}') is available"
    echo "   Go Path: $(which go)"
    echo "   GOPATH: $(go env GOPATH)"
else
    echo "❌ Go is not installed or not in PATH"
    exit 1
fi

# Verify protoc installation
if command -v protoc &> /dev/null; then
    echo "✅ protoc $(protoc --version | awk '{print $2}') is available"
else
    echo "⚠️  protoc is not installed"
fi

# Set project root
export VITACOIN_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$VITACOIN_ROOT"

echo ""
echo "📁 Project root: $VITACOIN_ROOT"
echo "✨ Environment ready!"
echo ""
echo "💡 Tip: Run 'make help' to see available commands"
