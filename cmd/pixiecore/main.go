// Copyright 2016 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"go.universe.tf/netboot/pixiecore"
	"go.universe.tf/netboot/pixiecore/cli"
	"go.universe.tf/netboot/third_party/ipxe"
)

// qemu-system-x86_64 -L . --bios /usr/share/edk2-firmware/ipv6/OVMF.fd -netdev bridge,br=br1,id=net0 -device virtio-net-pci,netdev=net0
// sudo ./pixiecore bootipv6 --listen-addr=2001:db8:f00f:cafe::4/64 --httpboot-url=http://[2001:db8:f00f:cafe::4]/bootx64.efi --ipxe-url=http://[2001:db8:f00f:cafe::4]/script.ipxe
// sudo ./pixiecore ipv6api --listen-addr=2001:db8:f00f:cafe::4  --api-request-url=http://[2001:db8:f00f:cafe::4]:8888

func main() {
	cli.Ipxe[pixiecore.FirmwareX86PC] = ipxe.MustAsset("undionly.kpxe")
	cli.Ipxe[pixiecore.FirmwareEFI32] = ipxe.MustAsset("ipxe-i386.efi")
	cli.Ipxe[pixiecore.FirmwareEFI64] = ipxe.MustAsset("ipxe-x86_64.efi")
	cli.Ipxe[pixiecore.FirmwareEFIBC] = ipxe.MustAsset("ipxe-x86_64.efi")
	cli.Ipxe[pixiecore.FirmwareX86Ipxe] = ipxe.MustAsset("ipxe.pxe")
	cli.CLI()
}
