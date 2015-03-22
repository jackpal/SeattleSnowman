// Copyright (C) 2015 John Howard Palevich. All Rights Reserved.

package router

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"

	"golang.org/x/crypto/ssh"
)

// A slice of net.IPs that defines set operations.
type IPs []net.IP

func (a IPs) Contains(ip net.IP) bool {
	for _, aa := range a {
		if aa.Equal(ip) {
			return true
		}
	}
	return false
}

func (a IPs) Append(ip net.IP) IPs {
	if a.Contains(ip) {
		return a
	}
	return append(a, ip)
}

func (a IPs) Remove(ip net.IP) (result IPs) {
	for _, aa := range a {
		if !aa.Equal(ip) {
			result = append(result, aa)
		}
	}
	return
}

func (a IPs) AddAll(b IPs) (result IPs) {
	table := make(map[string]bool)
	for _, ip := range a {
		key := string(ip)
		if !table[key] {
			table[key] = true
			result = append(result, ip)
		}
	}
	for _, ip := range b {
		key := string(ip)
		if !table[key] {
			table[key] = true
			result = append(result, ip)
		}
	}
	return
}

func (a IPs) RemoveAll(b IPs) (result IPs) {
	table := make(map[string]bool)
	for _, ip := range b {
		key := string(ip)
		table[key] = true
	}
	for _, ip := range a {
		key := string(ip)
		if !table[key] {
			result = append(result, ip)
		}
	}
	return
}

// A firewall can define address groups.
type Firewall interface {
	// Get current state of block group
	GetAddressGroup(groupName string) (ips IPs, err error)
	// Set new state of block group
	SetAddressGroup(groupName string, ips IPs) error
}

type edgeRouterFirewall struct {
	address string
	privateKeyPath string
	client *ssh.Client
}

func NewEdgeRouterFirewall(address string, privateKeyPath string) Firewall {
	return &edgeRouterFirewall{address, privateKeyPath, nil}
}

func (f *edgeRouterFirewall) GetAddressGroup(groupName string) (ips IPs, err error) {
	showCommand := fmt.Sprintf("show firewall group address-group %q\n", groupName)
	fullCommand := fmt.Sprintf("source /opt/vyatta/etc/functions/script-template\n\nconfigure\n%sexit\nexit\n", showCommand)
	result, err := f.routerRPC(fullCommand)
	if err != nil {
		return
	}
	addressGroup, err := parseAddressGroup(result)
	if err != nil {
		return
	}
	ips = addressGroup.address
	return
}

func (f *edgeRouterFirewall) SetAddressGroup(groupName string, ips IPs) (err error) {
	// Ignore error.
	currentIPs, _ := f.GetAddressGroup(groupName)
	addIPs, deleteIPs := computeDifference(currentIPs, ips)
	return f.updateAddressGroup(groupName, addIPs, deleteIPs)
}

func computeDifference(oldIPs IPs, newIPs IPs) (addIPs IPs, removeIPs IPs) {
	addIPs = newIPs.RemoveAll(oldIPs)
	removeIPs = oldIPs.RemoveAll(newIPs)
	return
}

// Return all elements of a that are not in b.
func subtractSet(a IPs, b IPs) (remainder IPs) {
	return a.RemoveAll(b)
}

func (f *edgeRouterFirewall) updateAddressGroup(groupName string, setIPs IPs, deleteIPs IPs) (err error) {
	log.Printf("updateAddressGroup(%q, %v, %v)",
		groupName, setIPs, deleteIPs)
	if len(setIPs) == 0 && len(deleteIPs) == 0 {
		// Nothing to do.
		log.Printf("nothing to do.")
		return
	}
	wrapper := "/opt/vyatta/sbin/vyatta-cfg-cmd-wrapper"
	cmd := wrapper + " begin\n"
	prefix := fmt.Sprintf("firewall group address-group %q address", groupName)
	for _, ip := range setIPs {
		cmd = cmd + fmt.Sprintf("%s set %s %s\n", wrapper, prefix, ip)
	}
	for _, ip := range deleteIPs {
		cmd = cmd + fmt.Sprintf("%s delete %s %s\n", wrapper, prefix, ip)
	}
	cmd = cmd + fmt.Sprintf("%s commit\n", wrapper)
	cmd = cmd + fmt.Sprintf("%s end\n", wrapper)
	_, err = f.routerRPC(cmd)
	if err != nil {
		return
	}
	return
}

func parsekey(file string) (private ssh.Signer, err error) {
	privateBytes, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}

	private, err = ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		return
	}
	return
}

func (f *edgeRouterFirewall) ensureClient() (err error) {
	if f.client != nil {
		return
	}
	pkey, err := parsekey(f.privateKeyPath)
	if err != nil {
		log.Printf("Failed to parse key %s", err)
		return
	}
	config := &ssh.ClientConfig{
		User: "ubnt",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(pkey),
		},
	}
	f.client, err = ssh.Dial("tcp", f.address, config)
	return
}

func (f *edgeRouterFirewall) ensureSession() (session *ssh.Session, err error) {
    for tries := 0; tries < 3; tries++ {
		err = f.ensureClient()
		if err != nil {
			log.Printf("Could not create ssh client: %v", f.address, err.Error())
			return
		}

		// Each ClientConn can support multiple interactive sessions,
		// represented by a Session.
		session, err = f.client.NewSession()
		if err != nil {
			log.Printf("Failed to create session: " + err.Error())
			// client might have disconnected. Try again.
			f.client.Close()
			f.client = nil
		}
	}
	return
}

func (f *edgeRouterFirewall) routerRPC(commands string) (result string, err error) {
	session, err := f.ensureSession()
	if err != nil {
		log.Printf("Failed to create session: " + err.Error())
		// client might
		return
	}

	defer session.Close()

	// Once a Session is created, you can execute a single command on
	// the remote side using the Run method.

	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer
	session.Stdout = &stdoutBuffer
	session.Stderr = &stderrBuffer
	session.Stdin = strings.NewReader(commands)
	if err = session.Start("/bin/vbash"); err != nil {
		log.Printf("Failed to start: " + err.Error())
		return
	}
	if err = session.Wait(); err != nil {
		log.Printf("Failed to finish running: " + err.Error())
		log.Printf("stdout: %q", stdoutBuffer.String())
		log.Printf("stderr: %q", stderrBuffer.String())
		return
	}
	result = stdoutBuffer.String()
	log.Printf("router: sent %q received %q", commands, result)
	return
}

/*
  An example show result

 address 192.168.1.208
 address 192.168.1.212
 address 192.168.1.213
 address 192.168.1.214
 address 192.168.1.215
 address 192.168.1.216
 address 192.168.1.217
 address 192.168.1.202
 address 192.168.1.210
 address 192.168.1.201
 description "Kids devices"

*/

type addressGroup struct {
	description string
	address     []net.IP
}

func parseAddressGroup(src string) (a addressGroup, err error) {
	lines := strings.Split(src, "\n")
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if len(line) == 0 {
			continue
		}
		kv := strings.SplitN(line, " ", 2)
		if len(kv) != 2 {
			err = fmt.Errorf("Parse error %q", line)
			return
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		if key == "address" {
			a.address = append(a.address, net.ParseIP(value))
		} else if key == "description" {
			a.description = strings.Trim(value, "\"") // TODO - handle embedded double-quotes
		} else {
			err = fmt.Errorf("Uknown key %q", key)
			return
		}
	}
	return
}
