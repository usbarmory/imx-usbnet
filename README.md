i.MX Ethernet over USB driver
=============================

This Go package implements a driver for Ethernet over USB emulation on NXP i.MX
SoCs, it meant to be used with `GOOS=tamago GOARCH=arm` as supported by the
[TamaGo](https://github.com/f-secure-foundry/tamago) framework for bare metal
Go on ARM SoCs.

It currently implements CDC-ECM networking and for this reason the Ethernet
device is only supported on Linux hosts.

The package leverages on [gVisor tcpip](https://pkg.go.dev/gvisor.dev/gvisor/pkg/tcpip) to implement a TCP/IP
networking stack on bare metal, exposed through Ethernet over USB.

Authors
=======

Andrea Barisani  
andrea.barisani@f-secure.com | andrea@inversepath.com  

Andrej Rosano  
andrej.rosano@f-secure.com   | andrej@inversepath.com  

Documentation
=============

The package API documentation can be found on
[pkg.go.dev](https://pkg.go.dev/github.com/f-secure-foundry/imx-usbnet).


For more information about TamaGo see its
[repository](https://github.com/f-secure-foundry/tamago) and
[project wiki](https://github.com/f-secure-foundry/tamago/wiki).

License
=======

tamago | https://github.com/f-secure-foundry/imx-usbnet  
Copyright (c) F-Secure Corporation

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/f-secure-foundry/imx-usbnet/blob/master/LICENSE) file.
