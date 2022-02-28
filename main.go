package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/hirochachacha/go-smb2"
)

var (
	user       string
	pass       string
	pipePrefix string
	host       string
	message    string
)

func init() {
	flag.StringVar(&host, "host", "", "remote host")
	flag.StringVar(&user, "user", "", "remote user")
	flag.StringVar(&pass, "pass", "", "remote user password")
	flag.StringVar(&pipePrefix, "pipe", "", "pipe on which write")
	flag.StringVar(&message, "message", "", "message to write on pipe")
	flag.Parse()

	if host == "" || user == "" || pass == "" || pipePrefix == "" || message == "" {
		flag.PrintDefaults()
		os.Exit(0)
	}
}

func main() {

	conn, err := net.Dial("tcp", host)
	check(err, "failed to dial")
	defer conn.Close()

	d := &smb2.Dialer{
		Initiator: &smb2.NTLMInitiator{
			User:     user,
			Password: pass,
		},
	}

	s, err := d.Dial(conn)
	check(err, "smb logon failed")
	defer s.Logoff()

	fs, err := s.Mount("IPC$")
	check(err, "couldn't mount IPC$")
	defer fs.Umount()

	files, err := fs.ReadDir(".")
	check(err, "unable to list directory")

	var pipe os.FileInfo
	for _, f := range files {
		if strings.HasPrefix(f.Name(), pipePrefix) {
			pipe = f
		}
	}

	if pipe == nil {
		fmt.Println("no such pipe found")
		os.Exit(1)
	}

	f, err := fs.OpenFile(pipe.Name(), os.O_WRONLY, 0)
	check(err, "failed opening pipe")
	defer f.Close()

	_, err = f.WriteString(message)
	check(err, "failed writing to pipe")

	fmt.Printf("wrote to: %s\n", f.Name())
}

func check(err error, msg string) {
	if err != nil {
		fmt.Printf("[!] %s\n", msg)
		panic(err)
	}
}
