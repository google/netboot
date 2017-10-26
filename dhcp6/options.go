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

func MakeOption(id uint16, value []byte) *Option {
	return &Option{ Id: id, Length: uint16(len(value)), Value: value}
}

type Options map[uint16][]*Option

func MakeOptions(bs []byte) (Options, error) {
	to_ret := make(Options)
	for len(bs) > 0 {
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
		_, present := to_ret[optionId]; if !present {
			to_ret[optionId] = make([]*Option, 1)
		}
		to_ret[optionId] = append(to_ret[optionId], &Option{ Id: optionId, Length: optionLength, Value: bs[4 : 4+optionLength]})
		bs = bs[4+optionLength:]
	}
	return to_ret, nil
}

func (o Options) HumanReadable() []string {
	to_ret := make([]string, 0, len(o))
	for _, multipleOptions := range(o) {
		for _, option := range(multipleOptions) {
			switch option.Id {
			case 3:
				to_ret = append(to_ret, o.HumanReadableIaNa(*option)...)
			default:
				to_ret = append(to_ret, fmt.Sprintf("Option: %d | %d | %d | %s\n", option.Id, option.Length, option.Value, option.Value))
			}
		}
	}
	return to_ret
}

func (o Options) HumanReadableIaNa(opt Option) []string {
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

func (o Options) AddOption(option *Option) {
	_, present := o[option.Id]; if !present {
		o[option.Id] = make([]*Option, 0)
	}
	o[option.Id] = append(o[option.Id], option)
}

func MakeIaNaOption(iaid []byte, t1, t2 uint32, iaAddr *Option) *Option {
	serializedIaAddr, _ := iaAddr.Marshal()
	value := make([]byte, 12 + len(serializedIaAddr))
	copy(value[0:], iaid[0:4])
	binary.BigEndian.PutUint32(value[4:], t1)
	binary.BigEndian.PutUint32(value[8:], t2)
	copy(value[12:], serializedIaAddr)
	return MakeOption(OptIaNa, value)
}

func MakeIaAddrOption(addr net.IP, preferredLifetime, validLifetime uint32) *Option {
	value := make([]byte, 24)
	copy(value[0:], addr)
	binary.BigEndian.PutUint32(value[16:], preferredLifetime)
	binary.BigEndian.PutUint32(value[20:], validLifetime)
	return MakeOption(OptIaAddr, value)
}

func MakeStatusOption(statusCode uint16, message string) *Option {
	value := make([]byte, 2 + len(message))
	binary.BigEndian.PutUint16(value[0:], statusCode)
	copy(value[2:], []byte(message))
	return MakeOption(OptStatusCode, value)
}

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
	to_ret := make(map[uint16]bool)

	_, present := o[OptOro]; if !present {
		return to_ret
	}

	oro_content := o[OptOro][0].Value
	for i := 0; i < int(o[OptOro][0].Length)/2; i++ {
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

func (o Options) HasClientArchType() bool {
	_, present := o[OptClientArchType]
	return present
}

func (o Options) ClientId() []byte {
	opt, exists := o[OptClientId]
	if exists {
		return opt[0].Value
	}
	return nil
}

func (o Options) ServerId() []byte {
	opt, exists := o[OptServerId]
	if exists {
		return opt[0].Value
	}
	return nil
}

func (o Options) IaNaIds() [][]byte {
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

func (o Options) ClientArchType() uint16 {
	opt, exists := o[OptClientArchType]
	if exists {
		return binary.BigEndian.Uint16(opt[0].Value)
	}
	return 0
}

func (o Options) BootfileUrl() []byte {
	opt, exists := o[OptBootfileUrl]
	if exists {
		return opt[0].Value
	}
	return nil
}

