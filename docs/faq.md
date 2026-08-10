# FAQ

Real questions, with answers grounded in how DockSight actually behaves.

---

## Installation

### I downloaded the binary and got `Syntax error: newline unexpected`

The file is not a binary. It is GitHub's HTML "not found" page, saved under the
binary's name because the download URL was wrong and `curl` was not told to fail:

```console
$ head -c 16 docksight-cli-v0.0.13-linux-amd64 | od -c
0000000  \n  \n  \n  \n      \n  \n   <   !   D   O   C   T   Y   P
```

The shell tried to execute HTML and choked on line 8 of the markup. Confirm with
`file`:

```bash
file docksight-cli-*     # want: ELF 64-bit LSB executable, x86-64, statically linked
```

Always download with `-f`, which turns a 404 into an error instead of a saved
error page:

```bash
curl -fLO https://github.com/Open-Source-Kigali/docksight/releases/latest/download/docksight-cli-v0.0.13-linux-amd64
```

The usual root cause is that the asset genuinely does not exist on that release
— check before blaming the network:

```bash
curl -s https://api.github.com/repos/Open-Source-Kigali/docksight/releases/latest | grep '"name": "docksight'
```

### `Permission denied` when running the downloaded binary

The execute bit is missing. GitHub serves release assets over plain HTTP with no
file mode, so every download arrives as `0644`:

```bash
chmod +x docksight-cli-v0.0.13-linux-amd64
```

This is also true of assets built on Windows, where NTFS has no execute bit to
preserve in the first place.

### `cannot execute binary file: Exec format error`

Wrong architecture for the machine. Check and re-download:

```bash
uname -m     # x86_64 → -linux-amd64,  aarch64 → -linux-arm64
```

### Can I install DockSight on Windows or macOS?

The **agent** installs on Windows. The **platform** does not, on either.

`docksight agent install` works end to end on Windows: it registers the agent
with the Service Control Manager, writes its configuration under
`C:\ProgramData\DockSight\agent`, and verifies the connection by reading the
agent's log. See [Windows support](agent.md#windows-support) for the paths, the
elevation requirement and what differs from systemd.

The platform installer still rejects anything but Linux during validation:

```text
✗ system validation failed: DockSight installs on Linux only, this host runs windows
```

It brings up a Compose stack against Linux filesystem paths, which has no
Windows equivalent yet. macOS is not supported for either: the release publishes
a macOS **CLI** build, but it can only run `version` and `--help`.

### `this process is not elevated, and registering a Windows service needs Administrator rights`

Registering a service writes to the Service Control Manager, which needs
Administrator. Start an elevated PowerShell — right-click *Windows PowerShell* →
*Run as administrator* — and run the install again.

The check runs during validation, before anything is downloaded or written, so a
failure here leaves the host untouched.

### Do I need to install the platform before the agents?

Yes, in practice. The agent needs a reachable platform URL, and installation
verification checks that it connected. You *can* install an agent first — it
will install and then retry the connection every 5 seconds until the platform
appears — but you lose the verification signal.

### The installer says `the Docker daemon is not running`

```bash
sudo systemctl start docker
sudo systemctl enable docker
```

If the message mentions *permission denied* instead, the daemon is up but your
user cannot reach its socket. Use `sudo`.

### `systemd is not the init system on this host`

On Linux the agent is a systemd service, so systemd must be PID 1. Common
causes: you are inside a container, or in WSL without systemd. For WSL:

```ini title="/etc/wsl.conf"
[boot]
systemd=true
```

Then `wsl --shutdown` and reopen.

This check does not run on Windows, where the Service Control Manager takes
systemd's place.

### Can I install without internet access?

Not today. Both installers resolve releases from `api.github.com` and download
assets from GitHub. An air-gapped installation would need the artifacts staged
manually, which is not yet supported.

---

## Networking

### Do I need to open ports on my Docker hosts?

**No.** That is the central design decision. Agents always dial out; nothing ever
connects to them. Hosts behind NAT, in another cloud, or on a home network work
with no port forwarding and no VPN. See
[why agents dial out](architecture.md#why-agents-dial-out).

### Which port does the platform use?

`2002` by default, published by nginx. Change `DOCKSIGHT_PORT` in
`/opt/docksight/.env` and restart the stack.

### Does the agent connection use TLS?

Only if you tell it to. The scheme is derived from the platform URL you give at
install time: `https://` becomes `wss://`, `http://` becomes `ws://`. Installing
with `http://10.0.0.5:2002` gives you an unencrypted session and nothing warns
you at runtime. See [Security](security.md#current-limitations).

### The agent installed fine but never connects

```bash
journalctl -u docksight-agent -n 50 --no-pager
grep url /etc/docksight-agent/config.yaml
```

In order of likelihood:

1. **Wrong port** in the platform URL. `:200` and `:2002` are both accepted at
   install time; only one of them is your platform.
2. Platform not reachable from that host — test with
   `curl -I http://platform:2002`.
3. Egress blocked by a firewall.

Fix the URL without reinstalling:

```bash
sudo docksight agent update --url http://platform.example.com:2002
```

### Can one platform manage hosts in different networks?

Yes. Each host only needs outbound reachability to the platform. They do not need
to reach each other.

### What happens when the platform goes down?

Containers keep running — the agent does not stop anything. The agent retries
every 5 seconds and re-registers when the platform returns. Only *management* is
unavailable in the meantime.

---

## Docker permissions

### Why does the agent run as root?

It reads `/var/run/docker.sock`, which is root-owned. Running as a non-root user
would require adding that user to the `docker` group — which grants
root-equivalent access anyway, just less visibly.

### Is giving DockSight Docker socket access dangerous?

It is exactly as dangerous as Docker access always is: **socket access is
root-equivalent on the host**. Anyone who can use it can start a container that
mounts `/`.

What DockSight does to limit exposure: the socket is read locally and never
published to the network, the agent exposes no inbound port, and the agent
implements a fixed set of operations rather than proxying arbitrary Docker API
calls. See [Security](security.md#the-docker-socket).

### Can I run the agent inside a container?

It is not supported. The installer requires systemd, and a containerised agent
would need the Docker socket bind-mounted in — which is precisely the pattern
security guidance warns against.

---

## Release process

### Why did my install fail with "publishes no asset matching docksight-agent-*"?

The release exists but nobody uploaded the agent binaries to it. The error lists
what *is* published:

```text
release v0.0.12 publishes no asset matching docksight-agent-* for linux-amd64
(has: docksight-cli-v0.0.12-linux-amd64, docksight-platform-v0.0.12.tar.gz, ...)
```

Fix it on the release side:

```bash
scripts/build-release.sh v0.0.12
scripts/publish-release.sh v0.0.12
```

`publish-release.sh` verifies against the API after uploading, which is what
prevents this recurring.

### Why doesn't `latest` point at my newest release?

GitHub's `/releases/latest` excludes drafts and prereleases. If the newest
release is either, installers keep resolving the previous one. Publish it
properly, or pin with `--version`.

### Can I run mixed versions?

Yes, by design. The CLI, platform and agent version independently, and the
protocol ignores unknown message types and unknown payload fields. A v0.0.9
agent works against a v0.0.13 platform, using only the subset it knows.

### Should release artifacts be committed to git?

No. `release/` is git-ignored. The artifacts are build outputs, rebuildable from
any tag, and binaries bloat repository history permanently.

---

## WebSocket and protocol

### Why WebSocket instead of REST polling or gRPC?

The platform must push commands and receive streams. Polling adds latency and
wastes requests; an inbound HTTP API on the agent would require open ports; gRPC
adds a proto toolchain across a Go/TypeScript boundary and needs extra plumbing
through HTTP/1.1 proxies. WebSocket gives one long-lived bidirectional
connection that traverses ordinary proxies. See
[why WebSocket](architecture.md#why-websocket).

### How do I know an agent is alive?

It sends `agent.heartbeat` every 30 seconds. TCP keepalives are not enough — a
connection can stay open through a NAT device long after the process behind it
stopped working. The heartbeat proves the agent's event loop is running.

### Why is there no heartbeat acknowledgement?

Silence is success. An ack would double heartbeat traffic to tell the agent
something the socket already tells it.

### Can I write my own agent?

Yes. [`protocol.md`](protocol.md) is a complete specification: implement
`agent.register`, `agent.heartbeat`, and whichever domains you need. Keep the
`{ type, payload }` envelope and ignore unknown fields.

---

## Platform

### Where is my data stored?

In Docker named volumes (`postgres-data`, `redis-data`), not in
`/opt/docksight`. Reinstalling the platform does not touch them.

### How do I back up?

Back up the `postgres-data` volume **and** `/opt/docksight/.env` together. The
password in `.env` must match the volume it was initialised with; restoring one
without the other leaves you locked out.

### Why won't `docker compose` work when I run it as my own user?

`.env` is `0600` and root-owned. Compose running as another user cannot read it,
so every variable interpolates empty — Postgres refuses a blank password, and
`JWT_SECRET` silently falls back to the placeholder in the compose file. Compose
reports this only as `warning: The "POSTGRES_PASSWORD" variable is not set`. Run
Compose with `sudo`.

### Can I change the generated passwords?

Not simply. `POSTGRES_PASSWORD` is baked into the database volume at first boot,
so changing it in `.env` alone locks the platform out. Rotating it means
changing it inside PostgreSQL as well. `JWT_SECRET` can be changed freely — it
only invalidates existing sessions.

### Does `docksight status` work?

Not yet. It prints a fixed string. Use:

```bash
docker compose -f /opt/docksight/dockersight-installation.yml ps
```

---

## Agent

### What is `identity.json` and can I delete it?

It holds the UUID that makes a machine *the same host* across restarts and
upgrades. Deleting it makes the agent generate a new UUID, so the platform sees a
brand-new host and the old record becomes an orphan. Never copy it between
machines — two agents claiming one identity is worse than two identities.

### Does upgrading the agent lose its registration?

No. `agent install` and `agent update` both preserve `identity.json`, so the host
keeps its identity and history.

### Do I have to retype the platform URL when updating?

No — `docksight agent update` reads it back from `/etc/docksight-agent/config.yaml`.
That is deliberate: retyping is how a wrong port gets introduced.

### How do I roll an agent back?

```bash
sudo docksight agent update --version v0.0.9
```

Only releases that actually publish agent binaries are valid targets.

### Why don't `docksight agent start` / `logs` work?

They are declared placeholders and return "not implemented yet". Use `systemctl`
and `journalctl` — see the [CLI reference](cli.md#placeholder-commands).

### The install said "running but no connection confirmed"

The service is healthy and stayed up, but the journal had no connection line yet
when verification finished — common on a slow link. Watch it:

```bash
journalctl -u docksight-agent -f
```

Look for `websocket connected` followed by `registration acknowledged`.

---

## Troubleshooting

### The installer failed halfway — is my host broken?

No. Installation writes to a private temporary directory first and only copies
into place after a successful download and extraction. A failure before that
point leaves nothing behind; a failure at the compose stage leaves the files
installed so you can inspect and retry. Every install command is safe to re-run.

### `open /tmp/docksight-install-...: permission denied`

An older CLI staged downloads at a fixed `/tmp` path, which collides with a
root-owned file left by an earlier `sudo` run — `/tmp` is sticky-bit `1777`, so
you may create files there but not truncate someone else's. Current versions
stage in a private per-run directory. Upgrade the CLI, or `sudo rm` the stale
file.

### A service is stuck restarting after install

Installation reports this rather than claiming success:

```text
✗ service failed to start: server: exited (code 3)
```

Then:

```bash
docker compose -f /opt/docksight/dockersight-installation.yml logs server
journalctl -u docksight-agent -n 100 --no-pager
```

### How do I completely start over?

=== "Platform"

    ```bash
    cd /opt/docksight
    docker compose -f dockersight-installation.yml down -v   # -v deletes data
    sudo rm -rf /opt/docksight /var/lib/docksight
    ```

=== "Agent"

    ```bash
    sudo systemctl disable --now docksight-agent
    sudo rm /etc/systemd/system/docksight-agent.service
    sudo systemctl daemon-reload
    sudo rm -rf /etc/docksight-agent /usr/local/bin/docksight-agent
    ```

    Removing `/etc/docksight-agent` discards the identity, so the host will
    register as new.

---

## Still stuck?

Open an issue with the command you ran, the full output, `docksight version`,
and the relevant logs. See [Contributing](contributing.md#reporting-bugs) —
and redact `.env` contents before pasting.
