#!/bin/bash
go test -c
./jet.test -test.bench . -test.benchtime 3s -test.cpuprofile cpu.out
go tool pprof --text --lines jet.test cpu.out