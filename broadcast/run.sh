#!/bin/bash

go build
../maelstrom/maelstrom test -w broadcast --bin maelstrom-broadcast --node-count 5 --time-limit 20 --rate 10 --nemesis partition
