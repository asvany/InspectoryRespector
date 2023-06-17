package main

import (
	"fmt"
	"time"

	"github.com/asvany/InspectoryRespector/ir_protocol"
	"github.com/golang/protobuf/ptypes"
)

func main() {
	t := time.Now()
	ts, err := ptypes.TimestampProto(t)
	if err != nil {
		fmt.Println("timestamp creation error:", err)
		return
	}

	message := &ir_protocol.WindowChange{
		Timestamp: ts,
	}

	fmt.Println("Message:", message)
}
