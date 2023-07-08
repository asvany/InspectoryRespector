package main

import (
	"fmt"
	"log"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type XInfo struct {
	conn *xgb.Conn
	root xproto.Window
}

func NewXInfo() (*XInfo, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return nil, err
	}

	root := xproto.Setup(conn).DefaultScreen(conn).Root

	return &XInfo{
		conn: conn,
		root: root,
	}, nil
}

func (xi *XInfo) GetActiveWindowData() (string, error) {
	reply, err := xproto.GetInputFocus(xi.conn).Reply()
	if err != nil {
		return "", err
	}

	focus := reply.Focus
	if focus == xproto.None {
		return "", fmt.Errorf("no window in focus")
	}

	name, err := xproto.GetAtomName(xi.conn, xproto.Atom(focus)).Reply()
	if err != nil {
		return "", err
	}

	return name.Name, nil
}

func main() {
	xi, err := NewXInfo()
	if err != nil {
		log.Fatal(err)
	}

	name, err := xi.GetActiveWindowData()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(name)
}
