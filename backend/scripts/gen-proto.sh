#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

PROTO_DIR="$ROOT_DIR/proto"
OUT_DIR="$ROOT_DIR/proto"

command -v protoc >/dev/null 2>&1 || {
  echo "protoc not found" >&2
  exit 1
}
command -v protoc-gen-go >/dev/null 2>&1 || {
  echo "protoc-gen-go not found" >&2
  exit 1
}
command -v protoc-gen-go-grpc >/dev/null 2>&1 || {
  echo "protoc-gen-go-grpc not found" >&2
  exit 1
}

protoc \
  -I "$PROTO_DIR" \
  --go_out="$OUT_DIR" \
  --go_opt=paths=source_relative \
  --go-grpc_out="$OUT_DIR" \
  --go-grpc_opt=paths=source_relative \
  "$PROTO_DIR/auth/v1/auth.proto"

echo "Proto generated: auth/v1"
