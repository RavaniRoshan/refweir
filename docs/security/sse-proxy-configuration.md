# Server-Sent Events (SSE) Reverse Proxy Configuration

Refweir streams real-time autonomous AI agent activity and telemetry over HTTP Server-Sent Events at `GET /api/v1/agent/activities/stream`.

To ensure uninterrupted streaming through reverse proxies, caching tiers, and API gateways, configure the following parameters.

## Nginx Configuration

Nginx buffers response bodies by default. To disable response buffering and enable long-lived streaming connections:

```nginx
location /api/v1/agent/activities/stream {
    proxy_pass http://127.0.0.1:8080;
    
    # Disable buffering for real-time SSE streaming
    proxy_buffering off;
    proxy_cache off;
    proxy_set_header X-Accel-Buffering no;
    
    # Keep-alive connection parameters
    proxy_http_version 1.1;
    proxy_set_header Connection "";
    
    # Extended timeouts for long-running idle streams (refweir sends heartbeats every 15s)
    proxy_read_timeout 3600s;
    proxy_send_timeout 3600s;
    
    # Standard forwarding headers
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

## Cloudflare CDN & Edge Proxy

1. **Disable Response Buffering:**
   Ensure `X-Accel-Buffering: no` is preserved (refweir automatically includes this header on all SSE streams).
2. **Proxy Read Timeout:**
   Cloudflare terminates connections that have no traffic for 100 seconds on standard plans. Refweir emits an `event: heartbeat` every 15 seconds to keep edge connections open indefinitely.
3. **Chunked Encoding:**
   Do not modify the `Transfer-Encoding: chunked` header; streaming responses require chunked transfer.

## Caddy Configuration

```caddy
reverse_proxy /api/v1/agent/activities/stream localhost:8080 {
    flush_interval -1
}
```

Setting `flush_interval -1` causes Caddy to immediately flush every event downstream without buffering.
