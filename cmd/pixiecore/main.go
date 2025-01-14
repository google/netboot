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
	"github.com/levigross/netboot/artifacts"
	"go.universe.tf/netboot/pixiecore"
	"go.universe.tf/netboot/pixiecore/cli"
)

func mustRead(filename string) []byte {
	f, err := artifacts.Content.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	return f
}

func main() {

	cli.Ipxe[pixiecore.FirmwareX86PC] = mustRead("ipxe/bin/undionly.kpxe")
	cli.Ipxe[pixiecore.FirmwareEFI32] = mustRead("ipxe/bin-i386-efi/ipxe.efi")
	cli.Ipxe[pixiecore.FirmwareEFI64] = mustRead("ipxe/bin-x86_64-efi/ipxe.efi")
	cli.Ipxe[pixiecore.FirmwareEFIBC] = mustRead("ipxe/bin-x86_64-efi/ipxe.efi")
	cli.Ipxe[pixiecore.FirmwareX86Ipxe] = mustRead("ipxe/bin/ipxe.pxe")
	cli.CLI()
}
