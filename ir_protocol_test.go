package main

import (
	"testing"
	"time"

	"github.com/asvany/InspectoryRespector/common"
	"github.com/asvany/InspectoryRespector/ir_protocol"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_timestamp(t *testing.T) {
	common.InitEnv("")

	now := time.Now()
	ts := timestamppb.New(now)

	message := &ir_protocol.WindowChange{
		Timestamp: ts,
	}

	t.Log("Message:", message)

}
