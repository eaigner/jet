#!/bin/bash
go test -race -run "^(Test|Benchmark)IntPg[A-Z](.*)"