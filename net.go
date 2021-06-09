// Ethernet over USB driver
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package usbnet implements a driver for Ethernet over USB emulation on i.MX6
// SoCs.
//
// It currently implements CDC-ECM networking and for this reason the Ethernet
// device is only supported on Linux hosts. Applications are meant to use the
// driver in combination with gVisor tcpip package to expose TCP/IP networking
// stack through Ethernet over USB.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/f-secure-foundry/tamago.
package usbnet

import (
	"fmt"
	"net"

	"github.com/f-secure-foundry/tamago/soc/imx6/usb"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/network/arp"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
)

// MTU represents the Ethernet Maximum Transmission Unit
var MTU uint32 = 1500

// Interface represents an Ethernet over USB interface instance.
type Interface struct {
	addr tcpip.Address

	nicid tcpip.NICID
	nic   *NIC

	stack *stack.Stack
	link  *channel.Endpoint

	device *usb.Device
}

func (iface *Interface) configure(deviceMAC string) (err error) {
	iface.stack = stack.New(stack.Options{
		NetworkProtocols: []stack.NetworkProtocolFactory{
			ipv4.NewProtocol,
			arp.NewProtocol},
		TransportProtocols: []stack.TransportProtocolFactory{
			tcp.NewProtocol,
			icmp.NewProtocol4},
	})

	linkAddr, err := tcpip.ParseMACAddress(deviceMAC)

	if err != nil {
		return
	}

	iface.link = channel.New(256, MTU, linkAddr)
	linkEP := stack.LinkEndpoint(iface.link)

	if err := iface.stack.CreateNIC(iface.nicid, linkEP); err != nil {
		return fmt.Errorf("%v", err)
	}

	if err := iface.stack.AddAddress(iface.nicid, ipv4.ProtocolNumber, iface.addr); err != nil {
		return fmt.Errorf("%v", err)
	}

	subnet, err := tcpip.NewSubnet("\x00\x00\x00\x00", "\x00\x00\x00\x00")

	if err != nil {
		return err
	}

	iface.stack.SetRouteTable([]tcpip.Route{{
		Destination: subnet,
		NIC:         iface.nicid,
	}})

	return
}

// EnableICMP adds an ICMP endpoint to the interface, it is useful to enable
// ping requests.
func (iface *Interface) EnableICMP() error {
	var wq waiter.Queue

	ep, err := iface.stack.NewEndpoint(icmp.ProtocolNumber4, ipv4.ProtocolNumber, &wq)

	if err != nil {
		return fmt.Errorf("endpoint error (icmp): %v", err)
	}

	fullAddr := tcpip.FullAddress{Addr: iface.addr, Port: 0, NIC: iface.nicid}

	if err := ep.Bind(fullAddr); err != nil {
		return fmt.Errorf("bind error (icmp endpoint): ", err)
	}

	return nil
}

// Device returns the USB device associated to the Ethernet instance.
func (iface *Interface) Device() *usb.Device {
	return iface.device
}

// ListenerTCP4 returns a net.Listener capable of accepting connections for the
// argument port on the Ethernet over USB device.
func (iface *Interface) ListenerTCP4(port uint16) (net.Listener, error) {
	fullAddr := tcpip.FullAddress{Addr: iface.addr, Port: port, NIC: iface.nicid}

	listener, err := gonet.ListenTCP(iface.stack, fullAddr, ipv4.ProtocolNumber)

	if err != nil {
		return nil, err
	}

	return (net.Listener)(listener), nil
}

// Init initializes an Ethernet over USB device.
func Init(deviceIP string, deviceMAC, hostMAC string, id int) (iface *Interface, err error) {
	hostAddress, err := net.ParseMAC(hostMAC)

	if err != nil {
		return
	}

	deviceAddress, err := net.ParseMAC(deviceMAC)

	if err != nil {
		return
	}

	iface = &Interface{
		nicid: tcpip.NICID(id),
		addr:  tcpip.Address(net.ParseIP(deviceIP)).To4(),
	}

	if err = iface.configure(deviceMAC); err != nil {
		return
	}

	iface.device = &usb.Device{}
	configureDevice(iface.device)

	iface.nic = &NIC{
		Host:   hostAddress,
		Device: deviceAddress,
		Link:   iface.link,
	}

	err = iface.nic.Init(iface.device, 0)

	return
}
