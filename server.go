package shell

import (
	"errors"
	"io/ioutil"
	"net"
	"sync"
	"time"
)

const MAX_CONNS = 180

var wg sync.WaitGroup

type Server struct {
	Conns []net.Conn
}

// NewServer - (ALL) This returns a pointer to a struct consisting of
// data used for a shell server.
func NewServer() *Server {
	sh := &Server{Conns: make([]net.Conn, 0, MAX_CONNS)}
	return sh
}

// listener - (SERVER) This is a thread-safe function that uses
// sync.WaitGroup to be functioning as a separate listening thread
// for listening for new shells to connect to a Server.
func (ss *Server) listener(port string) error {
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
		// sleep for 10 ms to allow accesses to resources locked by waitgroup
		time.Sleep(10 * time.Millisecond)
	}
}

// SpawnListenerDaemon - (SERVER) This is just calls a listening daemon
// goroutine. It is also thread-safe
func (ss *Server) SpawnListenerDaemon(port string) {
	wg.Add(1)
	defer wg.Done()

	go ss.listener(port)
}

// GetConnectionCount - (SERVER) This should be used to return the number of
// connected shells. It is also thread-safe.
func (ss *Server) GetConnectionCount() int {
	wg.Add(1)
	defer wg.Done()

	return len(ss.Conns)
}

// GetConnectedIPS - (SERVER) This is used to return a slice of strings
// of the ip addresses of connected shells.
func (ss *Server) GetConnectedIPS() []string {
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
func (ss *Server) SpawnShell(dataToSend string, connectionIndex int) (string, error) {
	wg.Add(1)
	defer wg.Done()

	// send the data
	ss.Conns[connectionIndex].Write([]byte(dataToSend))
	dataRecved, err := ioutil.ReadAll(ss.Conns[connectionIndex])
	if err != nil {
		return "", errors.New("unable to spawn shell")
	}
	dataRecvedStr := string(dataRecved[:])
	// sleep for 10 ms to allow accesses to resources locked by waitgroup
	time.Sleep(10 * time.Millisecond)
	return dataRecvedStr, nil
}
