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
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"
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
		timeout = flag.Duration(
			"timeout",
			4*time.Second,
			"Connection and Authentication timeout `delay`",
		)
		targetFile = flag.String(
			"targets",
			"",
			"Optional `file` with targets, one per line",
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
	if 0 == flag.NArg() && "" == *targetFile {
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
		wg  = new(sync.WaitGroup)
		ach = make(chan string) /* Attack targets */
	)
	for i := uint(0); i < *nPar; i++ {
		wg.Add(1)
		go Attacker(ach, conf, *interpreter, script, *timeout, wg)
	}

	/* Queue up targets */
	tch := make(chan string)
	go func() {
		for t := range tch {
			if err := SendTargets(ach, t); nil != err {
				log.Printf(
					"[%v] Unable to target: %v",
					t,
					err,
				)
			}
		}
		close(ach)
	}()

	/* Send targets from command line */
	for _, t := range flag.Args() {
		tch <- t
	}

	/* Send targets from the file, if we have one */
	if "" != *targetFile {
		if err := sendFromFile(tch, *targetFile); nil != err {
			log.Printf(
				"Error reading targets from %q: %v",
				*targetFile,
				err,
			)
		}
	}

	close(tch)
	wg.Wait()
	log.Printf("Done.")
}

/* sendFromFile sends non-blank, non-comment (#) lines from fn to ch */
func sendFromFile(ch chan<- string, fn string) error {
	/* Open file for reading */
	f, err := os.Open(fn)
	if nil != err {
		return err
	}
	defer f.Close()

	/* Read lines, send to ch */
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		/* Get a line, skip blanks and comments */
		l := strings.TrimSpace(scanner.Text())
		if "" == l || strings.HasPrefix(l, "#") {
			continue
		}
		ch <- l
	}
	return scanner.Err()
}
