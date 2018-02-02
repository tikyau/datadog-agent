// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package stages

// Stage represents a stage of a data pipeline that can be started and stopped
type Stage interface {
	Start()
	Stop()
}

// Group represents a set of stages from a data pipeline
type Group interface {
	Stage
	Add(stages ...Stage)
}
