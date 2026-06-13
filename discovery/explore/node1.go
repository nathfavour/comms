package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

type Metadata struct {
	ID        string `json:"id"`
	PID       int    `json:"pid"`
	Path      string `json:"path"`
	Preferred string `json:"preferred"`
}

func main() {
	id := "node1"
	absPath, _ := filepath.Abs(".")
	meta := Metadata{
		ID:        id,
		PID:       os.Getpid(),
		Path:      absPath,
		Preferred: "unix",
	}

	dump(id, "Starting up. My metadata: %+v", meta)

	// Step 1: Advertise presence
	discoveryFile := filepath.Join(absPath, id+".discovery")
	data, _ := json.Marshal(meta)
	os.WriteFile(discoveryFile, data, 0644)
	defer os.Remove(discoveryFile)

	dump(id, "Discovery file created: %s", discoveryFile)

	// Step 2: Discovery loop
	var peerMeta Metadata
	for {
		files, _ := os.ReadDir(absPath)
		for _, f := range files {
			if filepath.Ext(f.Name()) == ".discovery" && f.Name() != id+".discovery" {
				content, _ := os.ReadFile(filepath.Join(absPath, f.Name()))
				json.Unmarshal(content, &peerMeta)
				dump(id, "Discovered peer: %s (PID: %d)", peerMeta.ID, peerMeta.PID)
				goto Bargain
			}
		}
		time.Sleep(1 * time.Second)
	}

Bargain:
	// Step 3: Negotiate protocol
	// Simple rule: node with lower ID listens, higher ID dials
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
	dump(id, "Listening on %s", path)

	conn, err := l.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	dump(id, "Peer connected. Negotiating...")
	
	handle(id, conn)
}

func dial(id, path string) {
	// Wait a bit for listener
	time.Sleep(2 * time.Second)
	dump(id, "Dialing %s", path)
	conn, err := net.Dial("unix", path)
	if err != nil {
		dump(id, "Dial failed: %v", err)
		return
	}
	defer conn.Close()
	dump(id, "Connected to peer.")
	
	handle(id, conn)
}

func handle(id string, conn net.Conn) {
	fmt.Fprintf(conn, "Hello from %s\n", id)
	buf := make([]byte, 1024)
	n, _ := conn.Read(buf)
	dump(id, "Received: %s", string(buf[:n]))
}

func dump(id, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fmt.Printf("[%s] %s\n", id, msg)
	f, _ := os.OpenFile(id+".dump", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(time.Now().Format(time.RFC3339) + " " + msg + "\n")
}
