// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package stages

type serialGroup struct {
	stages []Stage
}

func NewSerialGroup() Group {
	return &serialGroup{
		stages: []Stage{},
	}
}

func (g *serialGroup) Add(stages ...Stage) {
	g.stages = append(g.stages, stages...)
}

func (g *serialGroup) Start() {
	stagesLen := len(g.stages)
	for i := 0; i < len(g.stages); i++ {
		stage := g.stages[stagesLen-i-1]
		stage.Start()
	}
}

func (g *serialGroup) Stop() {
	for _, stage := range g.stages {
		stage.Stop()
	}
}
