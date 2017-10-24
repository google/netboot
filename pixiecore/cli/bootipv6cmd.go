package cli

import (
	"github.com/spf13/cobra"
	"fmt"
	"go.universe.tf/netboot/pixiecorev6"
	"go.universe.tf/netboot/dhcp6"
)

var bootIPv6Cmd = &cobra.Command{
	Use:   "bootipv6",
	Short: "Boot a kernel and optional init ramdisks over IPv6",
	Run: func(cmd *cobra.Command, args []string) {
		addr, err := cmd.Flags().GetString("listen-addr")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}
		ipxeUrl, err := cmd.Flags().GetString("ipxe-url")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}
		httpBootUrl, err := cmd.Flags().GetString("httpboot-url")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}

		s := pixiecorev6.NewServerV6()

		s.Log = logWithStdFmt
		debug, err := cmd.Flags().GetBool("debug")
		if err != nil {
			s.Debug = logWithStdFmt
		}
		if debug { s.Debug = logWithStdFmt }

		if addr == "" {
			fatalf("Please specify address to bind to")
		} else {
		}
		if ipxeUrl == "" {
			fatalf("Please specify ipxe config file url")
		}
		if httpBootUrl == "" {
			fatalf("Please specify httpboot url")
		}

		s.Address = addr
		s.BootUrls = dhcp6.MakeStaticBootConfiguration(httpBootUrl, ipxeUrl)

		fmt.Println(s.Serve())
	},
}

func serverv6ConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("listen-addr", "", "", "IPv6 address to listen on")
	cmd.Flags().StringP("ipxe-url", "", "", "IPXE config file url, e.g. http://[2001:db8:f00f:cafe::4]/script.ipxe")
	cmd.Flags().StringP("httpboot-url", "", "", "HTTPBoot url, e.g. http://[2001:db8:f00f:cafe::4]/bootx64.efi")
	cmd.Flags().Bool("debug", false, "Enable debug-level logging")
}

func init() {
	rootCmd.AddCommand(bootIPv6Cmd)
	serverv6ConfigFlags(bootIPv6Cmd)
}
