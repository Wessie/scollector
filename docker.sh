#!/bin/sh

set -e

DIR=/go/src/github.com/bosun-monitor/scollector
docker run --rm -v "$(pwd)":$DIR -w $DIR golang:cross sh build.sh
