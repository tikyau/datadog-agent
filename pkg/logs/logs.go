// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package logs

import (
	log "github.com/cihub/seelog"

	"github.com/DataDog/datadog-agent/pkg/logs/auditor"
	"github.com/DataDog/datadog-agent/pkg/logs/config"
	"github.com/DataDog/datadog-agent/pkg/logs/input/container"
	"github.com/DataDog/datadog-agent/pkg/logs/input/listener"
	"github.com/DataDog/datadog-agent/pkg/logs/input/tailer"
	"github.com/DataDog/datadog-agent/pkg/logs/message"
	"github.com/DataDog/datadog-agent/pkg/logs/pipeline"
	"github.com/DataDog/datadog-agent/pkg/logs/restart"
	"github.com/DataDog/datadog-agent/pkg/logs/sender"
	"github.com/DataDog/datadog-agent/pkg/logs/status"
)

var (
	// isRunning indicates whether logs-agent is running or not
	isRunning bool
	// logs-agent data pipeline
	agentPipeline restart.Group
)

// Start starts logs-agent
func Start() error {
	err := config.Build()
	if err != nil {
		return err
	}
	go run()
	return nil
}

// run sets up the pipeline to process logs and send them to Datadog back-end
func run() {
	isRunning = true

	connectionManager := sender.NewConnectionManager(
		config.LogsAgent.GetString("log_dd_url"),
		config.LogsAgent.GetInt("log_dd_port"),
		config.LogsAgent.GetBool("dev_mode_no_ssl"),
	)

	messageChan := make(chan message.Message, config.ChanSize)
	auditor := auditor.New(messageChan)
	pipelineProvider := pipeline.NewProvider(connectionManager, messageChan)

	sources := config.GetLogsSources()

	networkListeners := listener.New(sources.GetValidSources(), pipelineProvider)
	containersScanner := container.New(sources.GetValidSources(), pipelineProvider, auditor)
	filesScanner := tailer.New(
		sources.GetValidSources(),
		config.LogsAgent.GetInt("log_open_files_limit"),
		pipelineProvider,
		auditor,
		tailer.DefaultSleepDuration,
	)

	restart.Start(auditor, pipelineProvider, networkListeners, containersScanner, filesScanner)
	status.Initialize(sources.GetSources())

	inputs := restart.NewParallelGroup(filesScanner, containersScanner, networkListeners)
	agentPipeline = restart.NewSerialGroup(inputs, pipelineProvider, auditor)
}

// Stop stops properly the logs-agent to prevent data loss
// All Stop methods are blocking which means that Stop only returns
// when the whole pipeline is flushed
func Stop() {
	if isRunning {
		log.Info("Stopping logs-agent")
		agentPipeline.Stop()
	}
}

// GetStatus returns logs-agent status
func GetStatus() status.Status {
	if !isRunning {
		return status.Status{IsRunning: false}
	}
	return status.Get()
}
