package main

/*
 * target.go
 * Resolve and send targets
 * By J. Stuart McMurray
 * Created 20180209
 * Last Modified 20180209
 */

import (
	"errors"
	"net"
)

// DEFPORT is the default SSH port
const DEFPORT = "22"

// SendTargets parses t, which may be an address, hostname, or CIDR block, and
// sends all the relevant addresses to ch.
func SendTargets(ch chan<- string, t string) error {
	/* If it parses as a CIDR block, it's not a hostname */
	if _, ipnet, err := net.ParseCIDR(t); nil == err {
		sendCIDR(ch, ipnet)
		return nil
	}

	/* If there's a port, save it before resolution */
	h, p, _ := net.SplitHostPort(t)
	if "" == h {
		h = t
	}
	if "" == p {
		p = DEFPORT
	}

	/* Get all addresses for target */
	ips, err := net.LookupIP(h)
	if nil != err {
		return err
	}
	if 0 == len(ips) {
		return errors.New("target resolves to no IP addresses")
	}

	for _, ip := range ips {
		ch <- net.JoinHostPort(ip.String(), p)
	}

	return nil
}

/* sendCIDR sends all of the addresses in ipnet to ch, with DEFPORT added */
func sendCIDR(ch chan<- string, ipnet *net.IPNet) {
	/* Get initial IP address in its own slice */
	ip := net.IP(make([]byte, len(ipnet.IP)))
	copy(ip, ipnet.IP)

	for ; ipnet.Contains(ip); func(i net.IP) {
		for j := len(i) - 1; 0 <= j; j-- {
			i[j]++
			if 0 < i[j] {
				break
			}
		}
	}(ip) {
		ch <- net.JoinHostPort(ip.String(), DEFPORT)
	}
}
