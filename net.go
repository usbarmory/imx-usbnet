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
var MTU = 1500

// Interface represents an Ethernet over USB interface instance.
type Interface struct {
	addr tcpip.Address

	nicid tcpip.NICID
	nic   *NIC

	stack *stack.Stack
	link  *channel.Endpoint

	device *usb.Device
}

func (n *Interface) configure(deviceMAC string) (err error) {
	n.stack = stack.New(stack.Options{
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

	n.link = channel.New(256, MTU, linkAddr)
	linkEP := stack.LinkEndpoint(n.link)

	if err := n.stack.CreateNIC(n.nicid, linkEP); err != nil {
		return fmt.Errorf("%v", err)
	}

	if err := n.stack.AddAddress(n.nicid, ipv4.ProtocolNumber, n.addr); err != nil {
		return fmt.Errorf("%v", err)
	}

	subnet, err := tcpip.NewSubnet("\x00\x00\x00\x00", "\x00\x00\x00\x00")

	if err != nil {
		return err
	}

	n.stack.SetRouteTable([]tcpip.Route{{
		Destination: subnet,
		nic:         n.nicid,
	}})

	return
}

// EnableICMP adds an ICMP endpoint to the interface, it is useful to enable
// ping requests.
func (n *Interface) EnableICMP() error {
	var wq waiter.Queue

	ep, err := n.stack.NewEndpoint(icmp.ProtocolNumber4, ipv4.ProtocolNumber, &wq)

	if err != nil {
		return fmt.Errorf("endpoint error (icmp): %v", err)
	}

	fullAddr := tcpip.FullAddress{Addr: n.addr, Port: 0, NIC: n.nicid}

	if err := ep.Bind(fullAddr); err != nil {
		return fmt.Errorf("bind error (icmp endpoint): ", err)
	}

	return nil
}

// Device returns the USB device associated to the Ethernet instance.
func (n *Network) Device() *usb.Device {
	return n.device
}

// ListenerTCP4 returns a net.Listener capable of accepting connections for the
// argument port on the Ethernet over USB device.
func (n *Network) ListenerTCP4(port uint16) (net.Listener, error) {
	fullAddr := tcpip.FullAddress{Addr: n.addr, Port: port, NIC: n.nicid}

	listener, err := gonet.ListenTCP(n.stack, fullAddr, ipv4.ProtocolNumber)

	if err != nil {
		return nil, err
	}

	return (net.Listener)(listener), nil
}

// Init initializes an Ethernet over USB device.
func Init(deviceIP string, deviceMAC, hostMAC string, id int) (n *Interface, err error) {
	hostAddress, err := net.ParseMAC(hostMAC)

	if err != nil {
		return
	}

	deviceAddress, err := net.ParseMAC(deviceMAC)

	if err != nil {
		return
	}

	n = &Network{
		nicid: tcpip.NICID(id),
		addr:  tcpip.Address(net.ParseIP(deviceIP)).To4(),
	}

	if err = n.configure(deviceMAC); err != nil {
		return
	}

	n.device = &usb.Device{}
	configureDevice(n.device)

	n.nic = &NIC{
		Host:   hostAddress,
		Device: deviceAddress,
		Link:   n.link,
	}

	err = n.nic.Init(n.device, 0)

	return
}
