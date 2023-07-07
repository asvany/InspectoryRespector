package geo

import (
	"testing"

	"github.com/asvany/InspectoryRespector/ir_protocol"
)

func TestGetLocation(t *testing.T) {
	loc_chan := make(chan *ir_protocol.Location)
	go GetLocation(loc_chan)
	loc := <-loc_chan

	t.Logf("Loc: %s\n", loc.String())
}
