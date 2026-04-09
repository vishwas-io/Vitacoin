#!/usr/bin/env python3
"""
Test VITAPAY payment flow via direct tx construction.
Since CLI commands don't exist for custom module, we construct txs manually.
"""
import subprocess
import json
import time
import sys

VM_CMD = "gcloud compute ssh vitacoin-testnet --zone=asia-south1-c --command"

def ssh(cmd):
    result = subprocess.run(
        f'{VM_CMD}=\'{cmd}\'',
        shell=True, capture_output=True, text=True, timeout=30
    )
    return result.stdout + result.stderr

def get_supply():
    out = ssh('curl -s http://localhost:1317/cosmos/bank/v1beta1/supply')
    # Parse from output
    for line in out.split('\n'):
        if '"amount"' in line and '"denom"' not in line:
            return line.strip().strip('"').strip(',').split('"')[1]
    return None

print("This script requires CLI tx commands which don't exist yet.")
print("Need to write vitacoin module CLI commands first.")
