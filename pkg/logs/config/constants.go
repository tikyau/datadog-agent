// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package config

// Pipeline constraints
const (
	ChanSize          = 100
	NumberOfPipelines = 4
)

// Default open files limit
const (
	DefaultTailingLimit = 100
)

// Date and time format
const (
	DateFormat = "2006-01-02T15:04:05.000000000Z"
)

// Severities
var (
	SevInfo  = []byte("<46>")
	SevError = []byte("<43>")
)
