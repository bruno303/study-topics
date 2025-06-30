# Let's Encrypt HTTPS

* Local HTTPS with mkcert (development).
* Certbot for Let’s Encrypt certs (Cloudflare + self-hosted).
* Port forwarding from Windows host → WSL.

## 1. Local HTTPS with mkcert (Development)

### Generate and trust certs

```bash
mkcert -install
mkcert example.local
```

Usage in Node.js
```javascript
const https = require('https');
const fs = require('fs');
const app = require('./app');

https.createServer({
  key: fs.readFileSync('example.local-key.pem'),
  cert: fs.readFileSync('example.local.pem'),
}, app).listen(443);
```

## 2. Certbot for Let’s Encrypt

### Cloudflare (Wildcard Certificates)

```bash
# Install Certbot + Cloudflare plugin
sudo apt install certbot python3-certbot-dns-cloudflare

# Configure Cloudflare API token
echo "dns_cloudflare_api_token = YOUR_API_TOKEN" > ~/cloudflare.ini
chmod 600 ~/cloudflare.ini

# Generate cert
sudo certbot certonly \
  --dns-cloudflare \
  --dns-cloudflare-credentials ~/cloudflare.ini \
  -d "*.example.com" \
  --preferred-challenges dns
```

### Self-Hosted (HTTP Challenge)

First, make sure that the dns update script is in a directory accessible and that will not be changed. It's a good practice to keep it on `/etc/letsencrypt/scripts`.

```bash
# Stop services using port 80
sudo systemctl stop nginx

# Generate cert
sudo certbot certonly \
  --manual \
  --preferred-challenges dns \
  --manual-auth-hook /etc/letsencrypt/scripts/duckdns-update.sh \
  --manual-cleanup-hook /etc/letsencrypt/scripts/duckdns-update.sh \
  -d "*.example.com" \
  --server https://acme-v02.api.letsencrypt.org/directory \
  --agree-tos \
  --email email@email
```

### Auto-Renewal
```bash
sudo crontab -e
```

```bash
0 3 * * * certbot renew --quiet --post-hook "systemctl restart nginx"
```

## 3. Port Forwarding (Windows Host → WSL)

### Forwarding (PowerShell as Admin)

```powershell
# Get the first IP from WSL (trim whitespace and split by space)
$wsl_ip = (wsl hostname -I).Trim() -split ' ' | Select-Object -First 1

# Check if IP is valid
if (-not ($wsl_ip -match '^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$')) {
    Write-Host "❌ Invalid WSL IP detected: $wsl_ip"
    exit 1
}

netsh interface portproxy add v4tov4 listenport=443 listenaddress=0.0.0.0 connectport=443 connectaddress=$wsl_ip
netsh interface portproxy add v4tov4 listenport=80 listenaddress=0.0.0.0 connectport=80 connectaddress=$wsl_ip

New-NetFirewallRule -DisplayName "WSL HTTP" -Direction Inbound -Action Allow -Protocol TCP -LocalPort 80 -ErrorAction SilentlyContinue
New-NetFirewallRule -DisplayName "WSL HTTPS" -Direction Inbound -Action Allow -Protocol TCP -LocalPort 443 -ErrorAction SilentlyContinue

Write-Host "✅ WSL static IP set to $wsl_ip. Ports 80/443 forwarded."
```

## 4. Nginx Config for HTTPS
```nginx
server {
    listen 443 ssl;
    server_name example.com;

    ssl_certificate /etc/letsencrypt/live/example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/example.com/privkey.pem;

    location / {
        proxy_pass http://localhost:3000;  # Forward to Node.js
    }
}
```
Restart Nginx:

```bash
sudo systemctl restart nginx
```
## Troubleshooting

### Certbot fails? Use --dry-run to test:

```bash
sudo certbot certonly --dry-run ...
```
