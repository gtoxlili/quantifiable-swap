# Quantifiable Swap

This project provides a simplified framework for RSI-based auto-trading. By integrating data from multiple providers (OKX, BitGet, Binance), it leverages real-time price sequences to detect oversold or overbought signals and can initiate trades using market orders.

## Features
- Universal Provider Interface  
  Easily switch among OKX, BitGet, and Binance or even combine them with the built-in PolymericProvider for aggregated data.

- RSI Calculation  
  The RSI hooking system (in quantifiable/rsi.go) efficiently computes and updates RSI values in real-time. It supports smoothing by tracking average gains and losses with minimal overhead.

- Customizable Signals  
  Default buy/sell logic checks recent RSI trends and thresholds. You can implement custom “canBuy” or “canSell” methods to tailor your trading strategy.

- Auto-Trading Logic  
  The swap/RSIWaper struct manages the main trading loop, periodically fetching new price updates and triggering trades when signals meet criteria (e.g., curRSI < 30 or curRSI > 70).

- Configurable Build System  
    [`Taskfile.dist.yml`](./Taskfile.dist.yml) and [`Makefile.example`](./Makefile.example) demonstrate a modern approach to building and deploying binaries with minimal environment dependencies.

## Project Structure
- quantifiable/  
  Contains RSI core logic and initialization.  
- swap/  
  Implements the trading workflow, including RSIWaper for auto-trading.  
- provider/  
  Defines data sources (OKX, BitGet, Binance, etc.) that conform to a common Provider interface.  
- sequence/  
  Maintains in-memory candle data, allowing easy concurrency-safe updates.  
- client/  
  Supplies an HTTP client with optional proxy.  
- constants/  
  Stores placeholder credential fields.

## Quick Start
1. Clone the repository.  
2. Inject credentials in build files:
    - In `Taskfile.dist.yml`: Update `OKX_API_KEY`, `OKX_API_SECRET`, and `OKX_API_PASSPHRASE`
    - Or in `Makefile.example`: Set the corresponding environment variables
3. Build with `task build` or `make build`.  
4. Run the generated executable to start your RSI-based trading and notifications.

## Code Highlights
- Efficient RSI Hook: Uses historical average gains/losses for faster incremental updates.  
- Modular Provider: Each exchange adapter is swappable at runtime, enabling flexible multi-source data.  
- Smart PriceSequence: Gathers 1-minute candles while capping array size for memory efficiency.  
- Clean Build Scripts: The Taskfile and Makefile show how to bundle environment variables dynamically.

Feel free to customize or extend this project according to your own trading requirements. Feedback and contributions are welcome!