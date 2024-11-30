package main

import (
	"context"
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

				var wg sync.WaitGroup
				for len(unacked) > 0 {
					// Make a copy of unacked neighbors to iterate through
					// TODO: can we improve this
					unacked_copy := make([]string, 0, len(unacked))
					unacked_mu.Lock()
					for neighbor := range unacked {
						unacked_copy = append(unacked_copy, neighbor)
					}
					unacked_mu.Unlock()

					wg.Add(len(unacked_copy))
					for _, neighbor := range unacked_copy {
						go func() {
							defer wg.Done()
							ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
							defer cancel()
							msg, err := n.SyncRPC(ctx, neighbor, body)
							if err != nil {
								return
							}

							var body map[string]any
							if err = json.Unmarshal(msg.Body, &body); err != nil {
								return
							}

							unacked_mu.Lock()
							if _, ok := unacked[neighbor]; ok && body["type"] == "broadcast_ok" {
								delete(unacked, neighbor)
							}
							unacked_mu.Unlock()
						}()
					}

					// Wait for all attempts to broadcast to finish before continuing
					wg.Wait()
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
