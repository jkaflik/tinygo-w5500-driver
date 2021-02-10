package wiznet

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"time"

	"github.com/jkaflik/tinygo-w5500-driver/wiznet/net"
)

// Socket w5500
type Socket struct {
	io.ReadCloser
	io.WriteCloser
	d *Device

	bufferSize uint16 // todo: should be moved outside
	Num        uint8

	receivedSize  uint16
	receiveOffset uint16
}

func (s *Socket) initBuffers() error {
	// buffer initialisation should be moved outside probably
	if s.bufferSize == 0 {
		s.bufferSize = 2048
	}

	err := s.d.WriteByte(s.getRegister(), SocketReceiveBufferSize, byte(s.bufferSize>>10))
	err = s.d.WriteByte(s.getRegister(), SocketTransmitBufferSize, byte(s.bufferSize>>10))

	return err
}

func (s *Socket) Open(protocol string, port uint16) error {
	if err := s.initBuffers(); err != nil {
		return err
	}

	if status := s.Status(); status != SocketClosedStatus {
		log.Printf("open on non-closed socket: 0x%x\n\r", status)

		if err := s.Close(); err != nil {
			return err
		}
	}

	mode, err := newSocketModeFromProtocol(protocol)
	if err != nil {
		return err
	}
	if err := s.writeMode(mode); err != nil {
		return err
	}
	if err := s.clearInterrupt(); err != nil {
		return err
	}

	if port == 0 { // pick random local port instead
		port = 49152 + uint16(rand.Intn(16383))
	}

	if err := s.d.Write16(s.getRegister(), SocketSourcePort, port); err != nil {
		return err
	}

	return s.execCommand(socketOpenCommand)
}

func (s *Socket) Close() error {
	return s.execCommand(socketCloseCommand)
}

func (s *Socket) Listen() error {
	status := s.Status()
	if status != SocketInitStatus {
		return fmt.Errorf("listen on uninitalized socket: 0x%x", status)
	}

	return s.execCommand(socketListenCommand)
}

func (s *Socket) Connect(ip net.IP, port uint16) error {
	status := s.Status()
	if status != SocketInitStatus {
		return fmt.Errorf("connect on uninitalized socket: 0x%x", status)
	}

	if err := s.d.Write(s.getRegister(), SocketDesintationIPAddress, ip); err != nil {
		return err
	}
	if err := s.d.Write16(s.getRegister(), SocketDestinationPort, port); err != nil {
		return err
	}

	if err := s.execCommand(socketConnectCommand); err != nil {
		return err
	}

	start := time.Now()
	for {
		status := s.Status()
		switch status {
		case SocketClosedStatus:
			return errors.New("connect failed")
		case SocketCloseWaitStatus:
		case SocketEstablishedStatus:
			return nil
		}

		if time.Now().Sub(start) > time.Second*5 { // todo: timeout 5s as a configurable paramater
			return errors.New("connect timeout") // todo: better errors
		}

		time.Sleep(time.Millisecond)
	}
}

func (s *Socket) Disconnect() error {
	return s.execCommand(socketDisconnectCommand)
}

type SocketStatus uint8

const (
	SocketClosedStatus      SocketStatus = 0x00
	SocketInitStatus                     = 0x13
	SocketListenStatus                   = 0x14
	SocketSYNSENTStatus                  = 0x15
	SocketRECVStatus                     = 0x16
	SocketEstablishedStatus              = 0x17
	SocketFinWaitStatus                  = 0x18
	SocketClosingStatus                  = 0x1A
	SocketTimeWaitStatus                 = 0x1B
	SocketCloseWaitStatus                = 0x1C
	SocketLastAckStatus                  = 0x1D
	SocketUDPStatus                      = 0x22
	SocketMACRAWStatus                   = 0x02
)

func (s *Socket) Status() SocketStatus {
	b, _ := s.d.ReadByte(s.getRegister(), SocketStatusRegister)
	return SocketStatus(b)
}

func (s *Socket) getRegister() uint8 {
	return s.Num*4 + 1
}

type socketMode uint8

const (
	socketClosedMode         socketMode = 0x00
	socketTCPProtocolMode               = 0x21
	socketUDPProtocolMode               = 0x02
	socketMacRAWProtocolMode            = 0x04
)

func newSocketModeFromProtocol(protocol string) (socketMode, error) {
	switch protocol {
	case "tcp":
		return socketTCPProtocolMode, nil
	case "udp":
		return socketUDPProtocolMode, nil
	}

	return 0, fmt.Errorf("unsupported protocol: %s", protocol)
}

func (s *Socket) clearInterrupt() error {
	return s.d.WriteByte(s.getRegister(), SocketInterruptRegister, 0xff)
}

type socketInterrupt uint8

const (
	socketSendOkInterrupt  socketInterrupt = 0x10
	socketTimeoutInterrupt                 = 0x08
	socketRecvInterrupt                    = 0x04
	socketDisconInterrupt                  = 0x02
	socketConInterupt                      = 0x01
)

func (s *Socket) readInterrupt() socketInterrupt {
	b, _ := s.d.ReadByte(s.getRegister(), SocketInterruptRegister)
	return socketInterrupt(b)
}

func (s *Socket) writeInterrupt(i socketInterrupt) error {
	return s.d.WriteByte(s.getRegister(), SocketInterruptRegister, byte(i))
}

func (s *Socket) readMode() (socketMode, error) {
	b, err := s.d.ReadByte(s.getRegister(), SocketModeRegister)
	return socketMode(b), err
}

func (s *Socket) writeMode(mode socketMode) error {
	return s.d.WriteByte(s.getRegister(), SocketModeRegister, byte(mode))
}

type socketCmd uint8

const (
	socketOpenCommand       socketCmd = 0x01
	socketListenCommand               = 0x02
	socketConnectCommand              = 0x04
	socketDisconnectCommand           = 0x08
	socketCloseCommand                = 0x10
	socketSendCommand                 = 0x20
	socketSendMacCommand              = 0x21
	socketSendKeepCommand             = 0x22
	socketRecvCommand                 = 0x40
)

func (s *Socket) execCommand(cmd socketCmd) error {
	if err := s.d.WriteByte(s.getRegister(), SocketCommandRegister, byte(cmd)); err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		b, err := s.d.ReadByte(s.getRegister(), SocketCommandRegister)

		if err != nil {
			return err
		}

		if b == 0x0 {
			return nil
		}

		if socketCmd(b) != cmd {
			return fmt.Errorf("invalid command set: 0x%x", b)
		}

		time.Sleep(time.Millisecond)
	}

	return errors.New("socket command timeout")
}

func (s *Socket) read(start, len uint16) ([]byte, error) {
	start &= s.bufferSize - 1
	address := uint16(s.Num)*2048 + 0xC000 + start

	return s.d.Read(s.Num*4+3, address, len)
}

func (s *Socket) Read(p []byte) (n int, err error) {
	l := uint16(len(p))
	if l == 0 {
		return 0, nil
	}

	availableSize, err := s.d.Read16(s.getRegister(), SocketReceiveReceivedSize)
	if err != nil {
		return 0, err
	}

	if availableSize == 0 {
		status := s.Status()

		if status == SocketListenStatus || status == SocketClosedStatus || status == SocketCloseWaitStatus {
			return 0, io.EOF
		}

		return 0, nil
	}

	if availableSize > l {
		availableSize = l
	}

	pointer, err := s.d.Read16(s.getRegister(), SocketReceiveReadPointer)
	if err != nil {
		return 0, err
	}

	buf, err := s.read(pointer, availableSize)
	if err != nil {
		return 0, err
	}

	copy(p, buf) // ?

	pointer += availableSize

	if err := s.d.Write16(s.getRegister(), SocketReceiveReadPointer, pointer); err != nil {
		return int(availableSize), err
	}

	if err := s.execCommand(socketRecvCommand); err != nil {
		return int(availableSize), err
	}

	return int(availableSize), nil
}

func (s *Socket) write(data []byte) error {
	pointer, err := s.d.Read16(s.getRegister(), SocketTransmitWritePointer)
	if err != nil {
		return err
	}

	// offset := pointer & (s.bufferSize - 1)
	// address := offset + (uint16(s.Num)*s.bufferSize + 0x4000)

	address := pointer + uint16(s.Num)*s.bufferSize

	if err := s.d.Write(s.Num*4+2, address, data); err != nil {
		return err
	}

	pointer += uint16(len(data))
	return s.d.Write16(s.getRegister(), SocketTransmitWritePointer, pointer)
}

func (s *Socket) Write(p []byte) (n int, err error) {
	l := len(p)
	if l > int(s.bufferSize) { // todo split buffer into chunks
		return 0, fmt.Errorf("%d exceeds available socket buffer (%d)", l, s.bufferSize)
	}

	status := s.Status()
	if status != SocketCloseWaitStatus && status != SocketEstablishedStatus {
		return 0, fmt.Errorf("write on closed socket (0x%.2x)", status)
	}

	var freeSize uint16
	for freeSize < uint16(l) { // wait until buffer has enough space
		freeSize, err = s.d.Read16(s.getRegister(), SocketTransmitFreeSize)
		if err != nil {
			return 0, err
		}

		status := s.Status()
		if status != SocketCloseWaitStatus && status != SocketEstablishedStatus {
			return 0, io.EOF // todo: better error
		}

		time.Sleep(time.Millisecond)
	}

	if err := s.write(p); err != nil {
		return 0, err
	}

	if err := s.execCommand(socketSendCommand); err != nil {
		return 0, err
	}

	for (s.readInterrupt() & socketSendOkInterrupt) != socketSendOkInterrupt {
		status := s.Status()
		if status == SocketClosedStatus {
			return 0, errors.New("socket closed during write")
		}

		time.Sleep(time.Millisecond)
	}

	return l, nil
}
