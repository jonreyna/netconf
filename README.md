[![Documentation](https://godoc.org/github.com/JReyLBC/netconf?status.svg)](http://godoc.org/github.com/JReyLBC/netconf)

I have a lot of documentation to do here, but at least there's a somewhat decent GoDoc, and this example shoud help too.

```go
package main

import (
	"encoding/xml"
	"io"
	"log"
	"net"
	"strings"

	"github.com/JReyLBC/netconf"
	"golang.org/x/crypto/ssh"
)

func main() {
	log.SetFlags(log.Flags() | log.Lshortfile)

	sshConfig := ssh.ClientConfig{
		User:            "happy_gopher",
		Auth:            []ssh.AuthMethod{ssh.Password("ConcurrencyIsNoParallelism!")},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := netconf.NewClient(&sshConfig, net.JoinHostPort("1.2.3.4", "830"))
	if err != nil {
		log.Panic(err)
	}

	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Panic(err)
	}

	defer session.Close()

	var serverHello netconf.HelloMessage

	decoder := session.NewDecoder()
	if err := decoder.DecodeHello(&serverHello); err != nil {
		log.Panic(err)
	}

	if _, err := io.Copy(session, strings.NewReader(netconf.DefaultHelloMessage)); err != nil {
		log.Panic(err)
	}

	type LLDPMethod struct {
		XMLName xml.Name `xml:"get-lldp-neighbors-information"`
	}

	if err := session.NewEncoder().Encode(&LLDPMethod{}); err != nil {
		log.Panic(err)
	}

	type Neighbor struct {
		LocalInterface         string `xml:"lldp-local-interface,omitempty"`
		LocalParentInterface   string `xml:"lldp-local-parent-interface-name,omitempty"`
		LocalPortID            string `xml:"lldp-local-port-id,omitempty"`
		RemoteChassisIDSubtype string `xml:"lldp-remote-chassis-id-subtype,omitempty"`
		RemoteChassisID        string `xml:"lldp-remote-chassis-id,omitempty"`
		RemotePortIDSubtype    string `xml:"lldp-remote-port-id-subtype,omitempty"`
		RemotePortID           string `xml:"lldp-remote-port-id,omitempty"`
		RemotePortDesc         string `xml:"lldp-remote-port-description,omitempty"`
		RemoteSystemName       string `xml:"lldp-remote-system-name,omitempty"`
	}

	type Results struct {
		Neighbors []Neighbor `xml:"lldp-neighbor-information,omitempty"`
	}

	lldpNbrResults := Results{}
	if err := decoder.Decode(&lldpNbrResults); err != nil {
		log.Panic(err)
	}

	log.Printf("%v", lldpNbrResults)
}
```
