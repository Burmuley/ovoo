# Single-Host Mail Setup: Ovoo + Postfix + OpenDKIM

This guide walks through configuring a single Linux host (Debian/Ubuntu) that runs all Ovoo services alongside Postfix (in a two-instance setup) and OpenDKIM. All examples use the placeholder domain `ovoodomain.example` and hostname `mx.ovoodomain.example`; substitute your real values everywhere you see them.

The configuration files referenced throughout this guide are provided as ready-to-use examples under [etc](etc/) next to this document.

---

## Table of Contents

1. [Architecture overview](#1-architecture-overview)
2. [Prerequisites](#2-prerequisites)
3. [System user and directories](#3-system-user-and-directories)
4. [Ovoo binary and configuration](#4-ovoo-binary-and-configuration)
5. [Systemd service units](#5-systemd-service-units)
6. [Postfix multi-instance setup](#6-postfix-multi-instance-setup)
7. [OpenDKIM setup](#7-opendkim-setup)
8. [Verification](#8-verification)

---

## 1. Architecture overview

Ovoo uses a **Postfix two-instance** pattern: one Postfix process handles inbound mail from the Internet, passes it through the Ovoo milter for alias rewriting, then hands it to a second Postfix process that applies DKIM signing and delivers it externally.

<p align="center">
    <img width="60%" src="../assets/overview/ovoo_overview.png" alt="Ovoo overview diagram" />
</p>


### Component roles

| Component | Listen address | Role |
|---|---|---|
| **postfix-in** | `0.0.0.0:25` | Accepts inbound SMTP from the Internet. Enforces SPF (policyd-spf), verifies DKIM (OpenDKIM), rewrites alias headers (Ovoo milter), then forwards to postfix-out. |
| **postfix-out** | `127.0.0.1:10026` | Loopback-only re-injection listener. Receives mail from postfix-in, signs outbound messages with DKIM (OpenDKIM), and delivers to external MX servers. |
| **Ovoo milter** | `127.0.0.1:6785` | Sendmail milter: intercepts messages, looks up alias/chain records via the Ovoo API, rewrites envelope and headers so aliases forward to protected addresses without exposing them. |
| **Ovoo socketmap** | `127.0.0.1:7788` | Answers Postfix `socketmap` queries for `relay_domains`; returns the set of alias domains Ovoo currently manages so Postfix knows which domains to accept mail for. |
| **Ovoo API** | `0.0.0.0:8808` | REST API and embedded Vue.js WebUI for managing users, aliases, protected addresses, and API tokens. Used internally by the milter and socketmap services. |
| **OpenDKIM** | `127.0.0.1:8891` | Signs outbound mail (postfix-out) and verifies inbound DKIM signatures (postfix-in). Uses Lua-based key and signing tables for flexible multi-domain support. |


## 2. Prerequisites

### Packages

```bash
apt update
apt install postfix postfix-policyd-spf-python opendkim sqlite3
```

> **Note:** `sqlite3` is the default Ovoo database driver. If you plan to use MySQL, install the MySQL client libraries instead and adjust `config.json` accordingly.

### DNS records

Configure these records for your domain before starting the services. The DKIM TXT record value is generated in [step 7](#7-opendkim-setup).

| Type | Name | Value |
|---|---|---|
| A (or AAAA) | `mx.ovoodomain.example` | Server public IP address |
| MX | `ovoodomain.example` | `mx.ovoodomain.example` (priority 10) |
| TXT | `ovoodomain.example` | `v=spf1 mx ~all` |
| TXT | `<selector>._domainkey.ovoodomain.example` | DKIM public key (see step 7) |


## 3. System user and directories

Create the `ovoo` system user and the directories it needs:

```bash
# Create system group and user
groupadd --system ovoo
useradd --system --gid ovoo --shell /usr/sbin/nologin \
        --home /var/lib/ovoo --no-create-home ovoo

# Config directory (readable by ovoo group)
install -d -o root -g ovoo -m 755 /usr/local/etc/ovoo

# Data directory (writable by ovoo — holds the SQLite database)
install -d -o ovoo -g ovoo -m 775 /var/lib/ovoo
```

---

## 4. Ovoo binary and configuration

### Install the binary

Download the latest release binary for your platform from the [GitHub releases page](https://github.com/Burmuley/ovoo/releases) and install it:

```bash
# Example: extract from a release tarball
tar xzf ovoo_Linux_x86_64.tar.gz
install -o root -g root -m 755 ovoo /usr/local/bin/ovoo
```

### Configuration file

Create `/usr/local/etc/ovoo/config.json`. The three Ovoo services (API, milter, socketmap) all read this single file.

```bash
# File must be readable by the ovoo group
touch /usr/local/etc/ovoo/config.json
chown root:ovoo /usr/local/etc/ovoo/config.json
chmod 640 /usr/local/etc/ovoo/config.json
```

**Example `config.json`:**

```json
{
  "api": {
    "listen_addr": "0.0.0.0:8808",
    "tls": {
      "cert": "/usr/local/etc/ovoo/tls/server.crt",
      "key":  "/usr/local/etc/ovoo/tls/server.key"
    },
    "database": {
      "driver": "gorm",
      "config": {
        "gorm": {
          "driver": "sqlite",
          "connection_string": "/var/lib/ovoo/ovoo.db"
        }
      }
    },
    "sysinfo": {
      "dkim_domain":   "ovoodomain.example",
      "dkim_selector": "<selector>"
    },
    "default_admin": {
      "login":     "admin@ovoodomain.example",
      "firstName": "Admin",
      "lastName":  "User",
      "password":  "<strong-password>"
    },
    "log": {
      "level":       "info",
      "destination": "stdout"
    },
    "oidc": {
      "google": {
        "client_id": "<google-client-id>.apps.googleusercontent.com",
        "client_secret": "<google-client-secret>",
        "issuer": "https://accounts.google.com",
        "extra_url_params": {
          "access_type": "offline",
          "prompt": "consent"
        }
      }
    }
  },
  "milter": {
    "listen_addr": "127.0.0.1:6785",
    "api": {
      "addr":            "https://127.0.0.1:8808",
      "tls_skip_verify": true,
      "auth_token":      "<api-token>"
    },
    "log": {
      "level":       "info",
      "destination": "stdout"
    }
  },
  "socketmap": {
    "network":     "tcp4",
    "listen_addr": "127.0.0.1:7788",
    "api": {
      "addr":            "https://127.0.0.1:8808",
      "tls_skip_verify": true,
      "auth_token":      "<api-token>"
    },
    "log": {
      "level":       "info",
      "destination": "stdout"
    }
  }
}
```

**Key fields:**

| Field | Description |
|---|---|
| `api.listen_addr` | Address and port the REST API and WebUI listen on. |
| `api.tls.cert` / `key` | TLS certificate and key for the API. Use a certificate from Let's Encrypt or your CA. |
| `api.database` | SQLite is the default. Set `connection_string` to the database file path. |
| `api.sysinfo.dkim_domain` | The domain that appears in DKIM signatures. Should match your alias domain. |
| `api.sysinfo.dkim_selector` | DKIM selector (the label before `._domainkey.` in DNS). |
| `api.default_admin` | Bootstrapped admin account created on first startup. Change the password immediately after first login. |
| `milter.listen_addr` | The TCP address the Ovoo milter listens on. Must match `smtpd_milters` in postfix-in `main.cf`. |
| `milter.api.auth_token` | API token the milter uses to authenticate with the Ovoo API. Create it via the WebUI or API after first boot. |
| `socketmap.listen_addr` | The TCP address the socketmap service listens on. Must match the `relay_domains` socketmap address in postfix-in `main.cf`. |
| `socketmap.api.auth_token` | API token the socketmap uses to authenticate. Can be the same token as the milter. |

> **TLS note:** The milter and socketmap connect to the API over TLS. If you use a self-signed certificate, set `tls_skip_verify: true`. In production with a valid CA-signed certificate, remove that field.

---

## 5. Systemd service units

Copy the provided unit files to the systemd directory and enable all three services:

```bash
cp etc/systemd/system/ovoo-api.service       /etc/systemd/system/
cp etc/systemd/system/ovoo-milter.service    /etc/systemd/system/
cp etc/systemd/system/ovoo-socketmap.service /etc/systemd/system/

systemctl daemon-reload
systemctl enable --now ovoo-api ovoo-milter ovoo-socketmap
```

### `ovoo-api.service`

```ini
[Unit]
Description=Ovoo Mail Aliasing API Service

[Service]
User=ovoo
Group=ovoo
ExecStart=/usr/local/bin/ovoo api -config /usr/local/etc/ovoo/config.json

[Install]
WantedBy=multi-user.target
```

### `ovoo-milter.service`

```ini
[Unit]
Description=Ovoo Mail Aliasing Milter Service

[Service]
User=ovoo
Group=ovoo
ExecStart=/usr/local/bin/ovoo milter -config /usr/local/etc/ovoo/config.json

[Install]
WantedBy=multi-user.target
Wants=ovoo-api.service
After=ovoo-api.service
```

### `ovoo-socketmap.service`

```ini
[Unit]
Description=Ovoo Mail Aliasing Socketmap Service

[Service]
User=ovoo
Group=ovoo
ExecStart=/usr/local/bin/ovoo socketmap -config /usr/local/etc/ovoo/config.json

[Install]
WantedBy=multi-user.target
Wants=ovoo-api.service
After=ovoo-api.service
```

---

## 6. Postfix multi-instance setup

This setup uses the Postfix `postmulti` multi-instance manager. Two named instances are created — `postfix-in` and `postfix-out` — each with its own config, queue, and data directories. The default `/etc/postfix` instance acts only as the manager and does not accept any SMTP connections itself.
This architecture solves issue for OpenDKIM to be able to validate DKIM signatures for incoming emails and also correctly sign outgoing mail after headers has been rewritten.

### 6.1. Prepare instance directories

```bash
for inst in postfix-in postfix-out; do
  install -d /etc/$inst
  install -d /var/lib/$inst
  install -d /var/spool/$inst
done
```

### 6.2. Default Postfix instance (multi-instance manager)

Replace `/etc/postfix/main.cf` with the following. This disables SMTP on the default instance and registers the two named instances:

```
compatibility_level = 3.8
myhostname = mx.ovoodomain.example
mydomain   = ovoodomain.example
myorigin   = $mydomain

# Disable direct SMTP listening on this (manager) instance
master_service_disable = inet

# No local delivery
mydestination        =
local_transport      = error:5.1.1 No local delivery
alias_database       =
alias_maps           =
local_recipient_maps =

# Multi-instance manager — instances are started in the order listed
multi_instance_enable      = yes
multi_instance_wrapper     = ${command_directory}/postmulti -p --
multi_instance_directories = /etc/postfix-out /etc/postfix-in
```

### 6.3. Register and configure instances

```bash
# Initialise the multi-instance manager (only needed once)
postmulti -e init

# Register the two instances
postmulti -I postfix-in  -G mta -e create
postmulti -I postfix-out -G mta -e create

# Enable both instances
postmulti -i postfix-in  -e enable
postmulti -i postfix-out -e enable
```

### 6.4. Inbound instance (`/etc/postfix-in`)

Copy the provided configuration files:

```bash
cp etc/postfix/in/main.cf   /etc/postfix-in/main.cf
cp etc/postfix/in/master.cf /etc/postfix-in/master.cf
```

**`etc/postfix/in/main.cf`** — key settings:

| Setting | Value | Purpose |
|---|---|---|
| `inet_interfaces` | `all` | Accept mail from the Internet on all interfaces. |
| `relay_domains` | `socketmap:inet:127.0.0.1:7788:relay_domain` | Query the Ovoo socketmap to determine which domains to accept mail for. |
| `default_transport` | `smtp:[127.0.0.1]:10026` | Route all accepted mail to the postfix-out re-injection port. |
| `smtpd_milters` | `inet:127.0.0.1:8891 inet:127.0.0.1:6785` | Run OpenDKIM (verify) and Ovoo milter (alias rewrite) on inbound messages. |
| `milter_default_action` | `tempfail` | Temporarily reject messages if a milter is unavailable. |
| `smtp_send_xforward_command` | `yes` | Pass original client IP through to postfix-out for logging. |
| `mydestination` | *(empty)* | No local delivery — this is a gateway only. |
| `multi_instance_name` | `postfix-in` | Registers this instance with the postmulti manager. |

**`etc/postfix/in/master.cf`** — key entries:

The public SMTP listener is configured with strict recipient restrictions and SPF enforcement:

```
smtp      inet  n  -  y  -  -  smtpd
  -o smtpd_recipient_restrictions=reject_invalid_helo_hostname,reject_non_fqdn_sender,\
reject_unknown_sender_domain,reject_non_fqdn_recipient,reject_unknown_recipient_domain,\
reject_unauth_destination,reject_rbl_client,permit
  -o smtpd_helo_required=yes
  -o disable_vrfy_command=yes
  -o smtpd_relay_restrictions=check_policy_service,unix:private/policyd-spf,permit

policyd-spf  unix  -  n  n  -  0
      spawn user=policyd-spf argv=/usr/bin/policyd-spf
```

See `etc/postfix/in/master.cf` for the complete file including all standard Postfix services.

### 6.5. Outbound instance (`/etc/postfix-out`)

Copy the provided configuration files:

```bash
cp etc/postfix/out/main.cf   /etc/postfix-out/main.cf
cp etc/postfix/out/master.cf /etc/postfix-out/master.cf
```

**`etc/postfix/out/main.cf`** — key settings:

| Setting | Value | Purpose |
|---|---|---|
| `inet_interfaces` | `loopback-only` | Accept connections only from localhost — postfix-in is the sole client. |
| `smtpd_authorized_xforward_hosts` | `$mynetworks` | Trust XFORWARD commands from postfix-in so original client info is preserved in logs. |
| `smtpd_milters` | `inet:127.0.0.1:8891` | Apply only OpenDKIM signing. The Ovoo milter must not run here (alias rewriting is already done). |
| `mydestination` | *(empty)* | No local delivery. |
| `multi_instance_name` | `postfix-out` | Registers this instance with the postmulti manager. |

**`etc/postfix/out/master.cf`** — key entry:

The only listener is the loopback re-injection port that postfix-in delivers to:

```
127.0.0.1:10026 inet  n  -  y  -  -  smtpd
```

See `etc/postfix/out/master.cf` for the complete file.

### 6.6. Start Postfix

```bash
systemctl restart postfix
```

Verify both instances are running:

```bash
postmulti -l
```

Expected output (abbreviated):

```
postfix-           -                  y  y  /etc/postfix
postfix-out        mta                y  y  /etc/postfix-out
postfix-in         mta                y  y  /etc/postfix-in
```

---

## 7. OpenDKIM setup

OpenDKIM verifies DKIM signatures on inbound mail (postfix-in) and signs outbound mail (postfix-out). Lua scripts are used for the key and signing tables, reading domain, key path, and selector from environment variables set in a systemd drop-in — no changes to the Lua files are needed when rotating keys or adding domains.

### 7.1. Generate a DKIM key pair

Choose a selector name (e.g. `mail2024`). The selector identifies which public key to use in DNS.

```bash
mkdir -p /etc/opendkim

# Generate a 2048-bit RSA key pair
opendkim-genkey -b 2048 \
  -d ovoodomain.example \
  -s <selector> \
  -D /etc/opendkim/

# Set ownership and permissions
chown -R opendkim:opendkim /etc/opendkim
chmod 600 /etc/opendkim/<selector>.private
chmod 644 /etc/opendkim/<selector>.txt
```

The generated `<selector>.txt` file contains the DNS TXT record to publish. Its content looks similar to:

```
<selector>._domainkey  IN  TXT  ( "v=DKIM1; h=sha256; k=rsa; "
        "p=MIGfMA0GCSqGSIb3DQEBAQUAA4GNA..." )
```

Publish this record under `ovoodomain.example` in your DNS zone.

### 7.2. Main configuration (`/etc/opendkim.conf`)

Copy the provided configuration:

```bash
cp etc/opendkim.conf /etc/opendkim.conf
```

**`etc/opendkim.conf`:**

```
Syslog         yes
SyslogSuccess  yes
LogWhy         yes

Canonicalization  simple
Mode              sv
SubDomains        no
OversignHeaders   From

# Lua-based tables allow per-domain key selection via environment variables
QueryCache    yes
KeyTable      lua:/etc/opendkim/keytable.lua
SigningTable  lua:/etc/opendkim/signingtable.lua
InternalHosts refile:/etc/opendkim/TrustedHosts

UserID  opendkim
UMask   007

# TCP socket — both Postfix instances connect here
Socket  inet:8891@localhost

PidFile  /run/opendkim/opendkim.pid

# DNSSEC trust anchor (provided by the dns-root-data package on Debian/Ubuntu)
TrustAnchorFile  /usr/share/dns/root.key
```

### 7.3. Lua scripts

Copy the scripts to `/etc/opendkim/`:

```bash
cp etc/opendkim/keytable.lua     /etc/opendkim/keytable.lua
cp etc/opendkim/signingtable.lua /etc/opendkim/signingtable.lua
chown opendkim:opendkim /etc/opendkim/keytable.lua /etc/opendkim/signingtable.lua
chmod 644 /etc/opendkim/keytable.lua /etc/opendkim/signingtable.lua
```

**`etc/opendkim/keytable.lua`** — returns `(domain, selector, key_path)` for each signing request:

```lua
-- Reads domain, key path, and selector from environment variables.
local default_domain   = os.getenv("OVOO_OPENDKIM_DEFAULT_DOMAIN")   or "VARIABLE_NOT_SET__OVOO_OPENDKIM_DEFAULT_DOMAIN"
local default_key      = os.getenv("OVOO_OPENDKIM_DEFAULT_KEY")      or "VARIABLE_NOT_SET__OVOO_OPENDKIM_DEFAULT_KEY"
local default_selector = os.getenv("OVOO_OPENDKIM_DEFAULT_SELECTOR") or "VARIABLE_NOT_SET__OVOO_OPENDKIM_DEFAULT_SELECTOR"

local domain   = default_domain
local sign_key = default_key
local selector = default_selector

-- OpenDKIM passes the key name via the global 'query' variable
if query ~= nil then
    domain = query
end

return domain, selector, sign_key
```

**`etc/opendkim/signingtable.lua`** — maps a sender address to a signing domain:

```lua
-- Extracts the domain part of the sender address for signing table lookup.
local default_domain_name = os.getenv("OVOO_OPENDKIM_DEFAULT_DOMAIN") or "VARIABLE_NOT_SET__OVOO_OPENDKIM_DEFAULT_DOMAIN"
local domain = default_domain_name

-- OpenDKIM passes the sender address via the global 'query' variable
if query ~= nil then
    domain = query:match("@(.+)$") or query
end

return domain
```


### 7.4. SystemD environment override

The Lua scripts read three environment variables at runtime. Set them via a systemd drop-in so they are injected into the OpenDKIM process on startup:

```bash
mkdir -p /etc/systemd/system/opendkim.service.d
```

Create `/etc/systemd/system/opendkim.service.d/override.conf`:

```ini
[Service]
Environment="OVOO_OPENDKIM_DEFAULT_DOMAIN=ovoodomain.example"
Environment="OVOO_OPENDKIM_DEFAULT_KEY=/etc/opendkim/<selector>.private"
Environment="OVOO_OPENDKIM_DEFAULT_SELECTOR=<selector>"
```

Replace `<selector>` with the selector name chosen in step 7.1.

Apply and start:

```bash
systemctl daemon-reload
systemctl enable --now opendkim
```

---

## 8. Verification

### 8.1. Service status

All six services should be active:

```bash
systemctl status ovoo-api ovoo-milter ovoo-socketmap opendkim postfix
```

### 8.2. Port binding check

Confirm every service is listening on its expected address:

```bash
ss -tlnp | grep -E ':8808|:6785|:7788|:8891|:10026|:25'
```

Expected bindings:

| Port | Service |
|---|---|
| `0.0.0.0:25` | postfix-in (public SMTP) |
| `127.0.0.1:10026` | postfix-out (re-injection) |
| `127.0.0.1:8891` | OpenDKIM |
| `127.0.0.1:6785` | Ovoo milter |
| `127.0.0.1:7788` | Ovoo socketmap |
| `0.0.0.0:8808` | Ovoo API |

### 8.3. Ovoo API health check

```bash
curl -sk https://127.0.0.1:8808/api/v1/sysinfo | jq .
```

Should return a JSON object including the configured DKIM domain and selector.

### 8.4. Socketmap domain query

```bash
postmap -q ovoodomain.example socketmap:inet:127.0.0.1:7788:relay_domain
```

A response beginning with `OK` means the socketmap is reachable and the domain is known to Ovoo.

### 8.5. Send a test message

Use `swaks` to inject a test message and observe it passing through the pipeline:

```bash
sudo apt- install swaks
swaks --to alias@ovoodomain.example \
      --from sender@example.com \
      --server mx.ovoodomain.example:25 \
      --header "Subject: Ovoo pipeline test"
```

Watch the logs in real time:

```bash
journalctl -u ovoo-milter -f &
journalctl -u ovoo-api    -f &
tail -f /var/log/mail.log
```

### 8.6. DKIM signature check

A successfully signed and verified message will contain headers like:

```
Authentication-Results: mx.ovoodomain.example;
    dkim=pass header.d=ovoodomain.example header.s=<selector>;
    spf=pass smtp.mailfrom=sender@example.com
DKIM-Signature: v=1; a=rsa-sha256; c=simple/simple; d=ovoodomain.example;
    s=<selector>; ...
```

You can verify the signing configuration locally:

```bash
echo "From: test@ovoodomain.example" | opendkim-testmsg -vvv
```

### 8.7. Troubleshooting reference

| Symptom | Where to look |
|---|---|
| Mail rejected with "relay access denied" | Check `ovoo-socketmap` logs; confirm the alias domain is registered in Ovoo via the WebUI. |
| "Milter connection refused" in mail.log | Confirm `ovoo-milter` is running; verify `milter.listen_addr` in `config.json` matches `smtpd_milters` in postfix-in `main.cf`. |
| DKIM signing failures | Check OpenDKIM env vars in the systemd override; verify the private key file path and permissions (`chmod 600`). |
| API returns HTTP 401 | Confirm `auth_token` in `config.json` matches a valid token created in the Ovoo WebUI under Settings → API Tokens. |
| SPF failures on legitimate mail | Review `smtpd_relay_restrictions` in postfix-in `master.cf`; check `journalctl -u postfix` for policyd-spf output. |
| postfix-out not signing outbound | Ensure postfix-out `smtpd_milters` points only to OpenDKIM (`inet:127.0.0.1:8891`), not the Ovoo milter. |
