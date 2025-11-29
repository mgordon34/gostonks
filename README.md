# Go Stonks

Go Stonks is an algorithmic trading suite developed in Go. Capabilities of the suite will include:
- Market data collection
- Strategy discovery
- Backtesting
- Live market data collection
- Broker integration
- Portfolio management


## Structure

Go Stonks is broken up in to functional services that communicate via a message broker, these services
are coordinated as containers and orchestrated via docker compose

### Services

#### Market Data Collection

Market data is collected via the market service. This service collects data from various
sources and is in charge of sending market data events to the message broker. These events are
typically collated into the form of a candle(OHLC format) of any timeframe specified. The timeframe
will also be specified in the candle object.

Market data is stored in a postgres database for historical use cases.

If the market service is tasked with providing historical data for backtesting services, it will
gather historical data based on the requested timeframe and send a sequential Market events to the
message broker. A predefined backtesting session id will be passed to ensure the appropriate backtesting
session picks up those events.


## Technology
Backend Services: Go
Message Broker: Rabbitmq
Frontend: Nextjs
