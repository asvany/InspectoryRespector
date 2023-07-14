package xwindow

import (
	"testing"

	"github.com/asvany/InspectoryRespector/common"
)

func TestMain_xwindow(t *testing.T) {
	common.InitEnv("../")
	
	xi, err := NewXInfo()
	if err != nil {
		t.Error(err)

	}
	defer xi.Close()
	propValues := make(WinProps)

	fp, err := xi.GetFullKey(&propValues)
	if err != nil {
		t.Error(err)
	}

	t.Logf("key:%v", propValues)

	t.Logf("fp:%v", fp)

}
