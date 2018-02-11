package main

/*
 * attack.go
 * Connect to an SSH server, run a script
 * By J. Stuart McMurray
 * Created 20180209
 * Last Modified 20180209
 */

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
)

// Attacker connects to targets and runs the script by passing it to the
// interpreter's stdin.  Targets are read from ch.
func Attacker(
	ch <-chan string,
	conf *ssh.ClientConfig,
	interpreter string,
	script []byte,
	timeout time.Duration,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for t := range ch {
		o, err := attack(
			t,
			conf,
			interpreter,
			bytes.NewReader(script),
			timeout,
		)
		/* Output with no error is a good thing */
		if nil == err {
			m := fmt.Sprintf("[%v] SUCCESS", t)
			if 0 != len(o) {
				m += fmt.Sprintf(": %q", string(o))
			}
			log.Printf("%v", m)
			continue
		}

		/* An error with nil output is a setup error, not a script
		error */
		if nil == o {
			log.Printf("[%v] ERROR: %v", t, err)
			continue
		}

		/* All other errors are execution errors */
		m := fmt.Sprintf("[%v] FAIL (%v)", t, err)
		if 0 != len(o) {
			m += fmt.Sprintf(": %q", string(o))
		}
		log.Printf("%v", m)
	}
}

/* attack tries to connect to t and tries to run script in the interpreter.  It
returns the output and any error encountered. */
func attack(
	t string,
	conf *ssh.ClientConfig,
	interpreter string,
	script io.Reader,
	timeout time.Duration,
) ([]byte, error) {
	/* Connect to target */
	c, err := net.DialTimeout("tcp", t, timeout)
	if nil != err {
		return nil, err
	}
	defer c.Close()

	/* Upgrade to SSH */
	var (
		sc    ssh.Conn
		chans <-chan ssh.NewChannel
		reqs  <-chan *ssh.Request
		done  = make(chan struct{})
	)
	go func() {
		defer close(done)
		sc, chans, reqs, err = ssh.NewClientConn(c, t, conf)
	}()

	/* Wait for timeout or handshake */
	select {
	case <-done: /* Handshake happened */
	case <-time.After(timeout): /* Timeout */
		return nil, errors.New("handshake timeout")
	}

	/* We have handshook by now, we hope */
	if nil != err {
		return nil, err
	}
	defer sc.Close()
	cc := ssh.NewClient(sc, chans, reqs)

	/* Start a session in which to run the script */
	s, err := cc.NewSession()
	if nil != err {
		return nil, err
	}
	s.Stdin = script
	defer s.Close()

	/* Run it and capture output */
	return s.CombinedOutput(interpreter)
}
