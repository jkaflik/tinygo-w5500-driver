package wiznet

import (
	"errors"
	"machine"
	"time"

	"github.com/jkaflik/tinygo-w5500-driver/wiznet/net"
)

const socketNum uint8 = 8

// Device is a W5500 chip interface
type Device struct {
	Bus       machine.SPI
	SelectPin machine.Pin
}

// Configure SPI interface and select pin
func (w *Device) Configure() error {
	time.Sleep(time.Millisecond * 600)

	if err := w.Bus.Configure(machine.SPIConfig{
		LSBFirst:  false,
		Mode:      machine.Mode0,
		Frequency: 8000000,
	}); err != nil {
		return err
	}

	w.SelectPin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	w.SelectPin.High()

	if !w.isW5500() {
		return errors.New("could not detect W5500 chip")
	}

	return nil
}

func (w *Device) isW5500() bool {
	if err := w.Reset(); err != nil {
		return false
	}

	w.WriteByte(CommonRegister, ModeRegister, 0x08)
	if r, _ := w.ReadByte(CommonRegister, ModeRegister); r != 0x08 {
		return false
	}

	w.WriteByte(CommonRegister, ModeRegister, 0x10)
	if r, _ := w.ReadByte(CommonRegister, ModeRegister); r != 0x10 {
		return false
	}

	w.WriteByte(CommonRegister, ModeRegister, 0x0)
	if r, _ := w.ReadByte(CommonRegister, ModeRegister); r != 0x0 {
		return false
	}

	v, _ := w.ReadByte(CommonRegister, VersionRegister)

	return v == 0x4
}

func (w *Device) Reset() error {
	if err := w.WriteByte(CommonRegister, ModeRegister, 0x80); err != nil {
		return err
	}

	for i := 0; i < 20; i++ {
		b, err := w.ReadByte(CommonRegister, ModeRegister)
		if err != nil {
			return err
		}

		if b == 0x0 {
			return nil
		}

		time.Sleep(1 * time.Millisecond)
	}

	return errors.New("chip initialisation failed after soft reset")
}

func (w *Device) NewSocket() *Socket {
	return &Socket{
		Num: 0,
		d:   w,
	}
}

type PHYConfiguration byte

func (c PHYConfiguration) IsLinkUp() bool {
	return (c & 0x01) != 0
}

func (c PHYConfiguration) Is100MbpsLink() bool {
	return (c & 0x02) != 0
}

func (w *Device) GetPHYConfiguration() PHYConfiguration {
	b, _ := w.ReadByte(CommonRegister, PHYConfigurationRegister)
	return PHYConfiguration(b)
}

// SetIPAddress sets client IP address
func (w *Device) SetIPAddress(ip net.IP) error {
	return w.Write(CommonRegister, IPAddressRegister, ip)
}

// GetIPAddress reads client ip address from a W5500 chip
func (w *Device) GetIPAddress() net.IP {
	b, _ := w.Read(CommonRegister, IPAddressRegister, 4)
	return net.IP(b)
}

// SetGatewayAddress sets client IP address
func (w *Device) SetGatewayAddress(ip net.IP) error {
	return w.Write(CommonRegister, GatewayIPAddressRegister, ip)
}

// GetGatewayAddress reads client ip address from a W5500 chip
func (w *Device) GetGatewayAddress() net.IP {
	b, _ := w.Read(CommonRegister, GatewayIPAddressRegister, 4)
	return net.IP(b)
}

// SetSubnetMask sets subnet mask
func (w *Device) SetSubnetMask(mask net.IPMask) error {
	return w.Write(CommonRegister, SubnetMaskRegister, mask)
}

// GetSubnetMask reads subnet mask from a W5500 chip
func (w *Device) GetSubnetMask() net.IPMask {
	b, _ := w.Read(CommonRegister, SubnetMaskRegister, 4)
	return net.IPMask(b)
}

// SetIPNet sets gateway and net mask
func (w *Device) SetIPNet(net net.IPNet) error {
	if err := w.SetGatewayAddress(net.IP); err != nil {
		return err
	}
	if err := w.SetSubnetMask(net.Mask); err != nil {
		return err
	}

	return nil
}

// GetIPNet reads network from W5500 chip
func (w *Device) GetIPNet() net.IPNet {
	gw := w.GetGatewayAddress()
	mask := w.GetSubnetMask()

	return net.IPNet{IP: gw, Mask: mask}
}

// SetHardwareAddr sets a mac address on a W5500 chip
func (w *Device) SetHardwareAddr(mac net.HardwareAddr) error {
	return w.Write(CommonRegister, HardwareAddressRegister, mac)
}

// GetHardwareAddr reads a mac address from a W5500 chip
func (w *Device) GetHardwareAddr() net.HardwareAddr {
	b, _ := w.Read(CommonRegister, HardwareAddressRegister, 6)
	return net.HardwareAddr(b)
}

// Read reads a len of bytes
func (w *Device) Read(control uint8, address uint16, len uint16) ([]byte, error) {
	control <<= 3
	control |= ReadAccessMode | VariableLengthMode

	data := []byte{
		byte((address & 0xFF00) >> 8),
		byte((address & 0x00FF) >> 0),
		control,
	}

	w.SelectPin.Low()
	defer w.SelectPin.High()

	if err := w.Bus.Tx(data, nil); err != nil {
		return nil, err
	}

	buf := make([]byte, len)
	err := w.Bus.Tx(nil, buf)

	// log.Printf("read: %.8b %.16b 0x%x\n\r", control, address, buf)

	return buf, err
}

func (w *Device) ReadByte(control uint8, address uint16) (byte, error) {
	b, err := w.Read(control, address, 1)
	if err != nil {
		return 0, err
	}
	return b[0], nil
}

func (w *Device) Read16(control uint8, address uint16) (uint16, error) {
	b, err := w.Read(control, address, 2)
	if err != nil {
		return 0, err
	}

	return (uint16(b[0]) << 8) | uint16(b[1]), nil
}

// Write writes a buf slice
func (w *Device) Write(control uint8, address uint16, buf []byte) error {
	control <<= 3
	control |= WriteAccessMode | VariableLengthMode

	data := []byte{
		byte((address & 0xFF00) >> 8),
		byte((address & 0x00FF) >> 0),
		control,
	}

	w.SelectPin.Low()
	err := w.Bus.Tx(data, nil)
	err = w.Bus.Tx(buf, nil)
	w.SelectPin.High()

	// log.Printf("write: %.8b %.16b 0x%x\n\r", control, address, buf)

	return err
}

func (w *Device) WriteByte(control uint8, address uint16, buf byte) error {
	return w.Write(control, address, []byte{buf})
}

func (w *Device) Write16(control uint8, address uint16, buf uint16) error {
	return w.Write(control, address, []byte{byte(buf >> 8), byte(buf & 0xFF)})
}
