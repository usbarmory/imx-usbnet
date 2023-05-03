// Ethernet over USB driver
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

package usbnet

import (
	"strings"

	"github.com/usbarmory/tamago/soc/nxp/usb"
)

// MaxPacketSize represents the USB data interface endpoint maximum packet size
var MaxPacketSize uint16 = 512

func addControlInterface(device *usb.Device, eth *NIC) (iface *usb.InterfaceDescriptor) {
	iface = &usb.InterfaceDescriptor{}
	iface.SetDefaults()

	iface.NumEndpoints = 1
	iface.InterfaceClass = usb.COMMUNICATION_INTERFACE_CLASS
	iface.InterfaceSubClass = usb.ETH_SUBCLASS

	iInterface, _ := device.AddString(`CDC Ethernet Control Model (ECM)`)
	iface.Interface = iInterface

	// Set IAD to be inserted before first interface, to support multiple
	// functions in this same configuration.
	iface.IAD = &usb.InterfaceAssociationDescriptor{}
	iface.IAD.SetDefaults()
	iface.IAD.InterfaceCount = 2
	iface.IAD.FunctionClass = iface.InterfaceClass
	iface.IAD.FunctionSubClass = iface.InterfaceSubClass

	iFunction, _ := device.AddString(`CDC`)
	iface.IAD.Function = iFunction

	header := &usb.CDCHeaderDescriptor{}
	header.SetDefaults()

	iface.ClassDescriptors = append(iface.ClassDescriptors, header.Bytes())

	union := &usb.CDCUnionDescriptor{}
	union.SetDefaults()

	numInterfaces := 1 + len(device.Configurations[0].Interfaces)
	union.MasterInterface = uint8(numInterfaces - 1)
	union.SlaveInterface0 = uint8(numInterfaces)

	iface.ClassDescriptors = append(iface.ClassDescriptors, union.Bytes())

	ethernet := &usb.CDCEthernetDescriptor{}
	ethernet.SetDefaults()

	iMacAddress, _ := device.AddString(strings.ReplaceAll(eth.HostMAC.String(), ":", ""))
	ethernet.MacAddress = iMacAddress

	iface.ClassDescriptors = append(iface.ClassDescriptors, ethernet.Bytes())

	ep2IN := &usb.EndpointDescriptor{}
	ep2IN.SetDefaults()
	ep2IN.EndpointAddress = 0x82
	ep2IN.Attributes = 3
	ep2IN.MaxPacketSize = 16
	ep2IN.Interval = 9
	ep2IN.Function = eth.Control

	iface.Endpoints = append(iface.Endpoints, ep2IN)

	device.Configurations[0].AddInterface(iface)

	return
}

func addDataInterfaces(device *usb.Device, eth *NIC) {
	iface0 := &usb.InterfaceDescriptor{}
	iface0.SetDefaults()

	iface0.NumEndpoints = 0
	iface0.InterfaceClass = usb.DATA_INTERFACE_CLASS

	device.Configurations[0].AddInterface(iface0)

	// CDC requires the use of a default interface setting with no
	// endpoints to signal a deactivated state, an additional interface
	// setting (with the data exchange endpoint pair) is used for normal
	// operation (see the Topology section in USB CDC specifications).

	iface1 := &usb.InterfaceDescriptor{}
	iface1.SetDefaults()

	iface1.AlternateSetting = 1
	iface1.NumEndpoints = 2
	iface0.InterfaceClass = usb.DATA_INTERFACE_CLASS

	iInterface, _ := device.AddString(`CDC Data`)
	iface1.Interface = iInterface

	ep1IN := &usb.EndpointDescriptor{}
	ep1IN.SetDefaults()
	ep1IN.EndpointAddress = 0x81
	ep1IN.Attributes = 2
	ep1IN.MaxPacketSize = MaxPacketSize
	ep1IN.Function = eth.Tx

	iface1.Endpoints = append(iface1.Endpoints, ep1IN)

	ep1OUT := &usb.EndpointDescriptor{}
	ep1OUT.SetDefaults()
	ep1OUT.EndpointAddress = 0x01
	ep1OUT.MaxPacketSize = MaxPacketSize
	ep1OUT.Attributes = 2
	ep1OUT.Function = eth.Rx

	iface1.Endpoints = append(iface1.Endpoints, ep1OUT)

	device.Configurations[0].AddInterface(iface1)

	eth.maxPacketSize = int(MaxPacketSize)

	return
}

// ConfigureDevice configures a USB device with default descriptors for a CDC
// Ethernet (ECM) device, suitable for Add().
func ConfigureDevice(device *usb.Device, serial string) {
	// Supported Language Code Zero: English
	device.SetLanguageCodes([]uint16{0x0409})

	// device descriptor
	device.Descriptor = &usb.DeviceDescriptor{}
	device.Descriptor.SetDefaults()

	// p5, Table 1-1. Device Descriptor Using Class Codes for IAD,
	// USB Interface Association Descriptor Device Class Code and Use Model.
	device.Descriptor.DeviceClass = 0xef
	device.Descriptor.DeviceSubClass = 0x02
	device.Descriptor.DeviceProtocol = 0x01

	// http://pid.codes/1209/2702/
	device.Descriptor.VendorId = 0x1209
	device.Descriptor.ProductId = 0x2702

	device.Descriptor.Device = 0x0001

	iManufacturer, _ := device.AddString(`WithSecure Foundry`)
	device.Descriptor.Manufacturer = iManufacturer

	iProduct, _ := device.AddString(`CDC Ethernet (ECM)`)
	device.Descriptor.Product = iProduct

	iSerial, _ := device.AddString(serial)
	device.Descriptor.SerialNumber = iSerial

	conf := &usb.ConfigurationDescriptor{}
	conf.SetDefaults()

	device.AddConfiguration(conf)

	// device qualifier
	device.Qualifier = &usb.DeviceQualifierDescriptor{}
	device.Qualifier.SetDefaults()
	device.Qualifier.NumConfigurations = uint8(len(device.Configurations))
}
