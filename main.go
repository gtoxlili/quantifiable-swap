package main

import (
	"github.com/gtoxlili/quantifiable-swap/provider"
	"github.com/gtoxlili/quantifiable-swap/swap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	okxProvider := provider.NewOkx()
	// bitGetProvider := provider.NewBitGet()
	bnProvider := provider.NewBinance().InjectOrderFunc(okxProvider.MarketOrder)
	byBitProvider := provider.NewByBit()

	eth15mWaper := swap.NewWaper("eth", "usdt", 15*time.Minute, "50", "200", okxProvider)
	eth1hWaper := swap.NewWaper("eth", "usdt", 1*time.Hour, "100", "300", bnProvider)
	eth4hWaper := swap.NewWaper("eth", "usdt", 4*time.Hour, "100", "600", bnProvider)
	aero1hWaper := swap.NewWaper("aero", "usdt", 1*time.Hour, "50", "200", byBitProvider)
	ondo15mWaper := swap.NewWaper("ondo", "usdt", 15*time.Minute, "50", "100", okxProvider)
	btc15mNotify := swap.NewNotify("btc", "usdt", 15*time.Minute, bnProvider)

	go eth15mWaper.RunWithCustomPeriod(21)
	go ondo15mWaper.RunWithCustomPeriod(21)
	go eth1hWaper.Run()
	go eth4hWaper.Run()
	go btc15mNotify.RunWithCustomPeriod(21)
	go aero1hWaper.Run()

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-signChan
}
