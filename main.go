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
	//bitGetProvider := provider.NewBitGet()
	bnProvider := provider.NewBinance().InjectOrderFunc(okxProvider.MarketOrder)
	bybitProvider := provider.NewByBit().InjectOrderFunc(okxProvider.MarketOrder)

	eth15mWaper := swap.NewRSIWaper("eth", "usdt", 15*time.Minute, "50", "100", okxProvider)
	eth1hWaper := swap.NewRSIWaper("eth", "usdt", 1*time.Hour, "50", "200", bybitProvider)
	eth4hWaper := swap.NewRSIWaper("eth", "usdt", 4*time.Hour, "100", "400", bnProvider)
	btc15mNotify := swap.NewRSINotify("btc", "usdt", 15*time.Minute, bnProvider)
	aero1hNotify := swap.NewRSINotify("aero", "usdt", 1*time.Hour, bybitProvider)

	go eth15mWaper.Run()
	go eth1hWaper.Run()
	go eth4hWaper.Run()
	go btc15mNotify.Run()
	go aero1hNotify.Run()

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-signChan
}
