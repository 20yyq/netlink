// @@
// @ Author       : Eacher
// @ Date         : 2023-09-12 14:06:55
// @ LastEditTime : 2023-09-14 10:22:45
// @ LastEditors  : Eacher
// @ --------------------------------------------------------------------------------<
// @ Description  : 
// @ --------------------------------------------------------------------------------<
// @ FilePath     : /20yyq/netlink/socket.go
// @@
package netlink

import (
	"os"
	"syscall"
)

type conn struct {
	f		*os.File
	rawConn	syscall.RawConn
}

func newConn(domain, typ, proto int) (*conn, error) {
	fd, err := syscall.Socket(domain, typ, proto)
	if err != nil {
		return nil, err
	}
	if err = syscall.SetNonblock(fd, true); err != nil {
		return nil, err
	}
	conn := &conn{f: os.NewFile(uintptr(fd), "netlink_dev")}
	if conn.rawConn, err = conn.f.SyscallConn(); err != nil {
		conn.f.Close()
		return nil, err
	}
	return conn, nil
}

func (c *conn) control(f func(uintptr)) error {
	return c.rawConn.Control(f)
}

func (c *conn) read(f func(uintptr)bool) error {
	return c.rawConn.Read(f)
}

func (c *conn) write(f func(uintptr)bool) error {
	return c.rawConn.Write(f)
}
