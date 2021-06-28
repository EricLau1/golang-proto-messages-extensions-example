#!/bin/bash

protoc -I=./messages --go_out=plugins=grpc:. ./messages/*.proto
