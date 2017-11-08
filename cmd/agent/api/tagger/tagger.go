// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2017 Datadog, Inc.

// Package agent implements the api endpoints for the `/agent` prefix.
// This group of endpoints is meant to provide high-level functionalities
// at the agent level.
package tagger

import (
	"encoding/json"
	"fmt"
	"net/http"

	apiutil "github.com/DataDog/datadog-agent/pkg/api/util"
	"github.com/DataDog/datadog-agent/pkg/tagger"
	"github.com/DataDog/dd-go/log"

	"github.com/gorilla/mux"
)

// SetupHandlers adds the specific handlers for /agent endpoints
func SetupHandlers(r *mux.Router) {
	r.HandleFunc("/tags/{container}", containerTagsHandler).Methods("GET")
}

func containerTagsHandler(w http.ResponseWriter, r *http.Request) {
	if err := apiutil.Validate(w, r); err != nil {
		return
	}
	vars := mux.Vars(r)
	container := vars["container"]
	tags, err := tagger.Tag(container, true)

	if err != nil {
		log.Errorf(err.Error())
	}

	if len(tags) == 0 {
		err := fmt.Errorf("bad url or container not monitored")
		log.Errorf("%s", err.Error())
		http.Error(w, err.Error(), 404)
		return
	}

	jsonTags, err := json.Marshal(tags)
	if err != nil {
		log.Errorf("Error marshalling tag list. Error: %v, tags: %v", err, tags)
		body, _ := json.Marshal(map[string]string{"error": err.Error()})
		http.Error(w, string(body), 500)
		return
	}

	w.Write(jsonTags)
}
