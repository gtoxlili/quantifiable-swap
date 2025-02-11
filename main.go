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
	bitGetProvider := provider.NewBitGet()
	bnProvider := provider.NewBinance()
	bybitProvider := provider.NewByBit()

	ethWaper := swap.NewRSIWaper("eth", "usdt", 15*time.Minute, "100", "400", okxProvider)
	btcWaper := swap.NewRSIWaper("btc", "usdt", 15*time.Minute, "50", "200", okxProvider)
	aeroWaper := swap.NewRSIWaper("aero", "usdt", 15*time.Minute, "50", "200", bybitProvider)
	ondoNotify := swap.NewRSINotify("ondo", "usdt", 1*time.Hour, bitGetProvider)
	ethNotify := swap.NewRSINotify("eth", "usdt", 15*time.Minute, bnProvider)

	go ethWaper.Run()
	go ondoNotify.Run()
	go btcWaper.Run()
	go aeroWaper.Run()
	go ethNotify.Run()

	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-signChan

	ethWaper.Stop()
	btcWaper.Stop()
	ondoNotify.Stop()
	aeroWaper.Stop()
	ethNotify.Stop()
}
