package dhcp6

import (
	"net/http"
	"time"
	"strings"
	"fmt"
	"net/url"
	"bytes"
	"net"
)

type BootConfiguration interface {
	GetBootUrl(id []byte, clientArchType uint16) ([]byte, error)
	GetPreference() []byte
	GetRecursiveDns() []net.IP
}

type StaticBootConfiguration struct {
	HttpBootUrl 		[]byte
	IPxeBootUrl			[]byte
	RecursiveDns		[]net.IP
	Preference			[]byte
	UsePreference		bool
}

func MakeStaticBootConfiguration(httpBootUrl, ipxeBootUrl string, preference uint8, usePreference bool,
		dnsServerAddresses []net.IP) *StaticBootConfiguration {
	ret := &StaticBootConfiguration{HttpBootUrl: []byte(httpBootUrl), IPxeBootUrl: []byte(ipxeBootUrl), UsePreference: usePreference}
	if usePreference {
		ret.Preference = make([]byte, 1)
		ret.Preference[0] = byte(preference)
	}
	ret.RecursiveDns = dnsServerAddresses
	return ret
}

func (bc *StaticBootConfiguration) GetBootUrl(id []byte, clientArchType uint16) ([]byte, error) {
	if 0x10 ==  clientArchType {
		return bc.HttpBootUrl, nil
	}
	return bc.IPxeBootUrl, nil
}

func (bc *StaticBootConfiguration) GetPreference() []byte {
	return bc.Preference
}

func (bc *StaticBootConfiguration) GetRecursiveDns() []net.IP {
	return bc.RecursiveDns
}

type ApiBootConfiguration struct {
	client    			*http.Client
	urlPrefix 			string
	RecursiveDns		[]net.IP
	Preference			[]byte
	UsePreference		bool
}

func MakeApiBootConfiguration(url string, timeout time.Duration, preference uint8, usePreference bool,
		dnsServerAddresses []net.IP) *ApiBootConfiguration {
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	ret := &ApiBootConfiguration{
		client:    &http.Client{Timeout: timeout},
		urlPrefix: url + "v1",
		UsePreference: usePreference,
	}
	if usePreference {
		ret.Preference = make([]byte, 1)
		ret.Preference[0] = byte(preference)
	}
	ret.RecursiveDns = dnsServerAddresses

	return ret
}

func (bc *ApiBootConfiguration) GetBootUrl(id []byte, clientArchType uint16) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/boot/%x/%d", bc.urlPrefix, id, clientArchType)
	resp, err := bc.client.Get(reqURL)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("%s: %s", reqURL, http.StatusText(resp.StatusCode))
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	url, _ := bc.makeURLAbsolute(buf.String())

	return []byte(url), nil
}

func (bc *ApiBootConfiguration) makeURLAbsolute(urlStr string) (string, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("%q is not an URL", urlStr)
	}
	if !u.IsAbs() {
		base, err := url.Parse(bc.urlPrefix)
		if err != nil {
			return "", err
		}
		u = base.ResolveReference(u)
	}
	return u.String(), nil
}

func (bc *ApiBootConfiguration) GetPreference() []byte {
	return bc.Preference
}

func (bc *ApiBootConfiguration) GetRecursiveDns() []net.IP {
	return bc.RecursiveDns
}
