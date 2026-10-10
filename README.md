# go-optel

A dependency-free Go client for Adobe's [helix-rum-js](https://github.com/adobe/helix-rum-js) wire format, the operational telemetry (OpTel) beacons that SLICC's web pages and native apps send. It is used by [slicc-cli](https://github.com/ai-ecoverse/slicc-cli). OpTel is Adobe's RUM-style operational telemetry, not OpenTelemetry.

```sh
go get github.com/ai-ecoverse/go-optel
```

```go
import optel "github.com/ai-ecoverse/go-optel"

client := optel.Configure("slicc-cli", optel.Options{})
defer client.Flush(2 * time.Second)

client.Sample(optel.Enter, "follow", "")

if err != nil {
	client.ReportError("dial", err)
}
```

## Checkpoints

A headless CLI has no clicks, views or Core Web Vitals, so it sends two checkpoints:

- `Enter` when the process starts, with the subcommand as `source`.
- `Error` for an operational failure, always through `ReportError`.

The first beacon of a process is preceded by `top`, as in the other clients. `Checkpoint` stays an open string, so other checkpoints are possible, but a CLI should rarely need them.

## Privacy

A CLI's error strings carry secrets by construction. Go's `net/http` and `net/url` errors include the whole request URL, with any token in its path or query, and file errors include the home directory and so the login name. `ReportError` runs every message through `Sanitize`:

- Every absolute URL becomes `scheme://host/...`.
- Absolute POSIX and Windows paths shrink to their first segment or drive.
- Only then is the message cut to 200 characters, so a cut never leaves part of a secret behind.

Never call `client.Sample(optel.Error, source, err.Error())` directly.

## Sampling

One decision per `Configure`, that is per process, not per beacon, as the web and Swift clients decide once per page view or session. The beacon goes to `<collect base>/.rum/<weight>`, by default `https://rum.hlx.page/.rum/100`.

| variable | values |
| --- | --- |
| `OPTEL_RATE` | `on` (weight 1, every process), `high` (10), `low` (1000), `off`; anything else is weight 100 |
| `OPTEL_DEBUG` | log each beacon to stderr |

Callers add their own opt-out on top; slicc-cli honors `SLICC_NO_TELEMETRY=1` and never configures a client for a development build.

## Development

`npm run lint` runs the slicc lint tools, `gofmt`, `go mod tidy -diff`, `go vet` and golangci-lint. `npm test` runs the integration tests in `test/integration/` against a local collector. Unit tests stay next to the code and out of git (`npm run test:unit`). Releases are GitHub tags cut by semantic-release on `main`; the Go module proxy serves them.
