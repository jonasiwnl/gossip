package main

import (
	"encoding/json"
	"log"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	messages := []int{}
	var mu sync.Mutex
	// var topology map[string][]string

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		mu.Lock()
		messages = append(messages, int(body["message"].(float64)))
		mu.Unlock()

		return_body := make(map[string]string)
		return_body["type"] = "broadcast_ok"

		return n.Reply(msg, return_body)
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		return_body := make(map[string]any)
		return_body["type"] = "read_ok"
		return_body["messages"] = messages

		return n.Reply(msg, return_body)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// topology = body["topology"].(map[string][]string)
		return_body := make(map[string]string)
		return_body["type"] = "topology_ok"

		return n.Reply(msg, return_body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
