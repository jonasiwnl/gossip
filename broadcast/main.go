package main

import (
	"encoding/json"
	"errors"
	"log"
	"slices"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	messages := make([]int, 0)
	neighbors := make([]string, 0)
	var messages_mu sync.Mutex

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		read_message, ok := body["message"].(float64)
		if !ok {
			return errors.New("can't convert message to type float64")
		}
		message := int(read_message)

		is_duplicated_message := slices.Contains(messages, message)
		// If this message isn't duplicated, i.e. it is _new_, add it to our messages list and propagate
		if !is_duplicated_message {
			messages_mu.Lock()
			messages = append(messages, message)
			messages_mu.Unlock()

			// Propagate
			go func() {
				unacked := make(map[string]struct{})
				var unacked_mu sync.Mutex
				for _, neighbor := range neighbors {
					unacked[neighbor] = struct{}{}
				}

				for len(unacked) != 0 {
					for neighbor := range unacked {
						n.RPC(neighbor, body, maelstrom.HandlerFunc(func(msg maelstrom.Message) error {
							var body map[string]any
							if err := json.Unmarshal(msg.Body, &body); err != nil {
								return nil
							}

							if body["type"] == "broadcast_ok" {
								unacked_mu.Lock()
								delete(unacked, neighbor)
								unacked_mu.Unlock()
							}

							return nil
						}))
					}
					// Wait to retry
					time.Sleep(10 * time.Millisecond)
				}
			}()
		}

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

		topology_map, ok := body["topology"].(map[string]interface{})
		if !ok {
			return errors.New("can't convert topology map to type map[string]interface{}")
		}

		for node, neighbors_interface := range topology_map {
			if node != n.ID() {
				continue
			}

			neighbors_slice, ok := neighbors_interface.([]interface{})
			if !ok {
				return errors.New("can't convert topology map neighbors to slice type")
			}

			neighbors = make([]string, len(neighbors_slice))

			for index, value := range neighbors_slice {
				asserted_value, ok := value.(string)
				if !ok {
					return errors.New("can't convert neighbor value to type string")
				}
				neighbors[index] = asserted_value
			}
		}

		return_body := make(map[string]string)
		return_body["type"] = "topology_ok"

		return n.Reply(msg, return_body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
