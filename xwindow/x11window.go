package xwindow

import (
	"fmt"
	"log"
	"strings"
	// "time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type XInfo struct {
	conn       *xgb.Conn
	screen     xproto.ScreenInfo
	rootWin    xproto.Window
	desktopVP  xproto.Atom
}

func NewXInfo() (*XInfo, error) {
	conn, err := xgb.NewConn()
	if err != nil {
		return nil, err
	}

	screen := xproto.Setup(conn).DefaultScreen(conn)
	rootWin := screen.Root
	desktopVP, err := xproto.InternAtom(conn, true, uint16(len("_NET_DESKTOP_VIEWPORT")), "_NET_DESKTOP_VIEWPORT").Reply()
	if err != nil {
		return nil, err
	}

	return &XInfo{
		conn:       conn,
		screen:     screen,
		rootWin:    rootWin,
		desktopVP:  desktopVP.Atom,
	}, nil
}

func (xi *XInfo) getViewPort() (string, error) {
	reply, err := xproto.GetProperty(xi.conn, false, xi.rootWin, xi.desktopVP, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		return "", err
	}

	if reply.Format != 32 || reply.Type != xproto.AtomCARDINAL {
		return "", fmt.Errorf("unexpected property format or type")
	}

	data := reply.Value
	if len(data) != 8 {
		return "", fmt.Errorf("unexpected property value length")
	}

	x := int32(data[0]) | int32(data[1])<<8 | int32(data[2])<<16 | int32(data[3])<<24
	y := int32(data[4]) | int32(data[5])<<8 | int32(data[6])<<16 | int32(data[7])<<24

	return fmt.Sprintf("WM_VIEWPORT_%d", (x/1920)+(y/1200)*4), nil
}

func (xi *XInfo) getActiveWindata() (string, error) {
	reply, err := xproto.GetProperty(xi.conn, false, xi.rootWin, xproto.AtomWmClass, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		return "", err
	}

	if reply.Format != 8 || reply.Type != xproto.AtomSTRING {
		return "", fmt.Errorf("unexpected property format or type")
	}

	cls := strings.ToUpper(string(reply.Value))
	if cls == "" {
		focus, err := xproto.QueryTree(xi.conn, xi.rootWin).Reply()
		if err != nil {
			return "", err
		}
		if len(focus.Children) > 0 {
			focusWindow := focus.Children[0]
			reply, err := xproto.GetProperty(xi.conn, false, focusWindow, xproto.AtomWmClass, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
			if err != nil {
				return "", err
			}
			if reply.Format == 8 && reply.Type == xproto.AtomSTRING {
				cls = strings.ToUpper(string(reply.Value))
			}
		}
	}

	reply, err = xproto.GetProperty(xi.conn, false, xi.rootWin, xproto.AtomWmName, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		return "", err
	}

	if reply.Format != 8 || reply.Type != xproto.AtomSTRING {
		return "", fmt.Errorf("unexpected property format or type")
	}

	name := string(reply.Value)

	var builder strings.Builder

	if cls != "" {
		builder.WriteString("WMCLASS_")
		builder.WriteString(cls)
		builder.WriteRune(' ')
	} else {
		builder.WriteString("CLS_NONE")
	}

	builder.WriteString(name)

	return builder.String(), nil
}

func (xi *XInfo) getFullKey() (string, error) {
	var builder strings.Builder

	viewPort, err := xi.getViewPort()
	if err != nil {
		return "", err
	}
	builder.WriteString(viewPort)
	builder.WriteRune(' ')

	activeWindata, err := xi.getActiveWindata()
	if err != nil {
		return "", err
	}
	builder.WriteString(activeWindata)

	return builder.String(), nil
}

func main() {
	xi, err := NewXInfo()
	if err != nil {
		log.Fatal(err)
	}
	defer xi.conn.Close()

	key, err := xi.getFullKey()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(key)
}
