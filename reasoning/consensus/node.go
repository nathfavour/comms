package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type NodeMeta struct {
	ID   string `json:"id"`
	Sock string `json:"sock"`
}

type Message struct {
	From     string `json:"from"`
	Type     string `json:"type"` // "DECISION", "VOTE_LEADER"
	Payload  string `json:"payload"`
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run node.go <id>")
	}
	id := os.Args[1]
	absPath, _ := filepath.Abs(".")
	sockPath := filepath.Join(absPath, fmt.Sprintf("%s.sock", id))

	dump(id, "Initializing. Phase: Discovery.")

	// 1. Discovery & Metadata Setup
	meta := NodeMeta{ID: id, Sock: sockPath}
	metaData, _ := json.Marshal(meta)
	discoveryPath := filepath.Join(absPath, id+".discovery")
	os.WriteFile(discoveryPath, metaData, 0644)
	defer os.Remove(discoveryPath)

	// Listen for peers
	l, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	defer os.Remove(sockPath)

	var peers []NodeMeta
	var mu sync.Mutex
	decisions := make(map[string]string)
	votes := make(map[string]string)

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil { return }
			go handleConn(id, conn, &mu, decisions, votes)
		}
	}()

	// 2. Wait for all 5 nodes to appear
	for {
		files, _ := os.ReadDir(absPath)
		var currentPeers []NodeMeta
		for _, f := range files {
			if filepath.Ext(f.Name()) == ".discovery" {
				data, _ := os.ReadFile(filepath.Join(absPath, f.Name()))
				var m NodeMeta
				if err := json.Unmarshal(data, &m); err == nil {
					currentPeers = append(currentPeers, m)
				}
			}
		}
		if len(currentPeers) == 5 {
			peers = currentPeers
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	dump(id, "Discovery complete. All 5 nodes active.")

	// 3. Exchange Decisions
	myDecision := fmt.Sprintf("Decision-from-%s", id)
	broadcast(id, peers, Message{From: id, Type: "DECISION", Payload: myDecision})

	// Wait for all decisions (self + 4 peers)
	for {
		mu.Lock()
		count := len(decisions)
		mu.Unlock()
		if count >= 4 { break }
		time.Sleep(200 * time.Millisecond)
	}
	dump(id, "Phase: Decision Exchange Complete.")

	// 4. Negotiate Leader
	var ids []string
	for _, p := range peers { ids = append(ids, p.ID) }
	sort.Strings(ids)
	leaderID := ids[0]
	
	dump(id, "Proposing leader: %s", leaderID)
	broadcast(id, peers, Message{From: id, Type: "VOTE_LEADER", Payload: leaderID})

	for {
		mu.Lock()
		count := len(votes)
		mu.Unlock()
		if count >= 4 { break }
		time.Sleep(200 * time.Millisecond)
	}

	// 5. Final Action
	if id == leaderID {
		dump(id, "I am the ELECTED leader. Consolidating results.")
		finalData := fmt.Sprintf("Group Decision finalized at %s\n", time.Now().Format(time.RFC3339))
		mu.Lock()
		for node, dec := range decisions {
			finalData += fmt.Sprintf("%s: %s\n", node, dec)
		}
		mu.Unlock()
		finalData += fmt.Sprintf("%s: %s\n", id, myDecision)
		os.WriteFile("consensus_final.dump", []byte(finalData), 0644)
		dump(id, "Final dump written to consensus_final.dump")
	} else {
		dump(id, "Leader %s is handling the dump.", leaderID)
	}

	time.Sleep(1 * time.Second)
	dump(id, "Shutdown clean.")
}

func handleConn(myID string, conn net.Conn, mu *sync.Mutex, dec, votes map[string]string) {
	defer conn.Close()
	var msg Message
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&msg); err != nil { return }

	mu.Lock()
	defer mu.Unlock()
	if msg.Type == "DECISION" {
		dec[msg.From] = msg.Payload
	} else if msg.Type == "VOTE_LEADER" {
		votes[msg.From] = msg.Payload
	}
}

func broadcast(myID string, peers []NodeMeta, msg Message) {
	for _, p := range peers {
		if p.ID == myID { continue }
		// Retry a few times if the peer hasn't started listening yet
		for i := 0; i < 5; i++ {
			conn, err := net.Dial("unix", p.Sock)
			if err == nil {
				json.NewEncoder(conn).Encode(msg)
				conn.Close()
				break
			}
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func dump(id, format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	fmt.Printf("[%s] %s\n", id, msg)
	f, _ := os.OpenFile(id+".dump", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	f.WriteString(time.Now().Format(time.RFC3339) + " " + msg + "\n")
}
