package cli

import (
	"github.com/spf13/cobra"
	"fmt"
	"go.universe.tf/netboot/pixiecorev6"
	"go.universe.tf/netboot/dhcp6"
	"time"
)

var ipv6ApiCmd = &cobra.Command{
	Use:   "ipv6api",
	Short: "Boot a kernel and optional init ramdisks over IPv6 using api",
	Run: func(cmd *cobra.Command, args []string) {
		addr, err := cmd.Flags().GetString("listen-addr")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}
		apiUrl, err := cmd.Flags().GetString("api-request-url")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}
		apiTimeout, err := cmd.Flags().GetDuration("api-request-timeout")
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
		if apiUrl == "" {
			fatalf("Please specify ipxe config file url")
		}

		s.Address = addr
		preference, err := cmd.Flags().GetUint8("preference")
		if err != nil {
			fatalf("Error reading flag: %s", err)
		}

		s.BootConfig = dhcp6.MakeApiBootConfiguration(apiUrl, apiTimeout, preference, cmd.Flags().Changed("preference"))

		fmt.Println(s.Serve())
	},
}

func serverv6ApiConfigFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("listen-addr", "", "", "IPv6 address to listen on")
	cmd.Flags().StringP("api-request-url", "", "", "Ipv6-specific API server url")
	cmd.Flags().Duration("api-request-timeout", 5*time.Second, "Timeout for request to the API server")
	cmd.Flags().Bool("debug", false, "Enable debug-level logging")
	cmd.Flags().Uint8("preference", 255, "Set dhcp server preference value")
}

func init() {
	rootCmd.AddCommand(ipv6ApiCmd)
	serverv6ApiConfigFlags(ipv6ApiCmd)
}

