package commands

import (
	"bytes"
	"strings"
	"testing"

	"github.com/halqme/blackboard/internal/cli/commandkit"
	"github.com/halqme/blackboard/internal/config"
	"github.com/halqme/blackboard/internal/store"
)

func TestCmdStatusPrintsTaskCreationGuidanceWhenNoActiveTask(t *testing.T) {
	out := &bytes.Buffer{}
	prev := commandkit.Out
	commandkit.Out = out
	defer func() { commandkit.Out = prev }()

	if err := CmdStatus(nil, store.State{Revision: 1}); err != nil {
		t.Fatalf("CmdStatus() error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "Current task: none") || !strings.Contains(got, `bb task new "<title>"`) {
		t.Fatalf("output = %q", got)
	}
}

func TestCmdNextSuggestsTaskCreationWhenNoActiveTask(t *testing.T) {
	out := &bytes.Buffer{}
	prev := commandkit.Out
	commandkit.Out = out
	defer func() { commandkit.Out = prev }()

	if err := CmdNext(nil, store.State{Revision: 1}); err != nil {
		t.Fatalf("CmdNext() error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, "No active task.") || !strings.Contains(got, `bb task new "<title>"`) {
		t.Fatalf("output = %q", got)
	}
}

func TestCmdStatusJSONSuggestsTaskCreationWhenNoActiveTask(t *testing.T) {
	out := &bytes.Buffer{}
	prev := commandkit.Out
	commandkit.Out = out
	defer func() { commandkit.Out = prev }()

	if err := CmdStatus([]string{"--json"}, store.State{Revision: 1}); err != nil {
		t.Fatalf("CmdStatus() error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, `"next_command": "bb task new \"\u003ctitle\u003e\""`) || !strings.Contains(got, `"has_active_task": false`) {
		t.Fatalf("output = %q", got)
	}
}

func TestCmdContextJSONSuggestsTaskCreationWhenNoActiveTask(t *testing.T) {
	out := &bytes.Buffer{}
	prev := commandkit.Out
	commandkit.Out = out
	defer func() { commandkit.Out = prev }()

	if err := CmdContext([]string{"--json"}, ".", config.Config{}, store.State{Revision: 1}); err != nil {
		t.Fatalf("CmdContext() error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, `"next_command": "bb task new \"\u003ctitle\u003e\""`) || !strings.Contains(got, `"has_active_task": false`) {
		t.Fatalf("output = %q", got)
	}
}

func TestCmdNextJSONSuggestsTaskCreationWhenNoActiveTask(t *testing.T) {
	out := &bytes.Buffer{}
	prev := commandkit.Out
	commandkit.Out = out
	defer func() { commandkit.Out = prev }()

	if err := CmdNext([]string{"--json"}, store.State{Revision: 1}); err != nil {
		t.Fatalf("CmdNext() error = %v", err)
	}
	got := out.String()
	if !strings.Contains(got, `"next_command": "bb task new \"\u003ctitle\u003e\""`) {
		t.Fatalf("output = %q", got)
	}
}
