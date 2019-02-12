package main

/*
 * config.go
 * Roll an SSH client config
 * By J. Stuart McMurray
 * Created 20180209
 * Last Modified 20190211
 */

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"golang.org/x/crypto/ssh"
)

// ClientConfig makes an ssh.ClientConfig from the given username and creds.
// It terminates the program on error.  If allowKI is true,
// keyboard-interactive auth will also be tried
func ClientConfig(user, keyFile, pass string, allowKI bool) *ssh.ClientConfig {
	var cc ssh.ClientConfig
	cc.Auth = make([]ssh.AuthMethod, 0)
	cc.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	/* Make sure we have a username */
	if "" == user {
		fmt.Fprintf(os.Stderr, "Username needed (-user)\n")
		os.Exit(3)
	}
	cc.User = user

	/* Get a key if we have one */
	var (
		key ssh.Signer
		err error
	)
	if key, err = readKey(keyFile); nil != err {
		fmt.Fprintf(
			os.Stderr,
			"Unable to read key from %v: %v\n",
			keyFile,
			err,
		)
		os.Exit(4)
	}
	if nil != key {
		cc.Auth = append(cc.Auth, ssh.PublicKeys(key))
	}

	/* Read the password if the user requested it */
	if "-" == pass {
		if pass, err = readPass(); nil != err {
			fmt.Fprintf(
				os.Stderr,
				"Error reading password from stdin: %v\n",
				err,
			)
			os.Exit(5)
		}
	}
	cc.Auth = append(cc.Auth, ssh.Password(pass))
	if allowKI {
		cc.Auth = append(cc.Auth, ssh.KeyboardInteractive(func(
			u string, i string, qs []string, e []bool,
		) ([]string, error) {
			as := make([]string, len(qs))
			for i := range as {
				as[i] = pass
			}
			return as, nil
		}))
	}

	return &cc
}

/* readKey reads the SSH key from f.  If f is empty, a nil ssh.Signer and error
are returned. */
func readKey(f string) (ssh.Signer, error) {
	/* If we have no file, we have no key, that's ok */
	if "" == f {
		return nil, nil
	}
	/* Slurp a key and parse it */
	b, err := ioutil.ReadFile(f)
	if nil != err {
		return nil, err
	}
	return ssh.ParsePrivateKey(b)
}

/* readPass reads a password as a single line from stdin */
func readPass() (string, error) {
	var (
		p   string
		b   = make([]byte, 1)
		n   int
		err error
	)
	/* Read from stdin until error or \n */
	for {
		/* Read a byte from stdin */
		n, err = os.Stdin.Read(b)
		/* If we actually got one, add it to the pasword, maybe */
		if 1 == n {
			switch b[0] {
			case '\n': /* Done */
				return p, nil
			case '\r': /* Silly windows newline */
			default:
				p += string(b)
			}
		}
		/* EOF's are benign and mean we're done with the password.  It happens
		if someone does `echo -n foo | ./program -pass -` */
		if io.EOF == err {
			return p, nil
		}
		if nil != err {
			return p, err
		}
	}
}
