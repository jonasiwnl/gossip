package main

import (
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	n.Handle("generate", func(msg maelstrom.Message) error {
		// Unmarshal the message body as an loosely-typed map.
		var body map[string]any
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		pid := strconv.Itoa(os.Getpid())
		ts := strconv.FormatInt(time.Now().UnixNano(), 10)
		uuid, err := uuid.NewRandom()
		if err != nil {
			log.Fatal(err)
		}
		guid := strings.Join([]string{pid, ts, uuid.String()}, ".")

		body["type"] = "generate_ok"
		body["id"] = guid

		return n.Reply(msg, body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
