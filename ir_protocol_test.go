package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/asvany/InspectoryRespector/ir_protocol"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func Test_timestamp(t *testing.T) {

	now := time.Now()
	ts := timestamppb.New(now)

	message := &ir_protocol.WindowChange{
		Timestamp: ts,
	}

	fmt.Println("Message:", message)

}
