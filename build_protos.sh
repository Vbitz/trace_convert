#!/bin/bash

export PATH=$HOME/go/bin/:$PATH

PERFETTO_PATH=$HOME/dev/org/perfetto

protoc --go_out=protos \
    --go_opt=paths=source_relative \
    -I=$PERFETTO_PATH \
    $PERFETTO_PATH/protos/perfetto/trace/perfetto_trace.proto