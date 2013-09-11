#!/bin/bash
go test -race -run "^(Test|Benchmark)_(.*)"