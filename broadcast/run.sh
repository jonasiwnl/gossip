#!/bin/bash

go build
../maelstrom/maelstrom test -w broadcast --bin maelstrom-broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100 --nemesis partition
