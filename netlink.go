// @@
// @ Author       : Eacher
// @ Date         : 2023-09-11 11:04:00
// @ LastEditTime : 2023-09-14 09:48:00
// @ LastEditors  : Eacher
// @ --------------------------------------------------------------------------------<
// @ Description  : 
// @ --------------------------------------------------------------------------------<
// @ FilePath     : /20yyq/netlink/netlink.go
// @@
package netlink

import (
	"fmt"
	"io"
	"sync"
	"syscall"
)

type NetlinkRoute struct {
	*conn
	isExchange	bool
	mutex		sync.Mutex

	Err 		error
	DevName		string
	Sal			*syscall.SockaddrNetlink
}

// start NetlinkRoute socket
func (nlr *NetlinkRoute) Init() error {
	if nlr.conn != nil {
		return fmt.Errorf("socket busy")
	}
	nlr.conn, nlr.Err = newConn(syscall.AF_NETLINK, syscall.SOCK_RAW|syscall.SOCK_CLOEXEC, syscall.NETLINK_ROUTE)
	if nlr.Err == nil {
		if err := nlr.control(nlr.initBind); nlr.Err == nil {
			nlr.Err = err
		}
	}
	return nlr.Err
}

func (nlr *NetlinkRoute) Exchange(sm *SendNLMessage, rm *ReceiveNLMessage) error {
	nlr.mutex.Lock()
	defer func(old bool){
		nlr.isExchange = old; nlr.mutex.Unlock()
	}(nlr.isExchange)
	if nlr.isExchange {
		return fmt.Errorf("io busy")
	}
	nlr.isExchange, sm.sa = true, syscall.Sockaddr(nlr.Sal)
	if err := nlr.write(sm.sendto); sm.Err == nil {
		if sm.Err = err; sm.Err == nil {
			if err = nlr.read(rm.recvfrom); rm.Err == nil {
				rm.Err = err
			}
			return rm.Err
		}
	}
	return sm.Err
}

func (nlr *NetlinkRoute) Receive() (<-chan *ReceiveNLMessage, error) {
	nlr.mutex.Lock()
	defer nlr.mutex.Unlock()
	if nlr.isExchange {
		return nil, fmt.Errorf("io busy")
	}
	nlr.isExchange = true
	notify := make(chan *ReceiveNLMessage, 5)
	go nlr.receive(notify)
	return notify, nil
}

func (nlr *NetlinkRoute) Send(sm *SendNLMessage) error {
	nlr.mutex.Lock()
	var is bool
	is, sm.Err = nlr.isExchange, fmt.Errorf("io busy")
	nlr.mutex.Unlock()
	if is {
		sm.Err, sm.sa = nil, syscall.Sockaddr(nlr.Sal)
		if err := nlr.write(sm.sendto); sm.Err == nil {
			sm.Err = err
		}
	}
	return sm.Err
}

func (nlr *NetlinkRoute) receive(r chan<- *ReceiveNLMessage) {
	for {
		rm := &ReceiveNLMessage{Data: make([]byte, ReceiveDataSize)}
		if err := nlr.read(rm.recvfrom); rm.Err == nil {
			rm.Err = err
		}
		r <- rm
		if rm.Err == io.EOF {
			close(r)
			break
		}
	}
	nlr.isExchange = false
}

func (nlr *NetlinkRoute) initBind(fd uintptr) {
	if nlr.Err = syscall.BindToDevice(int(fd), nlr.DevName); nlr.Err != nil {
		return
	}
	// nlr.Sal = &syscall.SockaddrNetlink{Family: syscall.AF_NETLINK, Groups: syscall.RTNLGRP_LINK}
	if nlr.Err = syscall.Bind(int(fd), nlr.Sal); nlr.Err != nil {
		return
	}
}

func (nlr *NetlinkRoute) Close() error {
	f := func (fd uintptr) {
		nlr.Err = syscall.Close(int(fd))
	}
	if err := nlr.control(f); nlr.Err == nil {
		nlr.Err = err
	}
	return nlr.Err
}
