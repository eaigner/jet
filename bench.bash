#!/bin/bash
go test -bench . -benchtime 5s -cpuprofile cpu.out
go tool pprof --text --lines jet.test cpu.out