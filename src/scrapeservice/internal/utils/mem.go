package utils

import (
	"fmt"
	"runtime"
)

func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Alloc: Allocated heap objects (currently in use)
	fmt.Printf("Alloc = %v MiB", m.Alloc/1024/1024)
	// TotalAlloc: Cumulative total allocated heap memory (grows continually)
	fmt.Printf("\tTotalAlloc = %v MiB", m.TotalAlloc/1024/1024)
	// Sys: Memory obtained from the OS
	fmt.Printf("\tSys = %v MiB", m.Sys/1024/1024)
	// NumGC: Number of completed Garbage Collection cycles
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}
