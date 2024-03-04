// Ethernet over USB driver
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usbnet

import (
	"encoding/binary"
	"errors"
	"net"

	"github.com/usbarmory/tamago/soc/nxp/usb"

	"gvisor.dev/gvisor/pkg/buffer"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// NIC represents an virtual Ethernet instance.
type NIC struct {
	// Host MAC address
	HostMAC net.HardwareAddr

	// Device MAC address
	DeviceMAC net.HardwareAddr

	// Link is a gVisor channel endpoint
	Link *channel.Endpoint

	// Device is the physical interface associated to the virtual one.
	Device *usb.Device

	// Rx is endpoint 1 OUT function, set by Init() to ECMRx if not
	// already defined.
	Rx func([]byte, error) ([]byte, error)

	// Tx is endpoint 1 IN function, set by Init() to ECMTx if not already
	// defined.
	Tx func([]byte, error) ([]byte, error)

	// Control is endpoint 2 IN function, set by Init() to ECMControl if
	// not already defined.
	Control func([]byte, error) ([]byte, error)

	maxPacketSize int
	buf           []byte
}

// Init initializes a virtual Ethernet instance on a specific USB device and
// configuration index.
func (eth *NIC) Init() (err error) {
	if eth.Link == nil {
		return errors.New("missing link endpoint")
	}

	if len(eth.HostMAC) != 6 || len(eth.DeviceMAC) != 6 {
		return errors.New("invalid MAC address")
	}

	if eth.Rx == nil {
		eth.Rx = eth.ECMRx
	}

	if eth.Tx == nil {
		eth.Tx = eth.ECMTx
	}

	if eth.Control == nil {
		eth.Control = eth.ECMControl
	}

	addControlInterface(eth.Device, eth)
	addDataInterfaces(eth.Device, eth)

	return
}

// ECMControl implements the endpoint 2 IN function.
func (eth *NIC) ECMControl(_ []byte, lastErr error) (in []byte, err error) {
	// ignore for now
	return
}

// ECMRx implements the endpoint 1 OUT function, used to receive Ethernet
// packet from host to device.
func (eth *NIC) ECMRx(out []byte, lastErr error) (_ []byte, err error) {
	if len(eth.buf) == 0 && len(out) < 14 {
		return
	}

	eth.buf = append(eth.buf, out...)

	// more data expected or zero length packet
	if len(out) == eth.maxPacketSize {
		return
	}

	hdr := eth.buf[0:14]
	proto := tcpip.NetworkProtocolNumber(binary.BigEndian.Uint16(eth.buf[12:14]))
	payload := eth.buf[14:]

	pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
		ReserveHeaderBytes: len(hdr),
		Payload:            buffer.MakeWithData(payload),
	})

	copy(pkt.LinkHeader().Push(len(hdr)), hdr)

	eth.Link.InjectInbound(proto, pkt)
	eth.buf = []byte{}

	return
}

// ECMTx implements the endpoint 1 IN function, used to transmit Ethernet
// packet from device to host.
func (eth *NIC) ECMTx(_ []byte, lastErr error) (in []byte, err error) {
	var pkt *stack.PacketBuffer

	if pkt = eth.Link.Read(); pkt.IsNil() {
		return
	}

	proto := make([]byte, 2)
	binary.BigEndian.PutUint16(proto, uint16(pkt.NetworkProtocolNumber))

	// Ethernet frame header
	in = append(in, eth.HostMAC...)
	in = append(in, eth.DeviceMAC...)
	in = append(in, proto...)

	for _, v := range pkt.AsSlices() {
		in = append(in, v...)
	}

	return
}
