package cli

import (
	"github.com/spf13/cobra"
	"fmt"
	"go.universe.tf/netboot/dhcp6"
	"go.universe.tf/netboot/dhcp6/pool"
	"time"
	"net"
	"strings"
	"go.universe.tf/netboot/pixiecore"
)

// pixiecore ipv6api --listen-addr=2001:db8:f00f:cafe::4  --api-request-url=http://[2001:db8:f00f:cafe::4]:8888

var ipv6ApiCmd = &cobra.Command{
	Use:   "ipv6api",
	Short: "Boot a kernel and optional init ramdisks over IPv6 using api",
	Run: func(cmd *cobra.Command, args []string) {
		addr, err := cmd.Flags().GetString("listen-addr")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}
		apiURL, err := cmd.Flags().GetString("api-request-url")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}
		apiTimeout, err := cmd.Flags().GetDuration("api-request-timeout")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}

		s := pixiecore.NewServerV6()
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
		if apiURL == "" {
			fatalf("Please specify ipxe config file url")
		}
		s.Address = addr
		preference, err := cmd.Flags().GetUint8("preference")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}
		dnsServers, err := cmd.Flags().GetString("dns-servers")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}
		dnsServerAddresses := make([]net.IP, 0)
		if cmd.Flags().Changed("dns-servers") {
			for _, dnsServerAddress := range strings.Split(dnsServers, ",") {
				dnsServerAddresses = append(dnsServerAddresses, net.ParseIP(dnsServerAddress))
			}
		}
		s.BootConfig = pixiecore.MakeAPIBootConfiguration(apiURL, apiTimeout, preference,
			cmd.Flags().Changed("preference"), dnsServerAddresses)

		addressPoolStart, err := cmd.Flags().GetString("address-pool-start")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}
		addressPoolSize, err := cmd.Flags().GetUint64("address-pool-size")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}
		addressPoolValidLifetime, err := cmd.Flags().GetUint32("address-pool-lifetime")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}
		s.AddressPool = pool.NewRandomAddressPool(net.ParseIP(addressPoolStart), addressPoolSize, addressPoolValidLifetime)
		s.PacketBuilder = dhcp6.MakePacketBuilder(addressPoolValidLifetime - addressPoolValidLifetime*3/100, addressPoolValidLifetime)

		fmt.Println(s.Serve())
	},
}

func serverv6APIConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("listen-addr", "", "", "IPv6 address to listen on")
	cmd.Flags().StringP("api-request-url", "", "", "Ipv6-specific API server url")
	cmd.Flags().Duration("api-request-timeout", 5*time.Second, "Timeout for request to the API server")
	cmd.Flags().Bool("debug", false, "Enable debug-level logging")
	cmd.Flags().Uint8("preference", 255, "Set dhcp server preference value")
	cmd.Flags().StringP("address-pool-start", "", "2001:db8:f00f:cafe:ffff::100", "Starting ip of the address pool, e.g. 2001:db8:f00f:cafe:ffff::100")
	cmd.Flags().Uint64("address-pool-size", 50, "Address pool size")
	cmd.Flags().Uint32("address-pool-lifetime", 1850, "Address pool ip address valid lifetime in seconds")
	cmd.Flags().StringP("dns-servers", "", "", "Comma separated list of one or more dns server addresses")
}

func init() {
	rootCmd.AddCommand(ipv6ApiCmd)
	serverv6APIConfigFlags(ipv6ApiCmd)
}

