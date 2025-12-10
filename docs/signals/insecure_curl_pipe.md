# Insecure Curl Pipe

## What this is

This signal detects recent use of the insecure `curl | bash` / `curl | sh` pattern in your shell history.

It approximates:

```bash
tail -3 ~/.$(basename "$SHELL")_history
```

and scans only the **last three commands** for patterns like:

- `curl https://example.com/install.sh | bash`
- `curl -sL https://sh.rustup.rs | sh`
- `curl -sSL https://installer | sudo bash`

The detection is intentionally tolerant of common flags (`-s`, `-S`, `-L`, `-sSL`, etc.) and whitespace so it catches real-world usage, including zsh extended history lines like:

```bash
: 1700000000:0;curl -sSL https://sh.rustup.rs | sh
```

## Why this matters

Running `curl ... | bash` (or `curl ... | sh`) is **convenient but dangerously brittle**. It turns a single network request into immediate code execution with almost no safety rails.

**Security risks:**

- **Partial execution vulnerability**: If the network drops halfway through download, your shell may execute a truncated script, leaving your system in an inconsistent or exploitable state.
- **Server-side pipe detection**: Many servers can detect when their output is being piped directly into a shell and serve **different content** than they would for a normal file download, enabling targeted supply-chain attacks.
- **Fileless malware / forensic evasion**: Because the script is never written to disk, there may be **no artifacts** for incident responders to review later.
- **Erosion of trust chain**: The only thing you are trusting is **TLS to that hostname in that moment**. There is no secondary integrity check (checksum, signature, or pinning) and no durable record of what you actually ran.

**Example attack:**

```bash
# Looks innocuous, but executes whatever that server responds with right now
curl -sL https://sh.rustup.rs | sh
```

If an attacker controls DNS, the origin server, or a CDN edge, they can transparently replace this with a backdoored installer or staged payload.

## How to remediate

### Prefer checksum.sh for one-liner installers

[checksum.sh](https://github.com/gavinuhma/checksum.sh) wraps the common one-liner installer pattern with a **mandatory integrity check**.

Transform this:

```bash
curl -sL https://sh.rustup.rs | sh
```

into this:

```bash
checksum https://sh.rustup.rs <SHA256> | sh
```

Where `<SHA256>` is the hash published by the project (release notes, README, or official download page). `checksum` will:

1. Download the script
2. Verify its SHA-256 checksum
3. Only pipe it to `sh` **if the checksum matches**

If the checksum does not match, **nothing is executed**.

### Use a download-inspect-execute workflow

Instead of executing remote code as it streams over the network, separate the steps:

```bash
# 1. Download
curl -sL https://example.com/install.sh -o /tmp/install.sh

# 2. Inspect
$EDITOR /tmp/install.sh    # or use less, bat, etc.

# 3. Execute explicitly
bash /tmp/install.sh
```

This gives you a stable artifact you can:

- Re-run later
- Store in version control for reproducibility
- Compare across machines or over time

### Manually verify checksums

If a project publishes checksums alongside installers, you can verify them yourself:

```bash
# Download script and checksum
curl -sL https://example.com/install.sh -o install.sh
curl -sL https://example.com/install.sh.sha256 -o install.sh.sha256

# Verify (Linux/macOS)
sha256sum -c install.sh.sha256  # or: shasum -a 256 -c install.sh.sha256

# Only execute if verification succeeds
bash install.sh
```

### Use vipe for interactive review

If you strongly prefer a one-liner but still want human review, use `vipe` (from [moreutils](https://joeyh.name/code/moreutils/)) to interpose an editor:

```bash
curl -sL https://example.com/install.sh | vipe | bash
```

`vipe` drops you into an editor with the fetched script; you save and quit to continue, or exit without saving to abort.

## Security best practices

- **Avoid `curl | bash` / `curl | sh` entirely** in automation, CI, and documentation.
- **Prefer package managers** (apt, dnf, brew, winget, etc.) or signed packages when available.
- **Always introduce an integrity check** when executing remote code: checksums (SHA-256), signatures (GPG, Sigstore), or tools like `checksum.sh`.
- **Keep a copy** of any installer scripts you run so you can compare and audit them later.
- **Review what you run**, especially when copying commands from blogs, gists, or chat.


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_INSECURE_CURL_PIPE=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
