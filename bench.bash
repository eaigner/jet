#!/bin/bash
go test -bench . -cpuprofile cpu.out
go tool pprof --text --lines jet.test cpu.out