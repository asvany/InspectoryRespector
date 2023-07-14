package ir_protocol

import "testing"

func Test_init(t *testing.T) {
	windowChange := &WindowChange{}
	if windowChange.Location != nil {
		t.Error("windowChange.Location should be nil")
	}
}
