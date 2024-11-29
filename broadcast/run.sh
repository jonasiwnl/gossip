#!/bin/bash

go build
../maelstrom/maelstrom test -w broadcast --bin maelstrom-broadcast --node-count 1 --time-limit 20 --rate 10
