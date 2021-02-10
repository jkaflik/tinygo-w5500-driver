package main

import (
	"bytes"
	"io"
	"log"
	"machine"
	"time"

	"github.com/jkaflik/tinygo-w5500-driver/wiznet"
	"github.com/jkaflik/tinygo-w5500-driver/wiznet/net"
)

func countdown(message string, duration time.Duration, tick time.Duration) {
	for rem := duration; rem >= 0; rem -= tick {
		log.Printf(message, rem.String())
		time.Sleep(tick)
	}
}

func main() {
	countdown("Starting in %s...\n", time.Second*2, time.Second)

	device := &wiznet.Device{
		Bus:       machine.SPI0,
		SelectPin: machine.D10,
	}
	if err := device.Configure(); err != nil {
		log.Printf("device configure err: %v\n", err)
		return
	}

	device.SetHardwareAddr(net.HardwareAddr{0xDE, 0xAD, 0xBE, 0xEF, 0xFE, 0xED})
	device.SetIPAddress(net.IP{192, 168, 1, 177})
	device.SetGatewayAddress(net.IP{192, 168, 1, 1})
	device.SetSubnetMask(net.IPMask{255, 255, 255, 0})

	socket := device.NewSocket()

	response, err := makeRequest(
		socket,
		net.IP{172, 67, 133, 228},
		80,
		"GET / HTTP/1.1 \r\nHost: ifconfig.co\r\nUser-Agent: curl/7.64.1\r\nAccept: */*\r\nConnection: close\r\n\r\n",
	)
	if err != nil {
		log.Printf("make request failed: %v\n", err)
	} else {
		log.Println(response)
	}

	blinkLed()
}

func makeRequest(socket *wiznet.Socket, ip net.IP, port uint16, request string) (string, error) {
	log.Printf("open tcp socket\n\r")
	if err := socket.Open("tcp", 80); err != nil {
		return "", err
	}

	defer socket.Close()

	log.Println("connect to ifconfig.co")
	if err := socket.Connect(ip, port); err != nil {
		return "", err
	}

	log.Println("write HTTP request")
	if _, err := io.WriteString(socket, request); err != nil {
		return "", err
	}

	log.Println("read HTTP response")

	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, socket); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func blinkLed() {
	led := machine.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	for {
		led.Low()
		time.Sleep(time.Millisecond * 500)
		led.High()
		time.Sleep(time.Millisecond * 500)
	}
}
