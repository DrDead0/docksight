# CLI reference

The `docksight` binary installs, updates and inspects both halves of DockSight.
This page documents every command that exists today, including the ones that are
placeholders.

---

## Command tree

```text
docksight
├── install          Install the platform on this server
├── update           Update the CLI and/or the platform
├── status           Show platform status          (placeholder)
├── version          Show CLI and platform versions
└── agent
    ├── install      Install the agent on this host
    ├── update       Update the agent on this host
    ├── uninstall    Remove the agent               (placeholder)
    ├── start        Start the agent service        (placeholder)
    ├── stop         Stop the agent service         (placeholder)
    ├── restart      Restart the agent service      (placeholder)
    ├── status       Show agent service status      (placeholder)
    ├── logs         Show agent service logs        (placeholder)
    └── config       Show or edit agent config      (placeholder)
```

Placeholders exist so the command surface is discoverable, but they fail loudly
rather than pretending to work:

```console
$ sudo docksight agent status
✗ docksight agent status is not implemented yet
```

---

## Global behaviour

**Exit codes** — `0` on success, `1` on any failure.

**Error output** — failures print once, prefixed with `✗`, without a usage dump:

```console
$ sudo docksight agent install --url http://platform:2002
✗ system validation failed: DockSight installs on Linux only, this host runs windows
```

A *usage* mistake still shows the flag reference:

```console
$ docksight agent install --nope
✗ unknown flag: --nope
Usage:
  docksight agent install [flags]
Flags:
      --url string       DockSight platform URL, e.g. https://platform.example.com
      --version string   install a specific release tag instead of the latest
```

**Privileges** — every install and update command writes to `/opt`, `/etc` or
`/usr/local/bin` and must run as root.

---

## `docksight install`

Installs the platform on this server.

```bash
sudo docksight install
```

**Flags:** none.

**What it does**

1. Validates OS, architecture, Docker Engine, Docker daemon and Compose
2. Prints the resolved configuration
3. Creates `/opt/docksight` and `/var/lib/docksight`
4. Installs the running binary to `/usr/local/bin/docksight`
5. Resolves the latest GitHub release
6. Downloads and extracts the platform bundle
7. Copies it into `/opt/docksight`
8. Generates `.env` with fresh credentials (only if absent)
9. Runs `docker compose up -d`
10. Waits until every service is running and healthy

**Output**

```console
$ sudo ./docksight-cli-v0.0.13-linux-amd64 install
→ [1/5] Checking operating system
✓ Checking operating system
...
→ Installing the CLI into /usr/local/bin/docksight
✓ CLI installed from /root/docksight-cli-v0.0.13-linux-amd64
→ Checking the latest DockSight release
✓ Latest release v0.0.13
✓ Installer package docksight-platform-v0.0.13.tar.gz found
→ Downloading installer package
✓ Download completed
✓ Platform installed
✓ Generated /opt/docksight/.env
→ Starting DockSight services
→ Waiting for services to become ready
✓ postgres: running (healthy)
✓ redis: running (healthy)
✓ server: running
✓ web: running
✓ nginx: running
✓ DockSight is running on http://localhost:2002
```

**Safe to re-run.** A second run upgrades the bundle, preserves `.env`, and
recreates the stack.

---

## `docksight update`

Updates the CLI and the platform. They are versioned independently.

```bash
sudo docksight update
```

| Flag | Type | Description |
| --- | --- | --- |
| `--cli` | bool | Update only the CLI binary |
| `--platform` | bool | Update only the platform bundle |
| `--version` | string | Target a specific release tag instead of the latest |
| `--force` | bool | Re-apply even if already at the target version |

`--cli` and `--platform` are mutually exclusive.

**Examples**

```bash
sudo docksight update                      # both, to latest
sudo docksight update --platform           # platform only
sudo docksight update --cli                # CLI only, stack untouched
sudo docksight update --version v0.0.13    # pin
sudo docksight update --version v0.0.9     # roll back
sudo docksight update --force              # re-apply current version
```

**Output when already current**

```console
$ sudo docksight update
→ Checking the latest DockSight release
✓ Latest release v0.0.13
→ CLI already at v0.0.13
→ Platform already at v0.0.13
✓ DockSight is up to date
```

**Internal behaviour** — reads `/opt/docksight/state.json` to decide what needs
work. A platform update replaces the bundle then runs `compose down` and `up`,
so there is a brief outage. `.env` and named volumes survive.

---

## `docksight version`

```console
$ docksight version
→ CLI version: v0.0.13
→ Platform version: v0.0.13
```

On a host with no platform installed:

```console
$ docksight version
→ CLI version: v0.0.13
→ Platform version: not installed
```

The CLI version is stamped at build time with `-ldflags`; a locally built binary
reports `dev`. The platform version comes from `state.json`.

---

## `docksight status`

```console
$ docksight status
→ Docksight status: all active
```

!!! warning "Placeholder"
    This prints a fixed string and does not inspect anything. Use
    `docker compose -f /opt/docksight/dockersight-installation.yml ps` instead.
    Tracked in the [Roadmap](roadmap.md#in-progress).

---

## `docksight agent install`

Installs the agent on this Docker host.

```bash
sudo docksight agent install --url https://platform.example.com
```

| Flag | Type | Description |
| --- | --- | --- |
| `--url` | string | Platform URL. Prompts interactively if omitted. |
| `--version` | string | Install a specific release tag instead of the latest |

**Examples**

```bash
# Interactive
sudo docksight agent install

# Non-interactive (required in scripts — a prompt would hang)
sudo docksight agent install --url https://platform.example.com

# Pinned
sudo docksight agent install --url https://platform.example.com --version v0.0.13

# Plain HTTP on a private network
sudo docksight agent install --url http://10.0.0.5:2002
```

**Output**

```console
→ Agent binary:  /usr/local/bin/docksight-agent
→ Configuration: /etc/docksight-agent/config.yaml
→ Service:       docksight-agent.service
→ Platform:      wss://platform.example.com/agents
→ Validating the host
→ [1/6] Checking operating system
✓ Checking operating system
→ [2/6] Checking architecture
✓ Checking architecture
→ [3/6] Checking Docker Engine
✓ Checking Docker Engine
→ [4/6] Checking Docker daemon
✓ Checking Docker daemon
→ [5/6] Checking systemd
✓ Checking systemd
→ [6/6] Checking internet connectivity
✓ Checking internet connectivity
→ Checking the latest DockSight release
✓ Latest release v0.0.13
→ Downloading docksight-agent-v0.0.13-linux-amd64
✓ Agent binary installed at /usr/local/bin/docksight-agent
→ Writing /etc/docksight-agent/config.yaml
✓ Configured platform wss://platform.example.com/agents
→ Identity will be created at /etc/docksight-agent/identity.json on first connection
→ Creating the systemd service
✓ Unit written to /etc/systemd/system/docksight-agent.service
✓ Service enabled at boot
→ Starting docksight-agent.service
✓ docksight-agent.service is running
✓ Agent connected to wss://platform.example.com/agents
✓ DockSight Agent installed and connected
```

**Safe to re-run.** Re-running upgrades the binary, rewrites the config and
restarts the service. `identity.json` is preserved, so the host keeps its
registration.

!!! note "This command does not put the CLI on PATH"
    Only `docksight install` self-installs. On an agent-only host, keep the
    downloaded binary or move it yourself:
    `sudo mv docksight-cli-* /usr/local/bin/docksight`.

---

## `docksight agent update`

Updates the agent, reusing the platform URL already on disk.

```bash
sudo docksight agent update
```

| Flag | Type | Description |
| --- | --- | --- |
| `--version` | string | Update to a specific release tag instead of the latest |
| `--url` | string | Repoint the agent at a different platform |

**Examples**

```bash
sudo docksight agent update                                    # latest
sudo docksight agent update --version v0.0.13                  # pinned
sudo docksight agent update --version v0.0.9                   # roll back
sudo docksight agent update --url https://new-platform.example.com
```

**With no agent installed**

```console
$ sudo docksight agent update
✗ no agent installation found at /etc/docksight-agent/config.yaml: run 'docksight agent install --url ...' first
```

**Internal behaviour** — identical to `agent install`, except the platform URL is
read from `config.yaml` instead of a flag. Retyping a URL is how a wrong port
gets introduced, so the update path avoids it.

---

## Placeholder commands

These are declared but not implemented:

| Command | Use instead |
| --- | --- |
| `docksight agent uninstall` | Manual steps in [Installation](installation.md#uninstalling) |
| `docksight agent start` | `sudo systemctl start docksight-agent` |
| `docksight agent stop` | `sudo systemctl stop docksight-agent` |
| `docksight agent restart` | `sudo systemctl restart docksight-agent` |
| `docksight agent status` | `systemctl status docksight-agent` |
| `docksight agent logs` | `journalctl -u docksight-agent -f` |
| `docksight agent config` | `cat /etc/docksight-agent/config.yaml` |

---

## Common workflows

=== "First platform install"

    ```bash
    curl -fLO https://github.com/rodriguecyber/docksight/releases/latest/download/docksight-cli-v0.0.13-linux-amd64
    chmod +x docksight-cli-v0.0.13-linux-amd64
    sudo ./docksight-cli-v0.0.13-linux-amd64 install
    ```

=== "Add a Docker host"

    ```bash
    curl -fLO https://github.com/rodriguecyber/docksight/releases/latest/download/docksight-cli-v0.0.13-linux-amd64
    chmod +x docksight-cli-v0.0.13-linux-amd64
    sudo ./docksight-cli-v0.0.13-linux-amd64 agent install --url https://platform.example.com
    ```

=== "Upgrade everything"

    ```bash
    # On the platform server
    sudo docksight update

    # On each agent host
    sudo docksight agent update
    ```

=== "Roll back one agent"

    ```bash
    sudo docksight agent update --version v0.0.9
    ```

---

## Related

- [Installation](installation.md) — walkthroughs and troubleshooting
- [Platform](platform.md) — what `install` and `update` actually change
- [Agent](agent.md) — configuration and service management
