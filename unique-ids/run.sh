#!/bin/bash

go build
../maelstrom/maelstrom test -w unique-ids --bin maelstrom-unique-ids --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition
