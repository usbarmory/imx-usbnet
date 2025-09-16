i.MX Ethernet over USB driver
=============================

This Go package implements TCP/IP connectivity through Ethernet over USB
(CDC-ECM) on NXP i.MX SoCs, to be used with `GOOS=tamago GOARCH=arm` as
supported by the [TamaGo](https://github.com/usbarmory/tamago) framework for
bare metal Go.

The CDC-ECM Ethernet over USB driver is supported natively on Linux and macOS
hosts while Windows requires using third-party drivers.

The package supports TCP/IP networking through gVisor (`go` branch)
[tcpip](https://pkg.go.dev/gvisor.dev/gvisor/pkg/tcpip)
stack pure Go implementation.

The interface TCP/IP stack can be attached to the Go runtime by setting
`net.SocketFunc` to the interface `Socket` function:

```
// i.MX Ethernet over USB interface
iface := usbnet.Interface{}

// initialize with IP, device MAC, host MAC
_ = usbnet.Init("10.0.0.1", "1a:55:89:a2:69:41", "1a:55:89:a2:69:42")

// Go runtime hook
net.SocketFunc = iface.Socket
```

See [tamago-example](https://github.com/usbarmory/tamago-example/blob/master/network/imx-usbnet.go)
for a full integration example.

Authors
=======

Andrea Barisani  
andrea@inversepath.com  

Andrej Rosano  
andrej@inversepath.com  

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
Copyright (c) The imx-usbnet authors. All Rights Reserved.

These source files are distributed under the BSD-style license found in the
[LICENSE](https://github.com/usbarmory/imx-usbnet/blob/master/LICENSE) file.
