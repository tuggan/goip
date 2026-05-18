# goip

A small self-hosted web server that shows the connecting client info.

## Quick start

```sh
make run     # builds and starts the server on http://0.0.0.0:3000
make test    # runs all tests
```

## Raw text endpoints

Each of the headers can be accessed as its own enpoint that only returns the value.

## Configuration

GoIP can be configured with a TOML file or command-line flags.
Use `-c <path>` to point to a config directory. By default it looks for
`goip.toml` in the current directory, `$HOME/.goip/`, and `/etc/goip/`.

| Flag               | Default        | Description                                               |
| ------------------ | -------------- | --------------------------------------------------------- |
| `-e`, `--endpoint` | `0.0.0.0:3000` | Address(es) to listen on                                  |
| `--tlsEndpoint`    | —              | Address(es) for HTTPS (requires `--tlsCert` + `--tlsKey`) |
| `--tlsKey`         | —              | Paths to TLS private key                                  |
| `--tlsCert`        | —              | Paths to TLS certificate                                  |
| `--trustedProxy`   | —              | Trusted proxy IP or CIDR range (repeatable)               |
| `--rateLimit`      | `10`           | Maximum requests per second per IP (`0` disables)         |
| `-c`, `--config`   | `.`            | Path to config directory                                  |

## Docker

```sh
make dockerimage
docker run -p 3000:3000 tuggan/goip
```

## Example

![Example of headers from request](example.png)

## Author

Dennis Vesterlund \<dennis@vestern.se\>
