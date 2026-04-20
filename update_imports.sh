#!/bin/bash
find ./vitacoin -name "*.go" -type f -exec sed -i ".bak" "s|github.com/vishwas-io/VITACOIN/vitacoin|vitacoin|g" {} \;
find ./vitacoin -name "*.go" -type f -exec sed -i ".bak" "s|github.com/vishwas-io/VITACOIN|vitacoin|g" {} \;
find ./vitacoin -name "*.go" -type f -exec sed -i ".bak" "s|github.com/vitacoin/vitacoin|vitacoin|g" {} \;
echo "Import paths updated to use local references"
