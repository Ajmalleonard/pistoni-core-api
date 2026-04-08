#!/usr/bin/env python3
import json, os, urllib.request, urllib.parse

TOKEN = os.getenv('CLOUDE_FLARE_KEYS','').strip()
ZONE_NAME = os.getenv('CF_ZONE_NAME','pistoni.io').strip()
RECORD_NAME = os.getenv('CF_RECORD_NAME','core.api').strip()
IPV4 = os.getenv('CF_IPV4','').strip()
IPV6 = os.getenv('CF_IPV6','').strip()
BASE='https://api.cloudflare.com/client/v4'

if not TOKEN:
    raise SystemExit('Missing CLOUDE_FLARE_KEYS')
if not IPV4 and not IPV6:
    raise SystemExit('Set CF_IPV4 and/or CF_IPV6')

headers={
    'Authorization': f'Bearer {TOKEN}',
    'Content-Type': 'application/json'
}

def req(method, path, payload=None):
    data=None if payload is None else json.dumps(payload).encode()
    r=urllib.request.Request(BASE+path, data=data, method=method, headers=headers)
    with urllib.request.urlopen(r, timeout=30) as resp:
        return json.loads(resp.read().decode())

def zone_id():
    q=urllib.parse.quote(ZONE_NAME)
    out=req('GET', f'/zones?name={q}&status=active')
    res=out.get('result',[])
    if not res: raise SystemExit(f'Zone not found: {ZONE_NAME}')
    return res[0]['id']

def upsert_record(zid, typ, name, content):
    qname=urllib.parse.quote(name)
    out=req('GET', f'/zones/{zid}/dns_records?type={typ}&name={qname}')
    records=out.get('result',[])
    payload={'type':typ,'name':name,'content':content,'ttl':120,'proxied':True}
    if records:
        rid=records[0]['id']
        out=req('PUT', f'/zones/{zid}/dns_records/{rid}', payload)
        action='updated'
    else:
        out=req('POST', f'/zones/{zid}/dns_records', payload)
        action='created'
    if not out.get('success'):
        raise SystemExit(f'Failed {typ}: {out}')
    print(f'{typ} {action}: {name} -> {content}')

zid=zone_id()
fqdn=f'{RECORD_NAME}.{ZONE_NAME}' if not RECORD_NAME.endswith(ZONE_NAME) else RECORD_NAME
if IPV4: upsert_record(zid,'A',fqdn,IPV4)
if IPV6: upsert_record(zid,'AAAA',fqdn,IPV6)
print('done')
