// Ethernet over USB driver
//
// Copyright (c) F-Secure Corporation
// https://foundry.f-secure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package usbnet implements TCP/IP connectivity through Ethernet over USB
// (CDC-ECM) on i.MX6 SoCs.
//
// The CDC-ECM Ethernet over USB driver is supported natively on Linux and
// macOS hosts, while Windows requires third-party drivers.
//
// The TCP/IP stack is implemented using gVisor pure Go implementation.
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package usbnet

import (
	"fmt"
	"net"
	"strconv"

	"github.com/usbarmory/tamago/soc/imx6/usb"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/header"
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

	Stack *stack.Stack
	Link  *channel.Endpoint

	device *usb.Device
}

func (iface *Interface) configure(deviceMAC string) (err error) {
	iface.Stack = stack.New(stack.Options{
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

	iface.Link = channel.New(256, MTU, linkAddr)
	linkEP := stack.LinkEndpoint(iface.Link)

	if err := iface.Stack.CreateNIC(iface.nicid, linkEP); err != nil {
		return fmt.Errorf("%v", err)
	}

	if err := iface.Stack.AddAddress(iface.nicid, ipv4.ProtocolNumber, iface.addr); err != nil {
		return fmt.Errorf("%v", err)
	}

	rt := iface.Stack.GetRouteTable()

	rt = append(rt, tcpip.Route{
		Destination: header.IPv4EmptySubnet,
		NIC:         iface.nicid,
	})

	iface.Stack.SetRouteTable(rt)

	return
}

// EnableICMP adds an ICMP endpoint to the interface, it is useful to enable
// ping requests.
func (iface *Interface) EnableICMP() error {
	var wq waiter.Queue

	ep, err := iface.Stack.NewEndpoint(icmp.ProtocolNumber4, ipv4.ProtocolNumber, &wq)

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

// ListenerTCP4 returns a net.Listener capable of accepting IPv4 TCP
// connections for the argument port on the Ethernet over USB device.
func (iface *Interface) ListenerTCP4(port uint16) (net.Listener, error) {
	fullAddr := tcpip.FullAddress{Addr: iface.addr, Port: port, NIC: iface.nicid}

	listener, err := gonet.ListenTCP(iface.Stack, fullAddr, ipv4.ProtocolNumber)

	if err != nil {
		return nil, err
	}

	return (net.Listener)(listener), nil
}

// Dial connects to an IPv4 TCP address, over the Ethernet over USB interface.
func (iface *Interface) DialTCP4(address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)

	if err != nil {
		return nil, err
	}

	p, err := strconv.Atoi(port)

	if err != nil {
		return nil, err
	}

	addr := net.ParseIP(host)
	fullAddr := tcpip.FullAddress{Addr: tcpip.Address(addr.To4()), Port: uint16(p)}

	conn, err := gonet.DialTCP(iface.Stack, fullAddr, ipv4.ProtocolNumber)

	if err != nil {
		return nil, err
	}

	return (net.Conn)(conn), nil
}

// Add adds an Ethernet over USB configuration to a previously configured USB
// device, it can be used in place of Init() to create composite USB devices.
func Add(device *usb.Device, deviceIP string, deviceMAC, hostMAC string, id int) (iface *Interface, err error) {
	hostAddress, err := net.ParseMAC(hostMAC)

	if err != nil {
		return
	}

	deviceAddress, err := net.ParseMAC(deviceMAC)

	if err != nil {
		return
	}

	iface = &Interface{
		nicid:  tcpip.NICID(id),
		addr:   tcpip.Address(net.ParseIP(deviceIP)).To4(),
		device: device,
	}

	if err = iface.configure(deviceMAC); err != nil {
		return
	}

	iface.nic = &NIC{
		Host:   hostAddress,
		Device: deviceAddress,
		Link:   iface.Link,
	}

	err = iface.nic.Init(iface.device, 0)

	return
}

// Init initializes an Ethernet over USB device, configured with the defaults
// as set by ConfigureDevice().
func Init(deviceIP string, deviceMAC, hostMAC string, id int) (iface *Interface, err error) {
	device := &usb.Device{}
	ConfigureDevice(device)

	return Add(device, deviceIP, deviceMAC, hostMAC, id)
}
