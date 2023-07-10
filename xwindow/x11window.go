package xwindow

import (
	"fmt"
	"strings"

	// "time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type XInfo struct {
	conn           *xgb.Conn
	screen         xproto.ScreenInfo
	rootWin        xproto.Window
	desktopVP      xproto.Atom
	utf8StringAtom xproto.Atom
	propertyWhitelist map[string]struct{}
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

	utf8StringAtom, err := xproto.InternAtom(conn, false, uint16(len("UTF8_STRING")), "UTF8_STRING").Reply()
	if err != nil {
		return nil, err
	}

	propertyWhitelist := map[string]struct{}{
		"WM_NAME":          {},
		"WM_CLASS":         {},
		// "WM_CLIENT_MACHINE": {},
		"_NET_WM_NAME":     {},
		"_NET_DESKTOP_VIEWPORT":{},
		"_NET_WORKAREA": {},
		// "_NET_WM_PID":      {},
		"_NET_WM_DESKTOP":  {},
		// "_NET_WM_STATE":    {},
		"WM_WINDOW_ROLE":{},
	}

	return &XInfo{
		conn:           conn,
		screen:         *screen,
		rootWin:        rootWin,
		desktopVP:      desktopVP.Atom,
		utf8StringAtom: utf8StringAtom.Atom,
		propertyWhitelist: propertyWhitelist,
	}, nil
}
func (xi *XInfo) CheckStringReply(reply *xproto.GetPropertyReply, err error) (*xproto.GetPropertyReply, error) {
	if err != nil {
		return reply, err
	}
	if reply.Format != 8 || reply.Type != xproto.AtomString && reply.Type != xi.utf8StringAtom {
		return reply, fmt.Errorf("unexpected property format or type")
	}
	return reply, err
}

func (xi *XInfo) getViewPort() (string, error) {
	reply, err := xproto.GetProperty(xi.conn, false, xi.rootWin, xi.desktopVP, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
	if err != nil {
		return "", err
	}

	if reply.Format != 32 || reply.Type != xproto.AtomCardinal {
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
	atomActiveWindow, _ := xproto.InternAtom(xi.conn, false, uint16(len("_NET_ACTIVE_WINDOW")), "_NET_ACTIVE_WINDOW").Reply()
	activeWindowProp, err := xproto.GetProperty(xi.conn, false, xi.rootWin, atomActiveWindow.Atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()

	if err != nil {
		return "", err
	}

	activeWindowID := xproto.Window(xgb.Get32(activeWindowProp.Value))

	props, err := xproto.ListProperties(xi.conn, activeWindowID).Reply()
	if err != nil {
		// handle error
	}

	propValues := make(map[string]string)

	for _, atom := range props.Atoms {
		
		nameReply, err := xproto.GetAtomName(xi.conn, atom).Reply()
		if err != nil {
			// handle error
		}
		if _, ok := xi.propertyWhitelist[nameReply.Name]; !ok {
			continue
		}

		valueReply, err := xproto.GetProperty(xi.conn, false, activeWindowID, atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
		if err != nil {
			// handle error
		}

		propValues[nameReply.Name] = string(valueReply.Value)
	}

	reply, err := xi.CheckStringReply(xproto.GetProperty(xi.conn, false, activeWindowID, xproto.AtomWmClass, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply())
	if err != nil {
		return "", err

	}

	cls :=string(reply.Value)
	

	reply, err = xi.CheckStringReply(xproto.GetProperty(xi.conn, false, activeWindowID, xproto.AtomWmName, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply())
	if err != nil {
		return "", err

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
