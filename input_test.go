package main

import (
	"testing"

	"github.com/asvany/InspectoryRespector/common"
)

func Test_InputList(t *testing.T) {
	common.InitEnv("")

	devices := InputList()
	for _, device := range devices {
		t.Log(device)
	}

}
