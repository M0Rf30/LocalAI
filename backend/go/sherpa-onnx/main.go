package main

// Note: this is started internally by LocalAI and a server is allocated for each model

import (
	"flag"
	"fmt"
	"os"

	grpc "github.com/mudler/LocalAI/pkg/grpc"
)

var (
	addr = flag.String("addr", "localhost:50051", "the address to connect to")
	help = flag.Bool("help", false, "show help")
)

func main() {
	flag.Parse()

	if *help {
		fmt.Printf("Sherpa-ONNX backend for LocalAI\n")
		fmt.Printf("Usage: %s [options]\n", os.Args[0])
		fmt.Printf("Options:\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if err := grpc.StartServer(*addr, &SherpaONNX{}); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start server: %v\n", err)
		os.Exit(1)
	}
}