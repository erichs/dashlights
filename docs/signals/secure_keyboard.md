# Secure Keyboard Entry

## What this is

This signal detects when Terminal.app, iTerm2, or Ghostty is running **without** Secure Keyboard Entry enabled.

**Signal behavior**: Only triggers when a terminal application is:
1. Currently running, AND
2. Has Secure Keyboard Entry disabled (or never enabled)

**Supported applications**:
- **Terminal.app**: macOS built-in terminal
- **iTerm2**: Popular third-party terminal emulator
- **Ghostty**: GPU-accelerated terminal emulator

**Platform**: macOS only

## Why this matters

### Keylogger Protection

Secure Keyboard Entry is a macOS security feature that prevents other applications from intercepting keystrokes sent to the terminal. Without it enabled:

- **Credential theft**: Malicious software can capture passwords, API tokens, and other secrets as you type them
- **Command interception**: Attackers can see every command you execute, including those containing sensitive data
- **Session hijacking**: SSH passwords, sudo credentials, and database passwords are all vulnerable

### Real-World Scenarios

**Scenario 1: Malware infection**
A malicious application (installed via compromised software or phishing) runs a keylogger in the background. Every time you type `sudo`, enter your password, or paste an API token, it gets captured.

**Scenario 2: Credential harvesting**
You're working in a coffee shop. Another user on the same network tricks your Mac into running a background process. Without Secure Keyboard Entry, they can harvest your keystrokes remotely.

**Scenario 3: Development secrets**
You paste AWS credentials, database passwords, or GitHub tokens into your terminal. Without protection, these can be intercepted by any malicious process with accessibility permissions.

### Developer Hygiene

For developers, terminals are where sensitive operations happen:
- `ssh` with passwords or passphrases
- `git push` with credentials
- Environment variable exports with API keys
- Database connections with passwords
- AWS/GCP/Azure CLI commands
- Kubernetes `kubectl` with secrets

## How to remediate

### Terminal.app

**Enable Secure Keyboard Entry**:

1. Open Terminal.app
2. Click **Terminal** in the menu bar
3. Click **Secure Keyboard Entry** to enable it (a checkmark appears when enabled)

**Or use the keyboard shortcut**: Press `Command + Shift + K`

**Verify it's enabled**:
```bash
defaults read com.apple.Terminal SecureKeyboardEntry
# Returns 1 when enabled
```

### iTerm2

**Enable Secure Keyboard Entry**:

1. Open iTerm2
2. Click **iTerm2** in the menu bar
3. Click **Secure Keyboard Entry** to enable it (a checkmark appears when enabled)

**Verify it's enabled**:
```bash
defaults read com.googlecode.iterm2 "Secure Input"
# Returns 1 when enabled
```

### Ghostty

**Enable Secure Keyboard Entry**:

1. Open Ghostty
2. Click **Ghostty** in the menu bar
3. Click **Secure Keyboard Entry** to enable it (a checkmark appears when enabled)

**Verify it's enabled**:
```bash
defaults read com.mitchellh.ghostty SecureInput
# Returns 1 when enabled
```

### Making it Permanent

Terminal.app, iTerm2, and Ghostty remember your Secure Keyboard Entry preference. Once enabled, it persists across sessions and restarts.

**Note**: Some users disable it because it can interfere with certain accessibility features or automation tools. If you need to disable it, understand the security implications.

## Security best practices

1. **Always enable Secure Keyboard Entry** when working with sensitive data

2. **Check the setting regularly**: The setting can be toggled accidentally via keyboard shortcut

3. **Be aware of limitations**: Secure Keyboard Entry only protects the specific terminal window. Other applications can still be keylogged.

4. **Use password managers**: Instead of typing passwords, use a password manager with auto-fill

5. **Use SSH keys**: Instead of SSH passwords, use SSH key authentication

6. **Use credential helpers**: Configure Git, AWS CLI, and other tools to use secure credential storage

7. **Avoid pasting secrets**: Use environment variables, secret managers, or credential helpers instead of pasting secrets directly

## How it Works

This signal:
1. Enumerates running processes to check if Terminal.app, iTerm2, or Ghostty is running
2. If running, reads the application's preferences plist to check the Secure Keyboard Entry setting
3. Only signals if the app is running AND the setting is disabled

**Performance**: Uses native process enumeration (~1-2ms) and cached plist reading, well within the 10ms budget.

## Disabling This Signal

To disable this signal, set the environment variable:
```bash
export DASHLIGHTS_DISABLE_SECURE_KEYBOARD=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
