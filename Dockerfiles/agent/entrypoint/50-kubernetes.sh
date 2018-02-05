#!/bin/bash

# Set a default config for Kubernetes if found
# Don't override /etc/datadog-agent/datadog.yaml if it exists

if [[ "${KUBERNETES_SERVICE_PORT}" ]]; then
    export KUBERNETES="yes"
fi

if [[ "${KUBERNETES}" ]]; then
    if [ ! -e /etc/datadog-agent/datadog.yaml ]; then
        ln -s /etc/datadog-agent/datadog-kubernetes.yaml /etc/datadog-agent/datadog.yaml
    fi
fi
