# putio-exporter

Prometheus Exporter for `put.io` account details.

This Prometheus Exporter will export metrics gathered about a user's `put.io` account, including disk space used/available.
This is useful for monitoring disk availability over time.

## 📚 Background

Put.io is a cloud-based storage service that allows users to store and access their files from anywhere with an internet connection. This project is NOT affiliated with Put.io.

## Metrics Gathered

| Metric name                          | Type                  | Description                                                          |
| ------------------------------------ | --------------------- | -------------------------------------------------------------------- |
| `putio_up`                           | Gauge (1- yes, 0- no) | Was the last metrics-gathering query to putio successful.            |
| `putio_account_active`               | Gauge (1- yes, 0- no) | Is the putio account being queried currently active.                 |
| `putio_account_plan_expiration_date` | Gauge (timestamp)     | The unix time when the currently active putio plan expires.          |
| `putio_disk_available`               | Gauge (bytes)         | The available disk space in the account (in bytes).                  |
| `putio_disk_size`                    | Gauge (bytes)         | The total disk space, available and used, in the account (in bytes). |
| `putio_simultaneous_download_limit`  | Gauge (count)         | The maximum amount of downloads permitted by the account tier.       |
| `putio_transfers_by_status`          | Gauge (count)         | The number of transfers by status type.                              |

> ❕ The `putio_transfers_by_status` metric has a label (`"type"`), which can be of status `ERROR`, `COMPLETED`, or `DOWNLOADING`.

## Running Locally

### 🪄 `docker-compose` (recommended)

1. In the docker-compose.yaml file, set `PUTIO_OAUTH_TOKEN` to your [put.io OAUTH token](https://help.put.io/en/articles/5972538-how-to-get-an-oauth-token-from-put-io).
1. Run `docker-compose up`.

### 🚢 Docker

Steps:

```bash
$ docker build putio-exporter:latest # Build the docker image
$ docker run --env PUTIO_OAUTH_TOKEN=your-putio-oauth-token -p 9101:9101 putio-exporter:latest # Run the docker image
```

### 🔧 Manual Build

Requirements:

- `go` (Created using `go version go1.20.2 darwin/amd64`)

Steps:

```bash
$ go build # Build the executable
$ ./putio-exporter # Run the executable
```

## ⚙️ Configuration Options

| Variable                      | Command Line Flag | Required                      | Description                                                                                                                        |
| ----------------------------- | ----------------- | ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `PUTIO_OAUTH_TOKEN`           | `--oauth-token`   | Required                      | The [put.io OAUTH token](https://help.put.io/en/articles/5972538-how-to-get-an-oauth-token-from-put-io) created from your account. |
| `PUTIO_EXPORTER_LISTEN_PORT`  | `--listen-port`   | Optional (default `9101`)     | The port at which the exporter listens for requests.                                                                               |
| `PUTIO_EXPORTER_METRICS_PATH` | `--metrics-path`  | Optional (default `/metrics`) | The path at which the exporter serves the collected metrics.                                                                       |

There are three ways to pass in configuration options:

- as environment variables (recommended)
- in an `.env` file
- as command-line arguments when using the executable directly
  - `./putio-exporter --metrics-path=9303 ...`

## 🚀 Contributing

All contributions are welcome. Please create a PR with a description of the proposed changes.

This software is distributed with the Apache 2.0 license.
