package geo


import (
	"testing"

)

func TestGetLocation(t *testing.T) {
	var loc Location
	GetLocation(&loc)
}
