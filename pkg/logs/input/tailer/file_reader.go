// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2018 Datadog, Inc.

// +build !windows

package tailer

import (
	"io"
	"os"
	"syscall"

	"github.com/DataDog/datadog-agent/pkg/logs/decoder"
	log "github.com/cihub/seelog"
)

// readForever lets the tailer tail the content of a file
// until it is closed.
func (t *Tailer) readForever() {
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
			t.wait()
			continue
		}
		if err != nil {
			t.source.Tracker.TrackError(err)
			log.Error("Err: ", err)
			return
		}
		if n == 0 {
			t.wait()
			continue
		}
		t.d.InputChan <- decoder.NewInput(inBuf[:n])
		t.incrementReadOffset(n)
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
	stat2, err := t.file.Stat()
	if err != nil {
		return true, nil
	}
	if inode(stat1) != inode(stat2) {
		return true, nil
	}

	if stat1.Size() < t.GetReadOffset() {
		return true, nil
	}
	return false, nil
}

// inode uniquely identifies a file on a filesystem
func inode(f os.FileInfo) uint64 {
	s := f.Sys()
	if s == nil {
		return 0
	}
	switch s := s.(type) {
	case *syscall.Stat_t:
		return uint64(s.Ino)
	default:
		return 0
	}
}
