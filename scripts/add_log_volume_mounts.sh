#!/bin/bash

# Script to add volume mounts to Kubernetes YAML files for message logging
# This script adds hostPath volume mounts to all service deployments
# so that logs written to /var/log/arpc-messages inside containers
# appear in /users/xzhu/online-boutique-arpc/logs on the host machine

# Services to update
SERVICES="ad cart checkout currency email frontend payment productcatalog recommendation shipping"

# Directories to update
DIRS="kubernetes/apply kubernetes/apply-basic kubernetes/apply-proxy kubernetes/apply-reliable kubernetes/apply-reliable-cc kubernetes/apply-reliable-cc-fc kubernetes/apply-reliable-cc-fc-encryption"

for dir in $DIRS; do
  for service in $SERVICES; do
    file="/users/xzhu/online-boutique-arpc/$dir/$service.yaml"
    if [ -f "$file" ]; then
      # Check if volumeMounts already exists
      if ! grep -q "volumeMounts:" "$file"; then
        # Find the line with "containerPort:" and add volumeMounts after it
        sed -i '/containerPort:.*$/a\        volumeMounts:\n        - name: message-logs\n          mountPath: /var/log/arpc-messages\n      volumes:\n      - name: message-logs\n        hostPath:\n          path: /users/xzhu/online-boutique-arpc/logs\n          type: DirectoryOrCreate' "$file"
        echo "Updated: $file"
      else
        echo "Skipped (already has volumeMounts): $file"
      fi
    fi
  done
done

echo ""
echo "Volume mounts added successfully!"
echo "Logs will be written to: /users/xzhu/online-boutique-arpc/logs"
