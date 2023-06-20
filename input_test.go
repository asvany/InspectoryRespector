package main

import (
	"testing"
)

func Test_InputList(t *testing.T) {
	devices := InputList()
	for _, device := range devices {
		t.Log(device)
	}

}
