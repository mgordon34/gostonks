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
are coordinated as containers and orchestrated via docker compose. Each service owns its own top-level
directory (e.g. `market/`) with a conventional Go layout:

- `market/cmd/market`: container entrypoint binaries.
- `market/internal/...`: service-scoped packages such as `market/internal/ingest` for data ingestion handlers.
- `internal/config`: shared utilities available to all services (currently hosts the env helper used to read Redis/Postgres configuration).
- `deploy/<service>`: Dockerfiles or container assets specific to that service.

This keeps cross-service boundaries explicit while allowing shared code to live under `internal/`.

### Services

#### Market Data Collection

Market data is collected via the market service. This service collects data from various
sources and is in charge of sending market data events to the message broker. These events are
typically collated into the form of a candle(OHLC format) of any timeframe specified. The timeframe
will also be specified in the candle object.

The service currently listens to Redis pub/sub messages on the `control` channel; control messages
are decoded into specific requests (e.g., `data_request`, `ingest_request`) and routed to handlers in
`market/internal/ingest`. Shared configuration such as `REDIS_HOST`/`REDIS_PORT` is consumed via the
`internal/config` package so other services can reuse the same helpers.

Market data is stored in a postgres database for historical use cases.

If the market service is tasked with providing historical data for backtesting services, it will
gather historical data based on the requested timeframe and send a sequential Market events to the
message broker. A predefined backtesting session id will be passed to ensure the appropriate backtesting
session picks up those events.


## Technology
Backend Services: Go
Message Broker: Rabbitmq
Frontend: Nextjs
