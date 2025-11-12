// Copyright 2024 Ducros Network
// Stratum Proxy for RandomX mining with xmrig compatibility

package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	// Stratum server config
	stratumAddr  = flag.String("stratum", "0.0.0.0:3333", "Stratum server listen address")
	stratumDiff  = flag.Float64("diff", 10000, "Initial difficulty for miners")

	// Geth RPC config
	gethRPC      = flag.String("geth", "http://localhost:8545", "Geth JSON-RPC endpoint")

	// Pool config
	poolAddr     = flag.String("pool-addr", "", "Pool payout address (miner etherbase)")
	poolFee      = flag.Float64("pool-fee", 1.0, "Pool fee percentage (1.0 = 1%)")

	// Logging
	verbose      = flag.Bool("v", false, "Verbose logging")

	// Mining config
	algo         = flag.String("algo", "rx/0", "RandomX algorithm variant (rx/0 for Ducros)")
)

func main() {
	flag.Parse()

	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// ASCII banner
	printBanner()

	// Validate config
	if *poolAddr == "" {
		log.Println("⚠️  WARNING: No pool address specified, using miner addresses directly")
	}

	// Create proxy server
	config := &ServerConfig{
		ListenAddr:     *stratumAddr,
		GethRPC:        *gethRPC,
		InitialDiff:    *stratumDiff,
		PoolAddress:    *poolAddr,
		PoolFee:        *poolFee,
		Verbose:        *verbose,
		Algorithm:      *algo,
	}

	server, err := NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	log.Printf("🚀 Starting Stratum proxy on %s", *stratumAddr)
	log.Printf("🔗 Connected to Geth: %s", *gethRPC)
	log.Printf("⛏️  Algorithm: %s", *algo)
	log.Printf("💎 Initial difficulty: %.0f", *stratumDiff)

	if *poolAddr != "" {
		log.Printf("💰 Pool address: %s", *poolAddr)
		log.Printf("💵 Pool fee: %.2f%%", *poolFee)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	log.Println("✅ Stratum proxy running. Press Ctrl+C to stop.")
	<-sigCh

	log.Println("🛑 Shutting down...")
	server.Stop()
	log.Println("👋 Goodbye!")
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║   ████████╗ ██╗   ██╗  ██████╗ ██████╗   ██████╗  ███████║
║   ██╔══██║ ██║   ██║ ██╔════╝ ██╔══██╗ ██╔═══██╗ ██╔════║
║   ██║  ██║ ██║   ██║ ██║      ██████╔╝ ██║   ██║ ███████║
║   ██║  ██║ ██║   ██║ ██║      ██╔══██╗ ██║   ██║ ╚════██║
║   ██████╔╝ ╚██████╔╝ ╚██████╗ ██║  ██║ ╚██████╔╝ ███████║
║   ╚═════╝   ╚═════╝   ╚═════╝ ╚═╝  ╚═╝  ╚═════╝  ╚══════║
║                                                           ║
║        Stratum Proxy - RandomX Mining Bridge             ║
║                  xmrig Compatible                         ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
`
	log.Println(banner)
}
