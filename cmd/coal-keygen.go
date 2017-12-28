package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"golang.org/x/crypto/nacl/box"
	"io/ioutil"
	"os"
)

func main() {
	outFiles := flag.String("o", "", "Required output file for public and private keys")
	flag.Parse()
	if *outFiles == "" {
		fmt.Fprintf(os.Stderr, os.Args[0]+": -o is a REQUIRED option\n")
		os.Exit(1)
	}
	publicKey, privateKey, err := box.GenerateKey(rand.Reader)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(2)
	}
	err = ioutil.WriteFile(*outFiles, ([]byte)(base64.StdEncoding.EncodeToString(privateKey[:])), 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(3)
	}
	err = ioutil.WriteFile(*outFiles+".pub", ([]byte)(base64.StdEncoding.EncodeToString(publicKey[:])), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(4)
	}
	return
}
