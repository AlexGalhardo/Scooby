# OpenBB: patterns for vukem-trading

Source: [OpenBB-finance/OpenBB](https://github.com/OpenBB-finance/OpenBB) (Apache-2.0) and [docs.openbb.co](https://docs.openbb.co/). Researched 2026-11-04. The repository named in the brief, `openbq-org/OpenBB`, resolved to a page whose links point to the official owner `OpenBB-finance`; it is treated as a mirror and `OpenBB-finance` as canonical. Several docs pages returned 404, so the provider list and caching behavior are *unverified*.

## What it is

The Open Data Platform (ODP): a Python-first data layer for analysts, quants and AI agents (`pip install openbb`). OpenBB Workspace is a separate hosted enterprise frontend.

## Provider contract (worth copying as a Go interface)

- A `Fetcher` has three stages: `transform_query` (normalize parameters), `extract_data` (call the source) and `transform_data` (map to the standard model). An alias map renames source fields to standard field names.
- A `Provider` declares a name, a description, a website, a map from model name to fetcher, and the credentials it needs.
- Credentials are named per provider (`<provider>_api_key`) and read from a user object or from an environment variable. A missing required credential fails before the fetcher runs.
- Providers register through package entry points, so the core never imports them.
- Caching is not documented on the provider page (*unverified*).

## Serving and MCP

- `openbb-platform-api` serves the platform as a FastAPI app on loopback for the Workspace.
- `openbb-mcp-server` turns REST routes into MCP tools (default `http://127.0.0.1:8001/mcp`; streamable HTTP, SSE or stdio). Tool names derive from the route path. Inbound and outbound auth use separate environment variables. A tool-discovery mode and a category allowlist keep the tool list small.
- The Workspace MCP bridge needs an open browser tab and a personal access token, one bridge per user. Connected MCP clients are treated as trusted.

## Coverage

Data-extension tables were not reachable. From search snippets only (*unverified*): yfinance needs no key; FMP, FRED and Polygon have free tiers; Benzinga and TradingEconomics are paid. No B3 provider was confirmed.

## Recommendations for vukem-trading

1. Define a Go `Provider` interface with one standard struct per data type (quote, bar, fundamentals, macro) and a three-stage fetch pipeline.
2. Name credentials per provider, support an environment fallback, and fail fast before any network call.
3. Expose read-only data through an MCP server with a bearer token and a tool allowlist. Do not expose order tools to agents by default.
4. Keep the core free of provider imports; register adapters at startup.
