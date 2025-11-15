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

	// VarDiff config
	varDiffTarget = flag.Float64("vardiff-target", 30.0, "Target time between shares in seconds")
	varDiffWindow = flag.Uint64("vardiff-window", 10, "Number of shares for vardiff calculation")

	// Ban system config
	maxInvalidStreak = flag.Uint64("max-invalid-streak", 10, "Max consecutive invalid shares before ban (0 = disabled)")

	// DoS protection config
	maxConnections = flag.Int("max-connections", 1000, "Max concurrent connections (0 = unlimited)")
	shareRateLimit = flag.Float64("share-rate-limit", 100.0, "Max shares per second per miner (0 = unlimited)")

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
		ListenAddr:         *stratumAddr,
		GethRPC:            *gethRPC,
		InitialDiff:        *stratumDiff,
		PoolAddress:        *poolAddr,
		PoolFee:            *poolFee,
		Verbose:            *verbose,
		Algorithm:          *algo,
		VarDiffTarget:      *varDiffTarget,
		VarDiffWindow:      *varDiffWindow,
		MaxInvalidStreak:   *maxInvalidStreak,
		MaxConnections:     *maxConnections,
		ShareRateLimit:     *shareRateLimit,
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

	log.Printf("⚙️  VarDiff: target %.1fs, window %d shares", *varDiffTarget, *varDiffWindow)
	if *maxInvalidStreak > 0 {
		log.Printf("🛡️  Ban system: max %d invalid shares", *maxInvalidStreak)
	} else {
		log.Printf("🛡️  Ban system: disabled")
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
