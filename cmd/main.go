package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/yushasama/tori/config"
	"github.com/yushasama/tori/monitor"
)

func profiler() {
	go func() {
		TICKER := time.NewTicker(30 * time.Second)
		defer TICKER.Stop()

		os.MkdirAll("profiles", 0755)
		runtime.SetBlockProfileRate(1)
		runtime.SetMutexProfileFraction(1)

		for t := range TICKER.C {
			timestamp := t.Format("2006-01-02_15-04-05")

			// === HEAP PROFILE ===
			heapPath := fmt.Sprintf("profiles/HEAP_%s.prof", timestamp)
			if f, err := os.Create(heapPath); err == nil {
				pprof.WriteHeapProfile(f)
				log.Printf("[PROFILER] HEAP PROFILE WRITTEN: %s", heapPath)
				f.Close()
			} else {
				log.Printf("[PROFILER] FAILED TO CREATE HEAP PROFILE: %v", err)
			}

			// === CPU PROFILE ===
			cpuPath := fmt.Sprintf("profiles/CPU_%s.prof", timestamp)
			if f, err := os.Create(cpuPath); err == nil {
				if err := pprof.StartCPUProfile(f); err == nil {
					time.Sleep(10 * time.Second)
					pprof.StopCPUProfile()
					log.Printf("[PROFILER] CPU PROFILE WRITTEN: %s", cpuPath)
				} else {
					log.Printf("[PROFILER] FAILED TO START CPU PROFILE: %v", err)
				}
				f.Close()
			} else {
				log.Printf("[PROFILER] FAILED TO CREATE CPU PROFILE: %v", err)
			}

			// === BLOCK PROFILE ===
			blockPath := fmt.Sprintf("profiles/BLOCK_%s.prof", timestamp)
			if f, err := os.Create(blockPath); err == nil {
				pprof.Lookup("block").WriteTo(f, 0)
				log.Printf("[PROFILER] BLOCK PROFILE WRITTEN: %s", blockPath)
				f.Close()
			}

			// === MUTEX PROFILE ===
			mutexPath := fmt.Sprintf("profiles/MUTEX_%s.prof", timestamp)
			if f, err := os.Create(mutexPath); err == nil {
				pprof.Lookup("mutex").WriteTo(f, 0)
				log.Printf("[PROFILER] MUTEX PROFILE WRITTEN: %s", mutexPath)
				f.Close()
			}

			// === RUNTIME STATS LOG ===
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			log.Printf("[MEMSTATS] Alloc=%dKB | TotalAlloc=%dKB | Sys=%dKB | NumGC=%d | LastGC=%v | NextGC=%dKB | Goroutines=%d",
				m.Alloc/1024, m.TotalAlloc/1024, m.Sys/1024, m.NumGC, time.Unix(0, int64(m.LastGC)), m.NextGC/1024, runtime.NumGoroutine())
		}
	}()
}

func main() {
	go func() {
		log.Println("[PPROF] Starting pprof server on http://localhost:6000/debug/pprof")
		if err := http.ListenAndServe("localhost:6000", nil); err != nil {
			log.Fatalf("[PPROF ERROR] %v", err)
		}
	}()

	profiler()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatalf("[CONFIG ERROR] %v", err)
	}

	if err := monitor.Start(ctx, cfg); err != nil {
		log.Fatalf("[MONITOR ERROR] %v", err)
	}
}
