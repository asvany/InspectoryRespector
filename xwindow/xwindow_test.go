package xwindow

import (
	"testing"
)

func TestMain_xwindow(t *testing.T) {
	xi, err := NewXInfo()
	if err != nil {
		t.Error(err)

	}
	defer xi.conn.Close()

	key, err := xi.getFullKey()
	if err != nil {
		t.Error(err)
	}

	t.Logf("key:%v", key)
}
