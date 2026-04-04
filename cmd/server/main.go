package main

import (
	"flag"
	"log"

	appserver "github.com/zxh326/kite/internal/server"
	"k8s.io/klog/v2"
	_ "net/http/pprof"
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	if err := appserver.RunUntilSignal("localhost:6060"); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
