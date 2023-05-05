// Copyright 2021 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

//go:build !linux

package network

import (
	"net"

	"github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"
)

type sourceOtherSuite struct {
	testing.IsolationSuite
}

var _ = gc.Suite(&sourceOtherSuite{})

func (s *sourceOtherSuite) TestNewNetAddr(c *gc.C) {
	addr, err := newNetAddr("192.168.20.1/24")
	c.Assert(err, jc.ErrorIsNil)

	c.Check(addr.String(), gc.Equals, "192.168.20.1/24")
	c.Assert(addr.IP(), gc.NotNil)
	c.Check(addr.IP().String(), gc.Equals, "192.168.20.1")
	c.Assert(addr.IPNet(), gc.NotNil)
	c.Check(addr.IPNet().String(), gc.Equals, "192.168.20.0/24")

	addr, err = newNetAddr("192.168.20.1")
	c.Assert(err, jc.ErrorIsNil)

	c.Check(addr.String(), gc.Equals, "192.168.20.1")
	c.Assert(addr.IP(), gc.NotNil)
	c.Check(addr.IP().String(), gc.Equals, "192.168.20.1")
	c.Assert(addr.IPNet(), gc.IsNil)

	addr, err = newNetAddr("fe80::5054:ff:fedd:eef0/64")
	c.Assert(err, jc.ErrorIsNil)

	c.Check(addr.String(), gc.Equals, "fe80::5054:ff:fedd:eef0/64")
	c.Assert(addr.IP(), gc.NotNil)
	c.Check(addr.IP().String(), gc.Equals, "fe80::5054:ff:fedd:eef0")
	c.Assert(addr.IPNet(), gc.NotNil)
	c.Check(addr.IPNet().String(), gc.Equals, "fe80::/64")

	addr, err = newNetAddr("fe80::5054:ff:fedd:eef0")
	c.Assert(err, jc.ErrorIsNil)

	c.Check(addr.String(), gc.Equals, "fe80::5054:ff:fedd:eef0")
	c.Assert(addr.IP(), gc.NotNil)
	c.Check(addr.IP().String(), gc.Equals, "fe80::5054:ff:fedd:eef0")
	c.Assert(addr.IPNet(), gc.IsNil)

	addr, err = newNetAddr("y u no parse")
	c.Assert(err, gc.ErrorMatches, `unable to parse IP address "y u no parse"`)
}

func (s *sourceOtherSuite) TestConfigSourceInterfaces(c *gc.C) {
	rawNICs := []net.Interface{{
		Index: 1,
		MTU:   65536,
		Name:  "lo",
		Flags: net.FlagUp | net.FlagLoopback,
	}, {
		Index:        2,
		MTU:          1500,
		Name:         "eth0",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f0"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        10,
		MTU:          1500,
		Name:         "br-eth0",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f0"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        11,
		MTU:          1500,
		Name:         "br-eth1",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f1"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        3,
		MTU:          1500,
		Name:         "eth1",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f1"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        12,
		MTU:          1500,
		Name:         "br-eth0.100",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f0"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        13,
		MTU:          1500,
		Name:         "eth0.100",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f0"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        14,
		MTU:          1500,
		Name:         "br-eth0.250",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f0"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        15,
		MTU:          1500,
		Name:         "eth0.250",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f0"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        16,
		MTU:          1500,
		Name:         "br-eth0.50",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f0"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        17,
		MTU:          1500,
		Name:         "eth0.50",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f0"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        18,
		MTU:          1500,
		Name:         "br-eth1.11",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f1"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        19,
		MTU:          1500,
		Name:         "eth1.11",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f1"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        20,
		MTU:          1500,
		Name:         "br-eth1.12",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f1"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        21,
		MTU:          1500,
		Name:         "eth1.12",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f1"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        22,
		MTU:          1500,
		Name:         "br-eth1.13",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f1"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}, {
		Index:        23,
		MTU:          1500,
		Name:         "eth1.13",
		HardwareAddr: parseMAC(c, "aa:bb:cc:dd:ee:f1"),
		Flags:        net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}}

	source := netPackageConfigSource{
		interfaces: func() ([]net.Interface, error) { return rawNICs, nil },
	}

	sourceNICs, err := source.Interfaces()
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(sourceNICs, gc.HasLen, len(rawNICs))

	for i, nic := range sourceNICs {
		raw := rawNICs[i]

		c.Check(nic.Name(), gc.Equals, raw.Name)
		c.Check(nic.Index(), gc.Equals, raw.Index)
		c.Check(nic.MTU(), gc.Equals, raw.MTU)
		c.Check(nic.HardwareAddr(), gc.DeepEquals, raw.HardwareAddr)
		c.Check(nic.IsUp(), jc.IsTrue)
	}
}

func (s *sourceOtherSuite) TestNICTypeDerivation(c *gc.C) {
	getType := func(string) LinkLayerDeviceType { return BondDevice }

	// If we have get value, return it.
	raw := &net.Interface{}
	c.Check(newNetNIC(raw, getType).Type(), gc.Equals, BondDevice)

	getType = func(string) LinkLayerDeviceType { return UnknownDevice }

	// Infer loopback from flags.
	raw = &net.Interface{
		Flags: net.FlagUp | net.FlagLoopback,
	}
	c.Check(newNetNIC(raw, getType).Type(), gc.Equals, LoopbackDevice)

	// Default to ethernet otherwise.
	raw = &net.Interface{
		Flags: net.FlagUp | net.FlagBroadcast | net.FlagMulticast,
	}
	c.Check(newNetNIC(raw, getType).Type(), gc.Equals, EthernetDevice)
}

func parseMAC(c *gc.C, val string) net.HardwareAddr {
	mac, err := net.ParseMAC(val)
	c.Assert(err, jc.ErrorIsNil)
	return mac
}
