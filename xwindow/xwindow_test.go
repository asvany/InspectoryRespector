package xwindow

import (
	"testing"
)

func TestMain_xwindow(t *testing.T) {
	xi, err := NewXInfo()
	if err != nil {
		t.Error(err)

	}
	defer xi.Close()
	propValues := make(WinProps)

	fp,err := xi.GetFullKey(&propValues)
	if err != nil {
		t.Error(err)
	}

	t.Logf("key:%v", propValues)

	t.Logf("fp:%v", fp)

}
