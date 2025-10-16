#!/usr/bin/env bash

set -e

###############################################################################
###                       Protocol Buffer Generation                        ###
###############################################################################

echo "🔧 Generating protocol buffer files..."

# Get the project root directory
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$PROJECT_ROOT"

# Install protoc-gen-gocosmos if not already installed
if ! command -v protoc-gen-gocosmos &> /dev/null; then
    echo "📦 Installing protoc-gen-gocosmos..."
    go install github.com/cosmos/gogoproto/protoc-gen-gocosmos@latest
fi

# Install protoc-gen-grpc-gateway if not already installed
if ! command -v protoc-gen-grpc-gateway &> /dev/null; then
    echo "📦 Installing protoc-gen-grpc-gateway..."
    go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@latest
fi

# Setup proto path
COSMOS_SDK_DIR=$(go list -f '{{ .Dir }}' -m github.com/cosmos/cosmos-sdk 2>/dev/null)
GOGO_PROTO_DIR=$(go list -f '{{ .Dir }}' -m github.com/cosmos/gogoproto 2>/dev/null)
GOOGLEAPIS_DIR=$(go list -f '{{ .Dir }}' -m github.com/googleapis/googleapis 2>/dev/null || echo "")

echo "� Running protoc..."

# Generate proto files
protoc \
  --proto_path="${PROJECT_ROOT}/proto" \
  --proto_path="${COSMOS_SDK_DIR}/proto" \
  --proto_path="${GOGO_PROTO_DIR}" \
  --gocosmos_out=plugins=grpc,\
Mgoogle/protobuf/any.proto=github.com/cosmos/cosmos-sdk/codec/types:. \
  --grpc-gateway_out=logtostderr=true,allow_colon_final_segments=true:. \
  $(find "${PROJECT_ROOT}/proto" -name "*.proto")

# Move generated files to types directory
echo "📁 Organizing generated files..."
mkdir -p x/vitacoin/types
if [ -d "github.com/vitacoin/vitacoin/x/vitacoin/types" ]; then
    mv github.com/vitacoin/vitacoin/x/vitacoin/types/*.go x/vitacoin/types/ 2>/dev/null || true
    rm -rf github.com
fi

echo "✅ Protocol buffer generation complete!"
echo ""
echo "Generated files location: x/vitacoin/types/"
