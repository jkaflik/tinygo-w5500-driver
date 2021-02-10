package wiznet

const (
	CommonRegister = 0b00000

	Socket0Register = 0b00001
	Socket0TxBuffer = 0b00010
	Socket0RxBuffer = 0b00011

	Socket1Register = 0b00101
	Socket1TxBuffer = 0b00110
	Socket1RxBuffer = 0b00111

	Socket2Register = 0b01001
	Socket2TxBuffer = 0b01010
	Socket2RxBuffer = 0b01011

	Socket3Register = 0b01101
	Socket3TxBuffer = 0b01110
	Socket3RxBuffer = 0b01111

	Socket4Register = 0b10001
	Socket4TxBuffer = 0b10010
	Socket4RxBuffer = 0b10011

	Socket5Register = 0b10101
	Socket5TxBuffer = 0b10110
	Socket5RxBuffer = 0b10111

	Socket6Register = 0b11001
	Socket6TxBuffer = 0b11010
	Socket6RxBuffer = 0b11011

	Socket7Register = 0b11101
	Socket7TxBuffer = 0b11110
	Socket7RxBuffer = 0b11111
)

const (
	ReadAccessMode  uint8 = 0x00 << 2
	WriteAccessMode       = 0x01 << 2

	VariableLengthMode  = 0x00
	FixedLengthOneMode  = 0x01
	FixedLengthTwoMode  = 0x02
	FixedLengthFourMode = 0x03
)

const (
	ModeRegister                   uint16 = 0x0000 // 1 byte
	GatewayIPAddressRegister              = 0x0001 // 4 bytes
	SubnetMaskRegister                    = 0x0005 // 4 bytes
	HardwareAddressRegister               = 0x0009 // 6 bytes
	IPAddressRegister                     = 0x000F // 4 bytes
	InterruptLowLevelTimerRegister        = 0x0013 // 2 bytes
	InterruptRegister                     = 0x0015 // 1 byte
	InterruptMaskRegister                 = 0x0016 // 1 byte
	// SocketInterruptRegister               = 0x0017 // 1 byte
	SocketInterruptMaskRegister = 0x0018 // 1 byte
	PHYConfigurationRegister    = 0x002E // 1 byte
	VersionRegister             = 0x0039 // 1 byte
)

const (
	SocketModeRegister               uint16 = 0x0000
	SocketCommandRegister                   = 0x0001
	SocketInterruptRegister                 = 0x0002
	SocketStatusRegister                    = 0x0003
	SocketSourcePort                        = 0x0004 // 2 bytes
	SocketDestinationHardwareAddress        = 0x0006 // 6 bytes
	SocketDesintationIPAddress              = 0x000C // 4 bytes
	SocketDestinationPort                   = 0x0010 // 2 bytes
	SocketMaximumSegmentSize                = 0x0012 // 2 bytes
	SocketIPTOS                             = 0x0015
	SocketIPTTL                             = 0x0016
	SocketReceiveBufferSize                 = 0x001E
	SocketTransmitBufferSize                = 0x001F
	SocketTransmitFreeSize                  = 0x0020 // 2 bytes
	SocketTransmitReadPointer               = 0x0022 // 2 bytes
	SocketTransmitWritePointer              = 0x0024 // 2 bytes
	SocketReceiveReceivedSize               = 0x0026 // 2 bytes
	SocketReceiveReadPointer                = 0x0028 // 2 bytes
	SocketReceiveWritePointer               = 0x002A // 2 bytes
	SocketInterruptMask                     = 0x002C
	SocketFragmentOffsetInIPHeader          = 0x002D // 2 bytes
	KeepaliveTimer                          = 0x002F
)
