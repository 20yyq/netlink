// @@
// @ Author       : Eacher
// @ Date         : 2023-09-12 08:12:21
// @ LastEditTime : 2023-09-14 10:30:34
// @ LastEditors  : Eacher
// @ --------------------------------------------------------------------------------<
// @ Description  : 
// @ --------------------------------------------------------------------------------<
// @ FilePath     : /20yyq/netlink/message.go
// @@
package netlink

import (
	"fmt"
	"time"
	"syscall"

	"github.com/20yyq/packet"
)

const ReceiveDataSize = 1024

type SendNLMessage struct {
	*packet.NlMsghdr
	sa		syscall.Sockaddr

	Err		error
	Data 	[]byte
}

func (sm *SendNLMessage) sendto(fd uintptr) bool {
	if sm.Len < uint32(packet.SizeofNlMsghdr + len(sm.Data)) {
		sm.Err = fmt.Errorf("data len error")
		return false
	}
	b := make([]byte, sm.Len)
	sm.WireFormatToByte((*[packet.SizeofNlMsghdr]byte)(b))
	copy(b[packet.SizeofNlMsghdr:], sm.Data)
	if err := syscall.Sendto(int(fd), b, 0, sm.sa); err != nil {
		sm.Err = err
		return false
	}
	return true
}

type ReceiveNLMessage struct {
	Data 		[]byte
	Idx 		int
	Err			error
	MsgList		[]*packet.NetlinkMessage
	Sa			syscall.Sockaddr
	Exchange	func(uintptr)bool
	OutTime 	time.Duration
}

func (rm *ReceiveNLMessage) recvfrom(fd uintptr) bool {
	if rm.Idx, rm.Sa, rm.Err = syscall.Recvfrom(int(fd), rm.Data, 0); rm.Err != nil {
		return false
	}
	rm.MsgList = packet.NewNetlinkMessage(rm.Data[:rm.Idx])
	return true
}
