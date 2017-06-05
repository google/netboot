package dhcp6

import (
	"encoding/binary"
	"fmt"
	"bytes"
	"net"
)

const (
	OptClientId       uint16 = 1  // IPMask
	OptServerId                = 2  // int32
	OptIaNa                    = 3  // IPs
	OptIaTa                    = 4  // IPs
	OptIaAddr                  = 5  // string
	OptOro                     = 6  // uint16
	OptPreference              = 7  // string
	OptElapsedTime             = 8  // IP
	OptRelayMessage            = 9  // IP
	OptAuth                    = 11 // []byte
	OptUnicast                 = 12 // IP
	OptStatusCode              = 13 // uint32
	OptRapidCommit             = 14 // byte
	OptUserClass               = 15 // IP
	OptVendorClass             = 16 // []byte
	OptVendorOpts              = 17 // string
	OptInterfaceId             = 18 // uint16
	OptReconfMsg               = 19 // uint32
	OptReconfAccept            = 20 // uint32
	OptRecursiveDns			   = 23 // []byte
	OptBootfileUrl             = 59
	OptBootfileParam           = 60 //[][]byte
	OptClientArchType          = 61 //[][]byte, sent by the client
	// 24? Domain search list
)

type Option struct {
	Id uint16
	Length uint16
	Value []byte
}

type Options map[uint16]*Option

func MakeOptions(bs []byte) (Options, error) {
	to_ret := make(Options)
	for len(bs) > 0 {
		optionLength := uint16(binary.BigEndian.Uint16(bs[2:4]))
		optionId := uint16(binary.BigEndian.Uint16(bs[0:2]))
		switch optionId {
			// parse client_id
			// parse server_id
			// parse IaNa # do I need to support IaTa?
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
		to_ret[optionId] = &Option{ Id: optionId, Length: optionLength, Value: bs[4 : 4+optionLength]}
		bs = bs[4+optionLength:]
	}
	return to_ret, nil
}

func (o Options) HumanReadable() []string {
	to_ret := make([]string, 0, len(o))
	for _, opt := range(o) {
		to_ret = append(to_ret, fmt.Sprintf("Option: %d | %d | %d | %s\n", opt.Id, opt.Length, opt.Value, opt.Value))
	}
	return to_ret
}

func (o Options) AddOption(option *Option) {
	o[option.Id] = option
}

func MakeIaNaOption(iaid []byte, t1, t2 uint32, iaAddr *Option) (*Option) {
	serializedIaAddr, _ := iaAddr.Marshal()
	value := make([]byte, 12 + len(serializedIaAddr))
	copy(value[0:], iaid[0:4])
	binary.BigEndian.PutUint32(value[4:], t1)
	binary.BigEndian.PutUint32(value[8:], t2)
	copy(value[12:], serializedIaAddr)
	return &Option{Id: OptIaNa, Length: uint16(len(value)), Value: value}
}

func MakeIaAddrOption(addr net.IP, preferredLifetime, validLifetime uint32) (*Option) {
	value := make([]byte, 24)
	copy(value[0:], addr)
	binary.BigEndian.PutUint32(value[16:], preferredLifetime)
	binary.BigEndian.PutUint32(value[20:], validLifetime)
	return &Option{ Id: OptIaAddr, Length: uint16(len(value)), Value: value}
}

func (o Options) Marshal() ([]byte, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, 1446))
	for _, v := range(o) {
		serialized, err := v.Marshal()
		if err != nil {
			return nil, fmt.Errorf("Error serializing option value: %s", err)
		}
		if err := binary.Write(buffer, binary.BigEndian, serialized); err != nil {
			return nil, fmt.Errorf("Error serializing option value: %s", err)
		}
	}
	return buffer.Bytes(), nil
}

func (o *Option) Marshal() ([]byte, error) {
	buffer := bytes.NewBuffer(make([]byte, 0, o.Length + 2))

	err := binary.Write(buffer, binary.BigEndian, o.Id)
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

func (o Options) UnmarshalOptionRequestOption() map[uint16]bool {
	oro_content := o[OptOro].Value
	to_ret := make(map[uint16]bool)

	for i := 0; i < int(o[OptOro].Length)/2; i++ {
		to_ret[uint16(binary.BigEndian.Uint16(oro_content[i*2:(i+1)*2]))] = true
	}
	return to_ret
}

func (o Options) RequestedBootFileUrlOption() bool {
	requested_options := o.UnmarshalOptionRequestOption()
	_, present := requested_options[OptBootfileUrl]
	return present
}

func (o Options) HasClientId() bool {
	_, present := o[OptClientId]
	return present
}

func (o Options) HasServerId() bool {
	_, present := o[OptServerId]
	return present
}

func (o Options) HasIaNa() bool {
	_, present := o[OptIaNa]
	return present
}

func (o Options) HasIaTa() bool {
	_, present := o[OptIaTa]
	return present
}