// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

package health

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterAndUnhealthy(t *testing.T) {
	cat := newCatalog()
	token := cat.register("test1")

	_, found := cat.components[token]
	require.True(t, found)

	status := cat.getStatus()
	assert.Len(t, status.Healthy, 0)
	assert.Len(t, status.Unhealthy, 1)
	assert.Contains(t, status.Unhealthy, "test1")
}

func TestRegisterTriplets(t *testing.T) {
	cat := newCatalog()
	cat.register("triplet")
	cat.register("triplet")
	cat.register("triplet")
	assert.Len(t, cat.components, 3)

}

func TestDeregister(t *testing.T) {
	cat := newCatalog()
	token1 := cat.register("test1")
	token2 := cat.register("test2")

	assert.Len(t, cat.components, 2)

	err := cat.deregister(token1)
	assert.NoError(t, err)
	assert.Len(t, cat.components, 1)
	assert.Contains(t, cat.components, token2)
}

func TestDeregisterBadToken(t *testing.T) {
	cat := newCatalog()
	token1 := cat.register("test1")

	assert.Len(t, cat.components, 1)

	err := cat.deregister(nil)
	assert.NotNil(t, err)
	assert.Len(t, cat.components, 1)
	assert.Contains(t, cat.components, token1)
}

func TestGetHealthy(t *testing.T) {
	cat := newCatalog()
	token := cat.register("test1")

	status := cat.getStatus()
	assert.Len(t, status.Healthy, 0)
	assert.Len(t, status.Unhealthy, 1)

	for i := 1; i < 10; i++ {
		cat.pingComponents()
		<-token.C
	}

	status = cat.getStatus()
	assert.Len(t, status.Healthy, 1)
	assert.Len(t, status.Unhealthy, 0)
}

func TestUnhealthyAndBack(t *testing.T) {
	cat := newCatalog()
	token := cat.register("test1")

	status := cat.getStatus()
	assert.Len(t, status.Healthy, 0)
	assert.Len(t, status.Unhealthy, 1)

	for i := 1; i < 10; i++ {
		cat.pingComponents()
	}

	status = cat.getStatus()
	assert.Len(t, status.Healthy, 0)
	assert.Len(t, status.Unhealthy, 1)

	for i := 1; i < 10; i++ {
		cat.pingComponents()
		<-token.C
	}

	status = cat.getStatus()
	assert.Len(t, status.Healthy, 1)
	assert.Len(t, status.Unhealthy, 0)
}
