package dhcp6

import (
	"encoding/binary"
	"fmt"
	"bytes"
	"net"
)

// DHCPv6 option IDs
const (
	// Client ID Option
	OptClientID     uint16 = 1
	// Server ID Option
	OptServerID            = 2
	// Identity Association for Non-temporary Addresses Option
	OptIaNa                = 3
	// Identity Association for Temporary Addresses Option
	OptIaTa                = 4
	// IA Address Option
	OptIaAddr              = 5
	// Option Request Option
	OptOro            = 6
	// Preference Option
	OptPreference     = 7
	// Elapsed Time Option
	OptElapsedTime    = 8
	// Relay Message Option
	OptRelayMessage   = 9
	// Authentication Option
	OptAuth           = 11
	// Server Unicast Option
	OptUnicast        = 12
	// Status Code Option
	OptStatusCode     = 13
	// Rapid Commit Option
	OptRapidCommit    = 14
	// User Class Option
	OptUserClass      = 15
	// Vendor Class Option
	OptVendorClass    = 16
	// Vendor-specific Information Option
	OptVendorOpts     = 17
	// Interface-Id Option
	OptInterfaceID    = 18
	// Reconfigure Message Option
	OptReconfMsg      = 19
	// Reconfigure Accept Option
	OptReconfAccept   = 20
	// Recursive DNS name servers Option
	OptRecursiveDNS   = 23
	// Boot File URL Option
	OptBootfileURL    = 59
	// Boot File Parameters Option
	OptBootfileParam  = 60
	// Client Architecture Type Option
	OptClientArchType = 61
)

// Option represents a DHCPv6 Option
type Option struct {
	ID     uint16
	Length uint16
	Value  []byte
}

// MakeOption creates an Option with given ID and value
func MakeOption(id uint16, value []byte) *Option {
	return &Option{ ID: id, Length: uint16(len(value)), Value: value}
}

// Options contains all options of a DHCPv6 packet
type Options map[uint16][]*Option

func MakeOptions(bs []byte) (Options, error) {
	ret := make(Options)
	for len(bs) > 0 {
		o, err := UnmarshalOption(bs)
		if err != nil {
			return nil, err
		}
		ret[o.ID] = append(ret[o.ID], &Option{ ID: o.ID, Length: o.Length, Value: bs[4 : 4+o.Length]})
		bs = bs[4+o.Length:]
	}
	return ret, nil
}

// UnmarshalOption de-serializes an Option
func UnmarshalOption(bs []byte) (*Option, error) {
	optionLength := uint16(binary.BigEndian.Uint16(bs[2:4]))
	optionId := uint16(binary.BigEndian.Uint16(bs[0:2]))
	switch optionId {
	// parse client_id
	// parse server_id
	//parse ipaddr
	case OptOro:
		if optionLength% 2 != 0 {
			return nil, fmt.Errorf("OptionID request for options (6) length should be even number of bytes: %d", optionLength)
		}
	default:
		if len(bs[4:]) < int(optionLength) {
			fmt.Printf("option %d claims to have %d bytes of payload, but only has %d bytes", optionId, optionLength, len(bs[4:]))
			return nil, fmt.Errorf("option %d claims to have %d bytes of payload, but only has %d bytes", optionId, optionLength, len(bs[4:]))
		}
	}
	return &Option{ ID: optionId, Length: optionLength, Value: bs[4 : 4+optionLength]}, nil
}

// HumanReadable presents DHCPv6 options in a human-readable form
func (o Options) HumanReadable() []string {
	to_ret := make([]string, 0, len(o))
	for _, multipleOptions := range(o) {
		for _, option := range(multipleOptions) {
			switch option.ID {
			case 3:
				to_ret = append(to_ret, o.humanReadableIaNa(*option)...)
			default:
				to_ret = append(to_ret, fmt.Sprintf("Option: %d | %d | %d | %s\n", option.ID, option.Length, option.Value, option.Value))
			}
		}
	}
	return to_ret
}

func (o Options) humanReadableIaNa(opt Option) []string {
	to_ret := make([]string, 0)
	to_ret = append(to_ret, fmt.Sprintf("Option: OptIaNa | len %d | iaid %x | t1 %d | t2 %d\n",
		opt.Length, opt.Value[0:4], binary.BigEndian.Uint32(opt.Value[4:8]), binary.BigEndian.Uint32(opt.Value[8:12])))

	if opt.Length <= 12 {
		return to_ret // no options
	}

	iaOptions := opt.Value[12:]
	for len(iaOptions) > 0 {
		l := uint16(binary.BigEndian.Uint16(iaOptions[2:4]))
		id := uint16(binary.BigEndian.Uint16(iaOptions[0:2]))


		switch id {
		case OptIaAddr:
			ip := make(net.IP, 16)
			copy(ip, iaOptions[4:20])
			to_ret = append(to_ret, fmt.Sprintf("\tOption: IA_ADDR | len %d | ip %s | preferred %d | valid %d | %v \n",
				l, ip, binary.BigEndian.Uint32(iaOptions[20:24]), binary.BigEndian.Uint32(iaOptions[24:28]), iaOptions[28:4+l]))
		default:
			to_ret = append(to_ret, fmt.Sprintf("\tOption: id %d | len %d | %s\n",
				id, l, iaOptions[4:4+l]))
		}

		iaOptions = iaOptions[4+l:]
	}

	return to_ret
}

// AddOption adds an option to Options
func (o Options) AddOption(option *Option) {
	_, present := o[option.ID]; if !present {
		o[option.ID] = make([]*Option, 0)
	}
	o[option.ID] = append(o[option.ID], option)
}

// MakeIaNaOption creates a Identity Association for Non-temporary Addresses Option
// with specified interface ID, t1 and t2 times, and an interface-specific option
// (an IA Address Option or a Status Option)
func MakeIaNaOption(iaid []byte, t1, t2 uint32, iaOption *Option) *Option {
	serializedIaOption, _ := iaOption.Marshal()
	value := make([]byte, 12 + len(serializedIaOption))
	copy(value[0:], iaid[0:4])
	binary.BigEndian.PutUint32(value[4:], t1)
	binary.BigEndian.PutUint32(value[8:], t2)
	copy(value[12:], serializedIaOption)
	return MakeOption(OptIaNa, value)
}

// MakeIaAddrOption creates an IA Address Option using IP address,
// preferred and valid lifetimes
func MakeIaAddrOption(addr net.IP, preferredLifetime, validLifetime uint32) *Option {
	value := make([]byte, 24)
	copy(value[0:], addr)
	binary.BigEndian.PutUint32(value[16:], preferredLifetime)
	binary.BigEndian.PutUint32(value[20:], validLifetime)
	return MakeOption(OptIaAddr, value)
}

// MakeStatusOption creates a Status Option with given status code and message
func MakeStatusOption(statusCode uint16, message string) *Option {
	value := make([]byte, 2 + len(message))
	binary.BigEndian.PutUint16(value[0:], statusCode)
	copy(value[2:], []byte(message))
	return MakeOption(OptStatusCode, value)
}

// MakeDNSServersOption creates a Recursive DNS servers Option with the specified list of IP addresses
func MakeDNSServersOption(addresses []net.IP) *Option {
	value := make([]byte, 16*len(addresses))
	for i, dnsAddress := range addresses {
		copy(value[i*16:], dnsAddress)
	}
	return MakeOption(OptRecursiveDNS, value)
}

// Marshal serializes Options
func (o Options) Marshal() ([]byte, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, 1446))
	for _, multipleOptions := range(o) {
		for _, o := range (multipleOptions) {
			serialized, err := o.Marshal()
			if err != nil {
				return nil, fmt.Errorf("Error serializing option value: %s", err)
			}
			if err := binary.Write(buffer, binary.BigEndian, serialized); err != nil {
				return nil, fmt.Errorf("Error serializing option value: %s", err)
			}
		}
	}
	return buffer.Bytes(), nil
}

// Marshal serializes the Option
func (o *Option) Marshal() ([]byte, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, o.Length + 2))

	err := binary.Write(buffer, binary.BigEndian, o.ID)
	if err != nil {
		return nil, fmt.Errorf("Error serializing option id: %s", err)
	}
	err = binary.Write(buffer, binary.BigEndian, o.Length)
	if err != nil {
		return nil, fmt.Errorf("Error serializing option length: %s", err)
	}
	err = binary.Write(buffer, binary.BigEndian, o.Value)
	if err != nil {
		return nil, fmt.Errorf("Error serializing option value: %s", err)
	}
	return buffer.Bytes(), nil
}

// UnmarshalOptionRequestOption de-serializes Option Request Option
func (o Options) UnmarshalOptionRequestOption() map[uint16]bool {
	ret := make(map[uint16]bool)

	_, present := o[OptOro]; if !present {
		return ret
	}

	value := o[OptOro][0].Value
	for i := 0; i < int(o[OptOro][0].Length)/2; i++ {
		ret[uint16(binary.BigEndian.Uint16(value[i*2:(i+1)*2]))] = true
	}
	return ret
}

// HasBootFileURLOption returns true if Options contains Boot File URL Option
func (o Options) HasBootFileURLOption() bool {
	requestedOptions := o.UnmarshalOptionRequestOption()
	_, present := requestedOptions[OptBootfileURL]
	return present
}

// HasClientID returns true if Options contains Client ID Option
func (o Options) HasClientID() bool {
	_, present := o[OptClientID]
	return present
}

// HasServerID returns true if Options contains Server ID Option
func (o Options) HasServerID() bool {
	_, present := o[OptServerID]
	return present
}

// HasIaNa returns true oif Options contains Identity Association for Non-Temporary Addresses Option
func (o Options) HasIaNa() bool {
	_, present := o[OptIaNa]
	return present
}

// HasIaTa returns true if Options contains Identity Association for Temporary Addresses Option
func (o Options) HasIaTa() bool {
	_, present := o[OptIaTa]
	return present
}

// HasClientArchType returns true if Options contains Client Architecture Type Option
func (o Options) HasClientArchType() bool {
	_, present := o[OptClientArchType]
	return present
}

// ClientID returns the value in the Client ID Option or nil if the option doesn't exist
func (o Options) ClientID() []byte {
	opt, exists := o[OptClientID]
	if exists {
		return opt[0].Value
	}
	return nil
}

// ServerID returns the value in the Server ID Option or nil if the option doesn't exist
func (o Options) ServerID() []byte {
	opt, exists := o[OptServerID]
	if exists {
		return opt[0].Value
	}
	return nil
}

// IaNaIDs returns a list of interface IDs in all Identity Association for Non-Temporary Addresses Options,
// or an empty list if none exist
func (o Options) IaNaIDs() [][]byte {
	options, exists := o[OptIaNa]
	ret := make([][]byte, 0)
	if exists {
		for _, option := range(options) {
			 ret = append(ret, option.Value[0:4])
		}
		return ret
	}
	return ret
}

// ClientArchType returns the value in the Client Architecture Type Option, or 0 if the option doesn't exist
func (o Options) ClientArchType() uint16 {
	opt, exists := o[OptClientArchType]
	if exists {
		return binary.BigEndian.Uint16(opt[0].Value)
	}
	return 0
}

// BootFileURL returns the value in the Boot File URL Option, or nil if the option doesn't exist
func (o Options) BootFileURL() []byte {
	opt, exists := o[OptBootfileURL]
	if exists {
		return opt[0].Value
	}
	return nil
}

