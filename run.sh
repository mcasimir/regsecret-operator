#!/bin/bash

set -e

export REGSECRET_OPERATOR_CONFIG=`cat config.example.json`

go run main.go --run-outside-cluster