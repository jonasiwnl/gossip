package main

import (
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
		pid := strconv.Itoa(os.Getpid())
		ts := strconv.FormatInt(time.Now().UnixNano(), 10)
		uuid, err := uuid.NewRandom()
		if err != nil {
			return err
		}
		guid := strings.Join([]string{pid, ts, uuid.String()}, ".")

		return_body := make(map[string]string)
		return_body["type"] = "generate_ok"
		return_body["id"] = guid

		return n.Reply(msg, return_body)
	})

	if err := n.Run(); err != nil {
		log.Fatal(err)
	}
}
