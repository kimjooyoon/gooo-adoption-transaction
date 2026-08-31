#!/usr/bin/env bash
set -euo pipefail

bin=${1:?path to the built gooo-adoption-transaction binary is required}
work=$(mktemp -d /tmp/gooo-adoption-transaction-conformance.XXXXXX)
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/output"
"$bin" run --source examples/adoption-transaction-v1/transaction.gooo --contract contracts/denominator-v1.json --out "$work/output"
test "$(find "$work/output" -type f | wc -l | tr -d ' ')" = 7
jq -e '.summary == {generated:12,closed:3,unknown:3,refuted:6}' "$work/output/transaction-manifest.json" >/dev/null
