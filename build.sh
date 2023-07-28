#!/bin/bash
clear
GOOS=linux go build -ldflags "-s -w" ct.go
GOOS=linux go build -ldflags "-s -w" ctcfg.go
echo "Built"

