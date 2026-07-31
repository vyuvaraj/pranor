# Wildcard DNS & Reverse Proxy Setup Guide

This guide explains how to set up wildcard DNS routing and a production reverse proxy (Nginx or Caddy) with TLS termination for `Pranor Tunnel`.

## 1. DNS Configuration

To route incoming subdomains (e.g., `app1.pranor.net`, `test-env.pranor.net`) to the Pranor Tunnel relay server, you must configure **wildcard DNS records** at your domain registrar/DNS provider (Cloudflare, AWS Route53, Namecheap, etc.).

### Recommended DNS Records

| Record Type | Host / Name | Value | TTL | Description |
|:---|:---|:---|:---|:---|
| **A** | `@` (or empty) | `YOUR_SERVER_IP` | Automatic / 300s | Routes the root domain `pranor.net` to the relay. |
| **A** | `*` | `YOUR_SERVER_IP` | Automatic / 300s | Wildcard record routing all subdomains to the relay. |

*Note: Replace `YOUR_SERVER_IP` with the public IPv4 address of your Pranor Tunnel relay server.*

---

## 2. Reverse Proxy Configuration with TLS

It is best practice to run the `pranor-tunnel server` behind a reverse proxy (like **Nginx** or **Caddy**) for TLS termination, security, and WebSocket upgrade handling.

### Option A: Nginx Configuration

Here is a configuration sample for Nginx supporting HTTP/2, WebSockets, and SSL termination.

```nginx
# HTTP - Redirect all traffic to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name pranor.net *.pranor.net;
    return 301 https://$host$request_uri;
}

# HTTPS - Proxy requests to Pranor Tunnel Relay
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name pranor.net *.pranor.net;

    # SSL Certificates (obtained via Certbot / Let's Encrypt Wildcard)
    ssl_certificate /etc/letsencrypt/live/pranor.net/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/pranor.net/privkey.pem;

    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;

    location / {
        proxy_pass http://127.0.0.1:8443;
        
        # Preserve original request headers
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # WebSocket support (Crucial for Pranor Tunnel client connections)
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";

        # Disable buffering for real-time proxying
        proxy_buffering off;
        proxy_read_timeout 3600s;
        proxy_send_timeout 3600s;
    }
}
```

### Option B: Caddy Configuration

Caddy automatically provisions and renews wildcard certificates using ACME (DNS-01 challenge).

```caddy
*.pranor.net, pranor.net {
    # Proxy all traffic to the local Pranor Tunnel relay server
    reverse_proxy 127.0.0.1:8443 {
        # Keep connection open for WebSocket streaming
        header_up Host {host}
        header_up X-Real-IP {remote}
    }
    
    # Enable wildcard SSL automatically (requires DNS provider credentials API)
    tls {
        dns cloudflare {env.CLOUDFLARE_API_TOKEN}
    }
}
```

---

## 3. Running the Server

Start the Pranor Tunnel relay server listening on the local loopback port matching your reverse proxy configuration:

```bash
pranor-tunnel server --port 8443 --domain pranor.net
```
