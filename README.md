i.MX Ethernet over USB driver
=============================

This Go package implements TCP/IP connectivity through Ethernet over USB
(CDC-ECM) on NXP i.MX SoCs, to be used with `GOOS=tamago GOARCH=arm` as supported by the
[TamaGo](https://github.com/usbarmory/tamago) framework for bare metal
Go on ARM SoCs.

The CDC-ECM Ethernet over USB driver is supported natively on Linux and macOS
hosts while Windows requires using third-party drivers.

The package supports TCP/IP networking through gVisor [tcpip](https://pkg.go.dev/gvisor.dev/gvisor/pkg/tcpip)
package (`go` branch) stack pure Go implementation.

Authors
=======

Andrea Barisani  
andrea.barisani@f-secure.com | andrea@inversepath.com  

Andrej Rosano  
andrej.rosano@f-secure.com   | andrej@inversepath.com  

Documentation
=============

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/usbarmory/imx-usbnet).


For more information about TamaGo see its
[repository](https://github.com/usbarmory/tamago) and
[project wiki](https://github.com/usbarmory/tamago/wiki).

License
=======

tamago | https://github.com/usbarmory/imx-usbnet  
Copyright (c) F-Secure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/imx-usbnet/blob/master/LICENSE) file.
