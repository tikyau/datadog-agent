// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package docker

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/DataDog/datadog-agent/pkg/aggregator/mocksender"
	"github.com/DataDog/datadog-agent/pkg/collector/check"
	"github.com/DataDog/datadog-agent/pkg/collector/corechecks/containers"
	"github.com/DataDog/datadog-agent/pkg/config"
	"github.com/DataDog/datadog-agent/pkg/tagger"
	"github.com/DataDog/datadog-agent/test/integration/utils"
)

var retryDelay = flag.Int("retry-delay", 1, "time to wait between retries (default 1 second)")
var retryTimeout = flag.Int("retry-timeout", 10, "maximum time before failure (default 10 seconds)")
var skipCleanup = flag.Bool("skip-cleanup", false, "skip cleanup of the docker containers (for debugging)")

// Must be repeated in the following dockerCfgString
const instanceTag = "instanceTag:MustBeHere"

var dockerCfgString = `
collect_events: true
collect_container_size: true
collect_images_stats: true
collect_exit_codes: true
tags:
  - instanceTag:MustBeHere
`

var datadogCfgString = `
docker_labels_as_tags:
    "high.card.label": +highcardlabeltag
    "low.card.label": lowcardlabeltag
docker_env_as_tags:
    "HIGH_CARD_ENV": +highcardenvtag
    "low_card_env": lowcardenvtag
`

var sender *mocksender.MockSender
var dockerCheck check.Check

func TestMain(m *testing.M) {
	flag.Parse()

	retryTicker := time.NewTicker(time.Duration(*retryDelay) * time.Second)
	timeoutTicker := time.NewTicker(time.Duration(*retryTimeout) * time.Second)
	var lastRunResult int
	var retryCount int

	err := setup()
	if err != nil {
		fmt.Printf("Test setup failed:\n%s\n", err.Error())
		tearOffAndExit(1)
	}

	for {
		select {
		case <-retryTicker.C:
			retryCount++
			if retryCount > 1 {
				fmt.Print("\n\n")
			}
			fmt.Printf("===== Starting run %d =====\n", retryCount)
			lastRunResult = doRun(m)
			if lastRunResult == 0 {
				tearOffAndExit(lastRunResult)
			}
		case <-timeoutTicker.C:
			fmt.Printf("\n\n===== FAILED after %d retries\n", retryCount)
			//fmt.Printf("Latest state:\n%s", sender.Mock.Calls)
			tearOffAndExit(lastRunResult)
		}
	}
}

// Called before for first test run: compose up
func setup() error {
	// Setup global conf
	config.Datadog.SetConfigType("yaml")
	err := config.Datadog.ReadConfig(strings.NewReader(datadogCfgString))
	if err != nil {
		return err
	}

	// Setup tagger
	err = tagger.Init()
	if err != nil {
		return err
	}

	// Start compose recipes
	for projectName, file := range defaultCatalog.composeFilesByProjects {
		compose := &utils.ComposeConf{
			ProjectName: projectName,
			FilePath:    file,
		}
		output, err := compose.Start()
		if err != nil {
			fmt.Println(string(output))
			return err
		}
	}
	return nil
}

// Reset the state and trigger a new run
func doRun(m *testing.M) int {
	// Setup docker check
	var dockerCfg = []byte(dockerCfgString)
	var dockerInitCfg = []byte("")
	dockerCheck = containers.DockerFactory()
	dockerCheck.Configure(dockerCfg, dockerInitCfg)

	// Setup mock sender
	sender = mocksender.NewMockSender(dockerCheck.ID())
	sender.SetupAcceptAll()

	dockerCheck.Run()
	return m.Run()
}

// Compose down, reset the external states and exit
func tearOffAndExit(exitcode int) {
	if *skipCleanup {
		os.Exit(exitcode)
	}

	// Stop compose recipes, ignore errors
	for projectName, file := range defaultCatalog.composeFilesByProjects {
		compose := &utils.ComposeConf{
			ProjectName: projectName,
			FilePath:    file,
		}
		output, err := compose.Stop()
		if err != nil {
			fmt.Println(string(output))
		}
	}
	os.Exit(exitcode)
}
