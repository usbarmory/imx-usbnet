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
	"errors"
	"net"
	"syscall"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
)

// Socket can be used as net.SocketFunc under GOOS=tamago to allow its use
// internal use within the Go runtime.
func (iface *Interface) Socket(ctx context.Context, network string, family, sotype int, laddr, raddr net.Addr) (c interface{}, err error) {
	var proto tcpip.NetworkProtocolNumber
	var lFullAddr tcpip.FullAddress
	var rFullAddr tcpip.FullAddress

	if laddr != nil {
		if lFullAddr, err = fullAddr(laddr.String()); err != nil {
			return
		}
	}

	if raddr != nil {
		if rFullAddr, err = fullAddr(raddr.String()); err != nil {
			return
		}
	}

	switch family {
	case syscall.AF_INET:
		proto = ipv4.ProtocolNumber
	default:
		return nil, errors.New("unsupported address family")
	}

	switch network {
	case "udp", "udp4":
		if sotype != syscall.SOCK_DGRAM {
			return nil, errors.New("unsupported socket type")
		}

		if c, err = gonet.DialUDP(iface.Stack, &lFullAddr, &rFullAddr, proto); c != nil {
			return
		}
	case "tcp", "tcp4":
		if sotype != syscall.SOCK_STREAM {
			return nil, errors.New("unsupported socket type")
		}

		if raddr != nil {
			if c, err = gonet.DialContextTCP(ctx, iface.Stack, rFullAddr, proto); err != nil {
				return
			}
		} else {
			if c, err = gonet.ListenTCP(iface.Stack, lFullAddr, proto); err != nil {
				return
			}
		}
	default:
		return nil, errors.New("unsupported network")
	}

	return
}
