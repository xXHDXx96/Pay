# alipay-payment

A Go payment gateway for enterprise merchants: **one QR code, many payment rails**
(Alipay, WeChat, bank card, and one-click crypto redemption), with a real-time
generated QR and an API surface for everything.

## Features

- **One code, many payments** — a single order can be paid over any configured
  rail; the merchant renders one QR and the customer chooses.
- **Real-time QR generation** — `POST /pay` returns a `qr_code_url` immediately;
  for Alipay the QR payload comes from `alipay.trade.precreate`.
- **Crypto redemption** — `POST /exchange/quote` prices crypto at live rates,
  `POST /exchange/redeem` broadcasts the on-chain transfer in one click.
- **Webhooks** — `POST /notify/alipay` verifies the platform signature (RSA2)
  before applying the trade state to the order.
- **Clean separation** — `store.Store` interface lets you swap the in-memory
  dev backend for Postgres without touching handlers.

## Layout

```
cmd/server/main.go            # entrypoint + routing
internal/config/config.go     # env-driven config (Vault/KMS in prod)
internal/store/store.go       # persistence interface (+ memory impl)
internal/model/order.go       # order + crypto redemption models
internal/alipay/              # RSA2 signing, precreate, query, notify verify
internal/handler/             # HTTP handlers
internal/exchange/            # rate oracle + on-chain broadcaster
```

## Run

```bash
cp .env .env.local && edit .env.local
go run ./cmd/server
```

Endpoints:

| Method | Path               | Description                              |
|--------|--------------------|------------------------------------------|
| POST   | `/pay`             | Create order, return QR code             |
| GET    | `/orders`          | List orders (`?merchant_id=`)            |
| GET    | `/orders/{id}`     | Fetch a single order                     |
| POST   | `/notify/alipay`   | Alipay webhook (signature-verified)      |
| POST   | `/exchange/quote`  | Get a crypto quote                       |
| POST   | `/exchange/redeem` | Broadcast on-chain transfer (one click) |
| GET    | `/exchange`        | Fetch redemption (`?order_id=`)          |

## Crypto flow

1. `POST /exchange/quote` → returns `crypto_amount = fiat_cny / exchange_rate`
2. `POST /exchange/redeem` with the customer's wallet → broadcasts the
   stablecoin transfer and returns the `tx_hash` with `status: CONFIRMED`.

The `Broadcaster` func is injected at startup; production wires `ethclient`
plus the USDT ABI. Set `ETH_RPC` and `CRYPTO_MERCHANT_WALLET` to enable it.

## Security notes

- Alipay private keys must never be committed; source from Vault/KMS.
- Webhook secret is verified via HMAC over the form body.
- This is a reference implementation; do not run unmodified in production
  without review.

## Tests

```bash
go test ./internal/...
go vet ./...
```