package xwindow

import (
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"sort"

	// "time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/xproto"
)

type WinProps map[string]string

type XInfo struct {
	conn              *xgb.Conn
	screen            xproto.ScreenInfo
	rootWin           xproto.Window
	desktopVP         xproto.Atom
	utf8StringAtom    xproto.Atom
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
		"WM_NAME":  {},
		"WM_CLASS": {},
		// "WM_CLIENT_MACHINE": {},
		"_NET_WM_NAME":          {},
		"_NET_DESKTOP_VIEWPORT": {},
		"_NET_WORKAREA":         {},
		// "_NET_WM_PID":      {},
		"_NET_WM_DESKTOP": {},
		// "_NET_WM_STATE":    {},
		"WM_WINDOW_ROLE": {},
	}

	return &XInfo{
		conn:              conn,
		screen:            *screen,
		rootWin:           rootWin,
		desktopVP:         desktopVP.Atom,
		utf8StringAtom:    utf8StringAtom.Atom,
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

func (xi *XInfo) Close() {
	xi.conn.Close()
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

func (xi *XInfo) getActiveWindowData(propValues *WinProps) error {
	atomActiveWindow, _ := xproto.InternAtom(xi.conn, false, uint16(len("_NET_ACTIVE_WINDOW")), "_NET_ACTIVE_WINDOW").Reply()
	activeWindowProp, err := xproto.GetProperty(xi.conn, false, xi.rootWin, atomActiveWindow.Atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()

	if err != nil {
		return err
	}

	activeWindowID := xproto.Window(xgb.Get32(activeWindowProp.Value))

	props, err := xproto.ListProperties(xi.conn, activeWindowID).Reply()
	if err != nil {
		return err
	}

	for _, atom := range props.Atoms {

		nameReply, err := xproto.GetAtomName(xi.conn, atom).Reply()
		if err != nil {
			return err
		}
		if _, ok := xi.propertyWhitelist[nameReply.Name]; !ok {
			continue
		}

		valueReply, err := xproto.GetProperty(xi.conn, false, activeWindowID, atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()
		if err != nil {
			return err
		}

		(*propValues)[nameReply.Name] = string(valueReply.Value)
	}

	pid, err := xi.getPIDForWindow(activeWindowID)
	if err != nil {
		return err
	}

	(*propValues)["_NET_WM_PID"] = fmt.Sprintf("%d", pid)

	getPIDInfoProps(pid, propValues)

	return nil
}

func getPIDInfoProps(pid uint32, propValues *WinProps) error {

	//get the name of the process

	cmdline, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return err
	}

	(*propValues)["cmdline"] = string(cmdline)

	// read the value of the process's cwd symlink
	cwd, err := os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid))
	if err != nil {
		return err
	}
	(*propValues)["cwd"] = cwd

	//get the exe of the process
	exe, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid))
	if err != nil {
		return err
	}
	(*propValues)["exe"] = exe

	return nil
}

func (xi *XInfo) getPIDForWindow(window xproto.Window) (uint32, error) {
	atom, err := xproto.InternAtom(xi.conn, false, uint16(len("_NET_WM_PID")), "_NET_WM_PID").Reply()
	if err != nil {
		return 0, err
	}

	//get the pid of the window
	reply, err := xproto.GetProperty(xi.conn, false, window, atom.Atom, xproto.GetPropertyTypeAny, 0, (1<<32)-1).Reply()

	if err != nil {
		return 0, err
	}

	if len(reply.Value) < 4 {
		return 0, fmt.Errorf("invalid _NET_WM_PID property length")
	}

	return binary.LittleEndian.Uint32(reply.Value), nil
}


// return all key value pairs in a sorted order as a string with string builder for easy appending
func getStringDataSorted(propValues *WinProps) string {
	var sb strings.Builder

	var keys []string
	for k := range *propValues {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for  _,k := range keys {
		sb.WriteString(fmt.Sprintf("%s=%s ", k, (*propValues)[k]))
	}
	return sb.String()
}

func (xi *XInfo) GetFullKey(propValues *WinProps) (string, error) {

	viewPort, err := xi.getViewPort()
	if err != nil {
		return fmt.Sprintf("%v", err), err
	}
	(*propValues)["WM_VIEWPORT"] = viewPort

	err = xi.getActiveWindowData(propValues)
	if err != nil {
		return fmt.Sprintf("%v", err), err
	}

	fp := getStringDataSorted(propValues)

	return fp, nil
}

