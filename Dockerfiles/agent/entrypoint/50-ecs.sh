#!/bin/bash

# Set a default config for ECS if found
# Don't override /etc/datadog-agent/datadog.yaml if it exists

if [[ "${ECS}" ]]; then
    if [ ! -e /etc/datadog-agent/datadog.yaml ]; then
        ln -s /etc/datadog-agent/datadog-ecs.yaml /etc/datadog-agent/datadog.yaml
    fi
fi
