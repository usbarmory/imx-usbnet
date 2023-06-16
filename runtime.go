// Ethernet over USB driver
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usbnet

import (
	"context"
	"fmt"
	"net"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
)

func (iface *Interface) hookGoNet() {
	net.DialFunc = func(ctx context.Context, network string, la net.Addr, ra net.Addr) (net.Conn, error) {
		switch network {
		case "tcp", "tcp4":
			return iface.DialContextTCP4(ctx, ra.String())
		case "udp", "udp4":
			return iface.DialUDP4("", ra.String())
		default:
			return nil, fmt.Errorf("unsupported network %s", network)
		}
	}

	net.ListenFunc = func(ctx context.Context, network string, la net.Addr) (net.Listener, error) {
		switch network {
		case "tcp", "tcp4":
		default:
			return nil, fmt.Errorf("unsupported network %s", network)
		}

		addr := la.(*net.TCPAddr)

		return iface.ListenerTCP4(uint16(addr.Port))
	}

	net.ListenPacketFunc = func(ctx context.Context, network string, la net.Addr) (net.PacketConn, error) {
		switch network {
		case "udp", "udp4":
		default:
			return nil, fmt.Errorf("unsupported network %s", network)
		}

		addr := la.(*net.UDPAddr)

		lAddr := tcpip.FullAddress{Addr: tcpip.Address(addr.IP), Port: uint16(addr.Port)}
		return gonet.DialUDP(iface.Stack, &lAddr, &tcpip.FullAddress{}, ipv4.ProtocolNumber)
	}
}
