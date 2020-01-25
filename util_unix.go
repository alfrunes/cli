package cli

// +build aix darwin dragonfly freebsd linux netbsd openbsd solaris

import "fmt"
import "golang.org/x/sys/unix"

const NewLine = "\n"

func getTerminalSize(fd int) (heightWidth [2]uint16, err error) {
	ws, err := unix.IoctlGetWinsize(fd, unix.TIOCGWINSZ)
	if err != nil {
		return [2]uint16{0, 0}, err
	}
	return [2]uint16{ws.Col, ws.Row}, nil
}

func elemInSlice(elem interface{}, slice []interface{}) bool {
	for _, e := range slice {
		if elem == e {
			return true
		}
	}
	return false
}

func joinSlice(slice []interface{}, sep string) string {
	var ret string
	lastIdx := len(slice) - 1
	for i, e := range slice {
		ret += fmt.Sprintf("%v", e)
		if i != lastIdx {
			ret += sep
		}
	}
	return ret
}
