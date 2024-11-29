#!/bin/bash

go build
../maelstrom/maelstrom test -w echo --bin maelstrom-echo --node-count 1 --time-limit 10
