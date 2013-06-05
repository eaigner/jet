#!/bin/bash
go test -bench . -cpuprofile cpu.out
go tool pprof --text jet.test cpu.out