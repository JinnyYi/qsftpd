package client

import (
	"fmt"
	"math/rand"
	"net"
	"strings"

	"github.com/yunify/qsftp/context"
	"github.com/yunify/qsftp/transfer"
	"github.com/yunify/qsftp/utils"
)

func (c *Handler) handlePASV() {
	addr, _ := net.ResolveTCPAddr("tcp", ":0")
	var tcpListener *net.TCPListener
	var err error

	portRange := context.Settings.DataPortRange
	listenHost := context.Settings.ListenHost
	publicHost := context.Settings.PublicHost

	if portRange != nil {
		for start := portRange.Start; start < portRange.End; start++ {
			port := portRange.Start + rand.Intn(portRange.End-portRange.Start)
			localAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", listenHost, port))
			if err != nil {
				continue
			}

			tcpListener, err = net.ListenTCP("tcp", localAddr)
			if err == nil {
				break
			}
		}

	} else {
		tcpListener, err = net.ListenTCP("tcp", addr)
	}

	if err != nil {
		context.Logger.ErrorF("Could not listen: %v", err)
		return
	}

	// The listener will either be plain TCP or TLS.
	var listener net.Listener
	if c.transferTLS {
		c.WriteMessage(550, fmt.Sprintf("Cannot get a TLS config: %v", err))
		//listener = tls.NewListener(tcpListener, tlsConfig)
	} else {
		listener = tcpListener
	}

	p := &transfer.PassiveHandler{
		TCPListener: tcpListener,
		Listener:    listener,
		Port:        tcpListener.Addr().(*net.TCPAddr).Port,
	}

	// We should rewrite this part.
	if c.command == "PASV" {
		p1 := p.Port / 256
		p2 := p.Port - (p1 * 256)

		quads := strings.Split(publicHost, ".")
		c.WriteMessage(227, fmt.Sprintf("Entering Passive Mode (%s,%s,%s,%s,%d,%d)", quads[0], quads[1], quads[2], quads[3], p1, p2))
	} else {
		c.WriteMessage(229, fmt.Sprintf("Entering Extended Passive Mode (|||%d|)", p.Port))
	}

	c.transfer = p
}

func (c *Handler) handlePORT() {
	c.transfer = &transfer.ActiveHandler{
		RemoteAddr: utils.ParseRemoteAddr(c.param),
	}
	c.WriteMessage(200, "PORT command successful")
}
