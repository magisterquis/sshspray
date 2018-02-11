sshspray
========

sshspray runs a script on multiple targets in parallel.  It relies on the
servers having the same creds.

For legal use only.

Authentication
--------------
A single username is used for every target.  Credentials can be in the form of
a password with `-pass` (which can be `-` to read a single line from stdin), an
SSH key with `-key`, or no creds at all, which only tries SSH's `none` auth
mechanism.  If the password is not set, a blank password will be attempted.

Script
------
The script is run by spawning an interpreter and passing the script on stdin.
For bash scripts, this means that every command must return or be backgrounded,
otherwise the connection will hang.

The command to use to spawn the script can be set with `-interpreter`, and the
script file with `-script`.  Under the hood, this opens an SSH channel, starts
a session, and requests the command specified with `-interpreter` be executed.
The contents of the script file are sent to the command's standard input.

An example script useful for backdooring a Linux host is included in this
repository as `sshspray.script`, which is the default script filename.

Targets
-------
Targets may be specified as IP addresses, CIDR ranges, or hostnames.  If a
hostname resolves to multiple IP addresses, all will be attacked.  By default,
a one second timeout is applied to connection attempts and SSH handshakes.  The
number of targets to be attacked in parallel is settable with `-parallel`.


Examples
--------
Try to run `bad.sh` against all hosts on a couple of small internal ranges with
the username `kitten` and password `meow`
```bash
sshspray -user kitten -password meow -script bad.sh 192.168.3.0/27 192.168.5.0/27
```

Use a stolen key and a stolen password against everything in a large network
```bash
sshspray -user itadmin -pass spring2018 -key id_rsa.itadmin -script bad.sh -parallel 1000 172.16.0.0/12
```

Run a script against a specific target list with long timeouts
```bash
sshspray -user root -pass password1 -targets ./vulnlist -timeout 1m
```

Usage
-----
```
Usage: sshspray [options] target [target...]

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
  -interpreter command
    	Interpreter command to which to pass script (default "/bin/sh")
  -key key
    	SSH private key
  -parallel N
    	Attack N targets in parallel (default 200)
  -pass password
    	SSH password, or - to read from stdin
  -script file
    	Name of script file to run on targets (default "sshspray.script")
  -targets file
    	Optional file with targets, one per line
  -timeout delay
    	Connection and Authentication timeout delay (default 4s)
  -user username
    	SSH username (default "root")
```

Windows
-------
Should work just fine.  Binaries available upon request.
