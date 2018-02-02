// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package stages

import (
	"sync"
)

type parallelGroup struct {
	stages []Stage
}

func NewParallelGroup() Group {
	return &parallelGroup{
		stages: []Stage{},
	}
}

func (g *parallelGroup) Add(stages ...Stage) {
	g.stages = append(g.stages, stages...)
}

func (g *parallelGroup) Start() {
	for _, stage := range g.stages {
		go stage.Start()
	}
}

func (g *parallelGroup) Stop() {
	wg := &sync.WaitGroup{}
	for _, stage := range g.stages {
		wg.Add(1)
		go func(s Stage) {
			s.Stop()
			wg.Done()
		}(stage)
	}
	wg.Wait()
}
