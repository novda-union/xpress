# Xpressgo Local HTTPS Telegram Mini App Design

## Purpose

Define a local development and demo setup that lets the customer-facing web app run inside a Telegram Mini App over trusted HTTPS without ngrok or similar tunneling services.

This design is for prototype and local demo use. It prioritizes reliable same-origin HTTPS behavior over production-grade deployment concerns.

## Goals

- run the customer-facing web app inside Telegram over HTTPS
- keep the phone number verification flow working inside the Mini App
- avoid mixed-content issues between the web app and the API
- avoid dependence on ngrok or similar external tunneling services
- keep the setup portable across different Wi-Fi networks with only a small manual device step
- keep `make fresh` responsible for the laptop-side automation

## Core Constraints

- Telegram Mini Apps require trusted HTTPS
- if the frontend loads from HTTPS, API requests and WebSocket requests must also resolve over HTTPS
- the client-facing web app and API should share one HTTPS origin
- internal service ports can remain unchanged
- the prototype does not need strict CORS restrictions
- the setup must work on both Android and iPhone, assuming the local CA is trusted on the device

## Recommended Architecture

Use one stable hostname:

- `xpressgo.home.arpa`

Use one HTTPS front door:

- `https://xpressgo.home.arpa`

Use a local reverse proxy listening on:

- port `443`

Behind the reverse proxy:

- web app dev server stays on `5173`
- Go API server stays on `8080`
- admin stays on `3000` and is out of scope for the Mini App path

## Same-Origin Routing Model

The reverse proxy terminates HTTPS and forwards requests by path.

### Frontend

Route these requests to the web app:

- `/`
- static assets
- client-side routes such as `/branch/:id`, `/cart`, `/orders`

### API

Route these requests to the Go server:

- `/auth/*`
- `/discover`
- `/branches/*`
- `/stores/*`
- `/orders`
- `/orders/*`
- `/ws`

### WebSocket

WebSocket traffic must also terminate on the same HTTPS origin:

- `wss://xpressgo.home.arpa/ws`

This avoids mixed-content issues and removes the need for prototype CORS juggling in the browser.

## Certificates And Trust

Use one persistent local CA and one persistent certificate for:

- `xpressgo.home.arpa`

Recommended behavior:

- create the CA and server certificate only if missing
- do not regenerate them on every `make fresh`
- keep the CA stable so the phone only has to trust it once

The certificate must be trusted by:

- the laptop running the project
- the Android test device
- the iPhone test device

Telegram’s in-app browser will behave according to the device trust store. If the CA is trusted by the device, the Mini App can load the local HTTPS origin.

## Network Change Model

The hostname stays stable:

- `xpressgo.home.arpa`

What changes when you move to a new Wi-Fi network:

- the laptop LAN IP

What remains stable:

- hostname
- certificate
- trusted CA
- Telegram bot `APP_URL`

### Manual Step Per Wi-Fi

When joining a new Wi-Fi network, manually update the device-side hostname mapping so:

- `xpressgo.home.arpa` points to the laptop’s current LAN IP

This is the only required phone-side step after the initial certificate trust setup.

## `make fresh` Responsibilities

`make fresh` should automate the laptop-side runtime setup as much as possible.

Recommended responsibilities:

1. rebuild and start the normal stack
2. ensure the local CA and certificate exist
3. start the HTTPS reverse proxy on `443`
4. start or expose the web app behind that proxy
5. configure runtime environment so the Telegram bot and frontend use `https://xpressgo.home.arpa`
6. print the current laptop LAN IP and the manual phone remap instruction

Recommended terminal output after startup:

```text
Local HTTPS Mini App URL: https://xpressgo.home.arpa
Current LAN IP: 192.168.x.x
Phone step: map xpressgo.home.arpa -> 192.168.x.x on your test device
```

## Configuration Changes

### Server

Server runtime should:

- stop relying on `http://localhost:5173` as the public app URL
- use `https://xpressgo.home.arpa` for Telegram Mini App links
- relax or remove prototype CORS restrictions for now

### Web

Web runtime should:

- use same-origin API requests in local HTTPS mode
- avoid hardcoding `http://localhost:8080` in the customer-facing path
- allow WebSocket connections against the same origin

### Reverse Proxy

The reverse proxy should:

- listen on `443`
- terminate TLS with the local certificate
- proxy frontend routes to the web app
- proxy API and WebSocket routes to the Go server

## Why This Design

This is preferred over a raw-IP HTTPS setup because:

- hostnames are more reliable for certificates than bare IPs
- the hostname stays stable across Wi-Fi changes
- Telegram and mobile webviews are less likely to hit TLS edge cases
- a same-origin HTTPS setup removes mixed-content and most CORS problems

This is preferred over network-wide DNS because:

- it avoids changing the whole Wi-Fi network
- it only requires a device-level hostname remap on the phone
- it is safer for client-office demos

## Manual Setup Expectations

### One-Time Setup

- trust the local CA on the phone
- trust the local CA on the laptop if needed

### Per-Network Setup

- determine the laptop LAN IP
- update the phone’s hostname mapping for `xpressgo.home.arpa`

This is acceptable for the prototype because the CA trust does not need to be repeated every time.

## Verification Targets

After implementation, verify:

- Telegram bot opens the Mini App at `https://xpressgo.home.arpa`
- the web app loads without certificate warnings on trusted devices
- `POST /auth/telegram` works through the HTTPS origin
- phone/contact verification works inside Telegram
- discovery, branch menu browsing, cart, and order creation work through the same HTTPS origin
- WebSocket order updates work through `wss://xpressgo.home.arpa/ws`

## Out Of Scope

- production-grade public deployment
- public DNS automation
- zero-touch phone-side reconfiguration across arbitrary Wi-Fi networks
- network-wide DNS changes for everyone on the current Wi-Fi
- admin panel HTTPS delivery in the same iteration unless needed later
