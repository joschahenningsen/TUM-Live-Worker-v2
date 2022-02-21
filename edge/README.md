# TUM-Live-Worker-v2/edge

The edge module is designed as a simple edge proxy and cache node for TUM-Live-Worker-v2.
It can be used when network traffic to worker nodes exceeds the available bandwidth, the architecture might look like this:
```
                               ┌───────┐
┌─────────────────┐ ──────────►│Edge 1 ├──────┐
│                 │            └───────┘      └──►┌───────────┐
│  Load Balancer  │                               │  Worker n │
│(DNS-RR/HTTP 302)│                               │           │
│                 │            ┌───────┐      ┌──►└───────────┘
└─────────────────┘ ──────────►│Edge 2 ├──────┘
                               └───────┘
```

## Configuration

The following configuration options are available via environment variables:

- `PORT`: The port on which the edge node should listen for incoming connections (default: 8080).
- `ORIGIN_PORT`: The port on which the workers hls files are available (default: 8085). 
- `ORIGIN_PROTO`: The protocol of the origin server (default: http).
