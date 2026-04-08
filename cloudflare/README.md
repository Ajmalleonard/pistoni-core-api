# Cloudflare DNS Automation (Pistoni Core)

This folder contains scripts/config to manage DNS records for Pistoni Core API.

## Target
- Domain: `pistoni.io`
- Subdomain: `core.api.pistoni.io`

## Usage
Set environment variables first:

```bash
export CLOUDE_FLARE_KEYS="<api-token>"
export CF_ZONE_NAME="pistoni.io"
export CF_RECORD_NAME="core.api"
export CF_IPV4="<server-ipv4>"
export CF_IPV6="<server-ipv6>"
```

Then run:

```bash
python3 cloudflare/upsert_dns.py
```
