// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

// +build windows

package tailer

import (
	"io"
	"os"

	"github.com/DataDog/datadog-agent/pkg/logs/decoder"
	log "github.com/cihub/seelog"
)

// readForever lets the tailer tail the content of a file
// until it is closed.
func (t *Tailer) readForever() {
	// For windows, this means reading the file until the end
	for {
		if t.shouldHardStop() {
			t.onStop()
			return
		}

		inBuf := make([]byte, 4096)
		n, err := t.file.Read(inBuf)
		if err == io.EOF {
			if t.shouldSoftStop() {
				t.onStop()
				return
			}
			t.file.Close()
			go t.readWhenNeeded()
			return
		}
		if err != nil {
			t.source.Tracker.TrackError(err)
			log.Error("Err: ", err)
			t.onStop()
			return
		}
		if n == 0 {
			t.file.Close()
			go t.readWhenNeeded()
			return
		}
		t.d.InputChan <- decoder.NewInput(inBuf[:n])
		t.incrementReadOffset(n)
	}
}

func (t *Tailer) readWhenNeeded() {
	for {
		t.wait()

		f, err := os.Open(t.path)
		if err != nil {
			f.Close()
			continue
		}
		stat, err := f.Stat()
		if err != nil {
			f.Close()
			continue
		}
		if stat.Size() > t.GetReadOffset() {
			f.Seek(t.GetReadOffset(), os.SEEK_SET)
			t.file = f
			go t.readForever()
		} else {
			f.Close()
			continue
		}
	}
}

func (t *Tailer) checkForRotation() (bool, error) {
	f, err := os.Open(t.path)
	defer f.Close()
	if err != nil {
		return false, err
	}
	stat1, err := f.Stat()
	if err != nil {
		return false, err
	}

	if stat1.Size() < t.GetReadOffset() {
		return true, nil
	}
	return false, nil
}
