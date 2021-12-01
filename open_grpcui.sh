#!/usr/bin/env bash

grpcui -proto greeting.proto -import-path=protos/ -plaintext localhost:50051
