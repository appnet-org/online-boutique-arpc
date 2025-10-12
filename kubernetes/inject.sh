#!/bin/bash

mkdir -p apply-proxy

for f in apply/*.yaml; do
    echo "Injecting $f..."
    python /users/xzhu/arpc/cmd/proxy/injector/symphony-injector.py -f "$f" > "apply-proxy/$(basename "$f")"
done
