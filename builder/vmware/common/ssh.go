// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-sdk/sdk-internals/communicator/ssh"
	"golang.org/x/net/proxy"
)

func CommHost(config *SSHConfig) func(multistep.StateBag) (string, error) {
	return func(state multistep.StateBag) (string, error) {
		driver := state.Get("driver").(Driver)
		comm := config.Comm

		host := comm.Host()
		if host != "" {
			return host, nil
		}

		// Snag the port from the communicator config. This way we can use it
		// to perform a 3-way handshake with all of the hosts we suspect in
		// order to determine which one of the hosts is the correct one.
		port := comm.Port()

		// Get the list of potential addresses that the guest might use.
		hosts, err := driver.PotentialGuestIP(state)
		if err != nil {
			log.Printf("IP lookup failed: %s", err)
			return "", fmt.Errorf("IP lookup failed: %s", err)
		}

		if len(hosts) == 0 {
			log.Println("IP is blank, no IP yet.")
			return "", errors.New("IP is blank")
		}

		var pAddr string
		var pAuth *proxy.Auth
		if config.Comm.SSH.SSHProxyHost != "" {
			pAddr = fmt.Sprintf("%s:%d", config.Comm.SSH.SSHProxyHost, config.Comm.SSH.SSHProxyPort)
			if config.Comm.SSH.SSHProxyUsername != "" {
				pAuth = new(proxy.Auth)
				pAuth.User = config.Comm.SSH.SSHProxyUsername
				pAuth.Password = config.Comm.SSH.SSHProxyPassword
			}
		}

		// Iterate through our list of addresses and dial up each one similar to
		// a really inefficient port-scan. This way we can determine which of
		// the leases that we've parsed was the correct one and actually has our
		// target ssh/winrm service bound to a tcp port.
		var connFunc func() (net.Conn, error)
		for index, host := range hosts {
			if pAddr != "" {
				// Connect via SOCKS5 proxy
				connFunc = ssh.ProxyConnectFunc(pAddr, pAuth, "tcp", fmt.Sprintf("%s:%d", host, port))
			} else {
				// No bastion host, connect directly
				connFunc = ssh.ConnectFunc("tcp", fmt.Sprintf("%s:%d", host, port))
			}
			conn, err := connFunc()

			// If we got a connection, then we should be good to go. Return the
			// address to the caller and pray that things work out.
			if err == nil {
				conn.Close()

				log.Printf("Detected IP: %s", host)
				return host, nil

			}

			// Otherwise we need to iterate to the next entry and keep hoping.
			log.Printf("Skipping lease entry #%d due to being unable to connect to the host (%s) with tcp port (%d).", 1+index, host, port)
		}

		return "", errors.New("Host is not up")
	}
}
