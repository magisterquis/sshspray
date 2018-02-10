// sshspray runs a script on multiple SSH servers, in parallel
package main

/*
 * sshspray.go
 * Run a script on multiple SSH servers, in parallel
 * By J. Stuart McMurary
 * Created 20180209
 * Last Modified 20180209
 */

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

func main() {
	var (
		user = flag.String(
			"user",
			"root",
			"SSH `username`",
		)
		pass = flag.String(
			"pass",
			"",
			"SSH `password`, or - to read from stdin",
		)
		keyFile = flag.String(
			"key",
			"",
			"SSH private `key`",
		)
		scriptFile = flag.String(
			"script",
			"sshspray.script",
			"Name of script `file` to run on targets",
		)
		interpreter = flag.String(
			"interpreter",
			"/bin/sh",
			"Interpreter `command` to which to pass script",
		)
		nPar = flag.Uint(
			"parallel",
			200,
			"Attack `N` targets in parallel",
		)
	)
	flag.Usage = func() {
		fmt.Fprintf(
			os.Stderr,
			`Usage: %v [options] target [target...]

Makes SSH connections to the given targets, which may be hostnames, IP
addresses, or CIDR blocks, requests runs the give script on them using the
given interpreter.  By default, it's similar to

cat scriptfile | ssh user@target /bin/sh

Auth can either be via password or SSH private key.  If neither are given,
none authentication will still be attempted.  This is almost certainly not what
you want.

No host key validation checking is performed as this is meant to be used at the
start of exercises before the blue team has time to do anything.

Options:
`,
			os.Args[0],
		)
		flag.PrintDefaults()
	}
	flag.Parse()

	log.SetOutput(os.Stderr)

	/* Make sure we have targets */
	if 0 == flag.NArg() {
		fmt.Fprintf(os.Stderr, "No targets specified\n")
		os.Exit(6)
	}

	/* Slurp the script file */
	if "" == *scriptFile {
		fmt.Fprintf(os.Stderr, "Script file needed (-script)\n")
		os.Exit(1)
	}
	script, err := ioutil.ReadFile(*scriptFile)
	if nil != err {
		fmt.Fprintf(
			os.Stderr,
			"Unable to read script from %v: %v\n",
			*scriptFile,
			err,
		)
		os.Exit(2)
	}
	log.Printf("Read %v-byte script from %v", len(script), *scriptFile)

	/* Roll an SSH config */
	conf := ClientConfig(*user, *keyFile, *pass)

	/* Start attackers */
	var (
		wg = new(sync.WaitGroup)
		ch = make(chan string)
	)
	for i := uint(0); i < *nPar; i++ {
		wg.Add(1)
		go Attacker(ch, conf, *interpreter, script, wg)
	}

	/* Send hosts to attackers */
	for _, v := range flag.Args() {
		if err := SendTargets(ch, v); nil != err {
			log.Printf(
				"[%v] Unable to send for targeting: %v",
				v,
				err,
			)
		}
	}

	close(ch)
	wg.Wait()
	log.Printf("Done.")
}
