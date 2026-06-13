package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Metadata struct {
	ID        string   `json:"id"`
	PID       int      `json:"pid"`
	Path      string   `json:"path"`
	Preferred string   `json:"preferred"`
	Resources []string `json:"resources"`
}

func main() {
	id := "node2"
	absPath, _ := filepath.Abs(".")
	meta := Metadata{
		ID:        id,
		PID:       os.Getpid(),
		Path:      absPath,
		Preferred: "unix",
		Resources: []string{"cpu:high", "disk:low"},
	}

	dump(id, "Phase: Initialization. Metadata: %+v", meta)

	discoveryFile := filepath.Join(absPath, id+".discovery")
	data, _ := json.Marshal(meta)
	os.WriteFile(discoveryFile, data, 0644)
	defer os.Remove(discoveryFile)

	var peerMeta Metadata
	for {
		files, _ := os.ReadDir(absPath)
		for _, f := range files {
			if filepath.Ext(f.Name()) == ".discovery" && f.Name() != id+".discovery" {
				content, _ := os.ReadFile(filepath.Join(absPath, f.Name()))
				json.Unmarshal(content, &peerMeta)
				dump(id, "Phase: Discovery. Found peer: %s", peerMeta.ID)
				goto Bargain
			}
		}
		time.Sleep(1 * time.Second)
	}

Bargain:
	sockPath := filepath.Join(absPath, "comms.sock")
	if id < peerMeta.ID {
		listen(id, sockPath)
	} else {
		dial(id, sockPath)
	}
}

func listen(id, path string) {
	os.Remove(path)
	l, err := net.Listen("unix", path)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	conn, err := l.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	
	handle(id, conn, true)
}

func dial(id, path string) {
	time.Sleep(2 * time.Second)
	conn, err := net.Dial("unix", path)
	if err != nil {
		return
	}
	defer conn.Close()
	
	handle(id, conn, false)
}

func handle(id string, conn net.Conn, isServer bool) {
	reader := bufio.NewReader(conn)
	
	if !isServer {
		fmt.Fprintf(conn, "PROPOSE_DUMP:shared_experiment.dump\n")
		dump(id, "Bargaining: Proposed shared_experiment.dump")
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		
		if strings.HasPrefix(line, "PROPOSE_DUMP:") {
			suggestion := strings.Split(line, ":")[1]
			dump(id, "Bargaining: Peer proposed %s. I accept.", suggestion)
			fmt.Fprintf(conn, "ACCEPT_DUMP:%s\n", suggestion)
		} else if strings.HasPrefix(line, "ACCEPT_DUMP:") {
			agreed := strings.Split(line, ":")[1]
			dump(id, "Bargaining: Peer accepted %s. Finalizing.", agreed)
			fmt.Fprintf(conn, "FINALIZED\n")
			break
		} else if line == "FINALIZED" {
			break
		}
	}
	dump(id, "Communication complete.")
}

func dump(id, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fmt.Printf("[%s] %s\n", id, msg)
	f, _ := os.OpenFile(id+".dump", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(time.Now().Format(time.RFC3339) + " " + msg + "\n")
}
