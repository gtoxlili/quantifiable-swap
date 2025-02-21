# Quantifiable Swap

This project is an automated trading and monitoring system implemented in Go. It provides a flexible "Job" task management mechanism, supports multiple data sources for market data, calculates various technical indicators, and integrates multi-channel log notifications (e.g., Telegram, Bark). It aims to offer an extensible and maintainable solution for quantitative trading in the cryptocurrency space.

----------------------------------------------------------------------------------------------------

## Table of Contents
- [Quantifiable Swap](#quantifiable-swap)
  - [Table of Contents](#table-of-contents)
  - [Project Overview](#project-overview)
  - [Key Features](#key-features)
  - [Core Modules](#core-modules)
    - [Job Management](#job-management)
    - [Data Providers](#data-providers)
    - [Technical Indicators](#technical-indicators)
    - [Logging System](#logging-system)
    - [Configuration Management](#configuration-management)
  - [Getting Started](#getting-started)
    - [1. Environment \& Dependencies](#1-environment--dependencies)
    - [2. Building \& Packaging](#2-building--packaging)
    - [3. Deployment \& Launch](#3-deployment--launch)
    - [4. Configuration File](#4-configuration-file)
  - [Taskflow (Taskfile)](#taskflow-taskfile)
  - [FAQ](#faq)
  - [License](#license)
  - [More Info](#more-info)

----------------------------------------------------------------------------------------------------

## Project Overview

Quantifiable Swap is an automated quantitative trading program designed for crypto futures/spots and related scenarios. It mainly consists of:

1. Multiple Exchange Integrations: Currently supports Okx and Bybit, using an interface-oriented design that can be easily extended to other exchanges.  
2. Real-Time & Historical Market Data: Capable of fetching historical candlesticks and real-time quotations, enabling strategy and indicator calculations.  
3. Technical Indicator Calculations: Includes indicators such as RSI (Relative Strength Index) and moving averages (MA). Combined with trading strategies, these indicators help determine when to place orders.  
4. Strategy Execution & Order Management: When a buy or sell signal is detected, orders are automatically placed through the provider interface with comprehensive logging.  
5. Multi-Channel Log Notifications: Supports console output, Telegram Bot, Bark App, etc. for real-time alerts on trading anomalies or triggered strategies.  

The project is written in Go with minimal dependencies and can run on multiple platforms including Linux and macOS.

----------------------------------------------------------------------------------------------------

## Key Features

1. **Job Task Management**  
   - Provides a unified Manager that can dynamically create, delete, start, and stop strategies (Jobs).  
   - Supports an independent subscriber list, allowing different users to share or subscribe to trade strategy logs and notifications.

2. **Multiple Exchange Data Sources**  
   - Uses an interface-based architecture for consistency. Currently implements OkxProvider, ByBitProvider, and PolymericProvider.  
   - Each Provider supports fetching historical K-lines and placing market orders, applicable to both spot and futures trading.

3. **Comprehensive Technical Indicators**  
   - Built-in calculations for common indicators like RSI and MA, implemented in an extensible way (within the indicator folder).  
   - Adopts methods like sliding windows or caching to efficiently compute the latest indicator values.

4. **Flexible Notification & Logging System**  
   - Integrates multiple logging outputs: console (console), Telegram (tglog), and Bark (bark).  
   - Different log levels (info, warn, error) can trigger automatic alerts or notifications (e.g., pinning a message in Telegram).  
   - The unified dynamic subscriber management allows pushing trading signals to multiple chat channels or groups with ease.

5. **Taskfile for Streamlined Building & Deployment**  
   - A Taskfile.yml manages local packaging and remote deployment, enabling one-command automation from build to production.  
   - Includes cross-compilation examples for multiple platforms, which can also be coupled with CI/CD pipelines.

----------------------------------------------------------------------------------------------------

## Core Modules

### [Job Management](job)
- Defines the Job structure and the Manager.  
- Enables creation, deletion, start, and stop actions for Jobs either via configuration or dynamic calls.  
- Each Job corresponds to one trading strategy and a list of subscribers, executing tasks independently and sending out log notifications.

### [Data Providers](market)
- Provides unified interfaces for various exchanges or data sources, including market data and order execution.  
- OkxProvider, ByBitProvider, and PolymericProvider are implemented; others can be added as needed.  
- Supports custom rate limiters, K-line data parsing, and handling order confirmations.

### [Technical Indicators](indicator)
- Houses the logic for various indicators, for instance, RSI and MA.  
- Offers a standardized Indicator interface that supports custom and compound indicators.  
- Easily referenced or combined in trading strategies.

### [Logging System](logger)
- Exposes a unified Logger interface with configurable outputs: console, Telegram, Bark, etc.  
- Supports info, warn, and error levels, where higher levels may trigger alerts (e.g., pinned messages in Telegram).  
- Integrates `pretty` console/tglog/bark modules for flexible rendering and distribution of logs.

### [Configuration Management](common/config)
- Parses and saves config.yaml, allowing local persistence of current strategy settings.  
- Job, Provider, and Credential details can all be declared or modified in this file.  
- Loads configuration on startup and can also save changes at runtime.

----------------------------------------------------------------------------------------------------

## Getting Started

### 1. Environment & Dependencies
1. Install Go (>= 1.18+).  
2. Enable Go Modules (default for Go 1.11+).  
3. Register required exchange accounts and obtain API Key / Secret / Passphrase as needed.  
4. If you plan to use Telegram or Bark notifications, prepare your Bot Token, Channel ID, or app token.

### 2. Building & Packaging
In the project root directory, you can either use the Taskfile or manually use go build:

• Using GitHub Actions (Recommended)
  - Fork this project and configure the following Secrets:
    - `SSH_PRIVATE_KEY`: SSH private key for deployment server
    - `REMOTE_ADDRESS`: Remote server address
    - `OKX_API_KEY`/`OKX_SECRET_KEY`/`OKX_PASSPHRASE`: OKX API credentials
    - `BYBIT_API_KEY`/`BYBIT_API_SECRET`: ByBit API credentials  
    - `TG_BOT_TOKEN`/`TG_CHAT_ID`: Telegram bot configuration
    - `BARK_TOKEN`: Bark notification token
    - `PROXY_ADDR`: Proxy address (optional)
  - Push to main branch or manually trigger workflow to build and deploy

• Using Taskfile
  - Install the [Task](https://taskfile.dev/) tool, then in the project root directory run:  
    » task  
    (By default, this will compile for linux-amd64 and may upload the build, depending on your Taskfile.yml settings.)  
  - Or run:  
    » task linux-amd64  
    » task darwin-amd64  
    to cross-compile for the desired platform. Compilation outputs go to the dist/ directory.

• Manual Build  
  - From the root directory:  
    » go mod download  
    » go build -o dist/quantifiable-swap  
  - The resulting binary will be at dist/quantifiable-swap.

### 3. Deployment & Launch
1. GitHub Actions Automated Deployment
   - Configure repository Secrets and Variables
   - Push code to main branch to trigger automatic build
   - Actions will compile, deploy and start service via PM2

2. Manual Deployment  
   - Upload the dist/quantifiable-swap binary and your config.yaml to your server.  
   - Navigate to the directory and run ./quantifiable-swap to start; you can use PM2, Supervisor, etc. for process management.  

3. Automated Deployment via Taskfile  
   - Configure REMOTE_ADDRESS, REMOTE_DIR, etc. in Taskfile.yml.  
   - Run » task deploy-binary (or the default » task) to automatically upload the binary and start it under PM2.  
   - If needed, run » task deploy-config to deploy your configuration file.

### 4. Configuration File
Below is a simplified sample of config.yaml:
```yaml
- type: trader
  provider:
    name: "okx"
    injectOrder: ""
  symbol:
    base: "BTC"
    quote: "USDT"
  bar: "15"           # K-line timeframe in minutes
  Subscribers: [584544685]  # List of Telegram IDs
```

- type: Strategy or task type, e.g., trader, monitor, etc.  
- provider: Specifies the data provider (okx, bybit, etc.) and optionally an injectOrder provider.  
- symbol: Base/quote pair.  
- bar: Candlestick timeframe.  
- Subscribers: Sends logs/notifications to the specified Telegram ID(s).

When the program exits, it automatically writes all active Jobs back to config.yaml for quick reloading next time.

----------------------------------------------------------------------------------------------------

## Taskflow (Taskfile)

The Taskfile.yml file in this project provides these common tasks:
- clean: Removes and recreates the dist directory.  
- build: Cross-compiles the program, supporting environment variable injections (API Key, Passphrase, etc.).  
- deploy-binary: Uploads the build artifacts and starts them via PM2 on the remote server.  
- deploy-config: Uploads or retrieves the configuration file.  
- pull-config: Retrieves the latest remote configuration file.  
- linux-amd64 / darwin-amd64: Cross-compile for Linux/macOS amd64 respectively.

These tasks rely on variables injected into Taskfile.yml (e.g., TG_BOT_TOKEN, OKX_API_KEY, REMOTE_ADDRESS). Adjust them according to your environment before use.

----------------------------------------------------------------------------------------------------

## FAQ

1. **How do I switch between exchanges?**  
   Change provider.name in config.yaml to the desired exchange identifier (okx / bybit / polymeric) to use the corresponding Provider.

2. **How do I add a new indicator or custom logic?**  
   - Create a new file under indicator/ for indicator calculations, referencing existing RSI or MA implementations.  
   - Incorporate it into your strategy logic by integrating with Job scheduling.

3. **What if I encounter API timeouts or rate limits?**  
   The market/ modules implement basic limiters and retry mechanisms. You can add custom retry logic or exponential backoff as needed.

4. **How to handle too many or too few log notifications?**  
   - Adjust logging behavior and levels in logger.go or the Job’s notification triggers.  
   - You can also modify the pretty/* modules’ handleXXX() methods to further filter or aggregate log output.

----------------------------------------------------------------------------------------------------

## License

This project is released under the GNU General Public License v3.0, granting you the freedom to access, modify, and redistribute the source code under the same license terms.  
For more details, see the LICENSE file in the repository.

----------------------------------------------------------------------------------------------------

## More Info

- For deeper insights into each exchange API or strategy logic, refer to the relevant provider files or indicator implementations under indicator/.  
- main.go provides the application’s entry point, where initialization, configuration loading, and Bot startup procedures are orchestrated.  
- Feel free to open an issue for any questions or suggestions.  

Contact information is available on the project’s GitHub page for business or collaboration inquiries. Happy trading and best of luck with your strategies!
