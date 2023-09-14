// @@
// @ Author       : Eacher
// @ Date         : 2023-09-12 08:12:21
// @ LastEditTime : 2023-09-14 08:18:49
// @ LastEditors  : Eacher
// @ --------------------------------------------------------------------------------<
// @ Description  : 
// @ --------------------------------------------------------------------------------<
// @ FilePath     : /20yyq/netlink/message.go
// @@
package netlink

import (
	"fmt"
	"syscall"

	"github.com/20yyq/packet"
)

const ReceiveDataSize = 1024

type SendMessage struct {
	*packet.NlMsghdr
	sa		syscall.Sockaddr

	Err		error
	Data 	[]byte
}

func (sm *SendMessage) sendto(fd uintptr) bool {
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

type ReceiveMessage struct {
	Data 	[]byte
	Idx 	int
	Err		error
	MsgList	[]*packet.NetlinkMessage
	Sa		syscall.Sockaddr
}

func (rm *ReceiveMessage) recvfrom(fd uintptr) bool {
	if rm.Idx, rm.Sa, rm.Err = syscall.Recvfrom(int(fd), rm.Data, 0); rm.Err != nil {
		return false
	}
	rm.MsgList = packet.NewNetlinkMessage(rm.Data[:rm.Idx])
	return true
}
