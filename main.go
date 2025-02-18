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

	eth15mWaper := swap.NewWaper("eth", "usdt", 15*time.Minute, "150", "300", okxProvider)
	eth1hWaper := swap.NewWaper("eth", "usdt", 1*time.Hour, "200", "400", bnProvider)
	eth4hWaper := swap.NewWaper("eth", "usdt", 4*time.Hour, "300", "600", bnProvider)
	aero1hWaper := swap.NewWaper("aero", "usdt", 1*time.Hour, "100", "200", byBitProvider)
	btc15mNotify := swap.NewNotify("btc", "usdt", 15*time.Minute, bnProvider)
	aero15mNotify := swap.NewWaper("aero", "usdt", 15*time.Minute, "100", "100", byBitProvider)

	go eth15mWaper.Run()
	go eth1hWaper.Run()
	go eth4hWaper.Run()
	go btc15mNotify.Run()
	go aero1hWaper.Run()
	go aero15mNotify.Run()

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-signChan
}
