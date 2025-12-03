# Security Signals

Dashlights performs over 30 concurrent security checks, organized into five categories:

## Identity & Access Management (IAM)

1. 🩲 **[Naked Credential](docs/signals/naked_credentials.md)** - Finds raw secrets in environment variables [[code](src/signals/naked_credentials.go)]
2. 🪝 **[AWS CLI Alias Hijacking](docs/signals/aws_alias_hijack.md)** - Detects malicious AWS CLI aliases that override core commands [[code](src/signals/aws_alias_hijack.go)]

## Operational Security (OpSec)

3. 👑 **[Danger Zone](docs/signals/root_login.md)** - Detects when running as root user (UID 0) [[code](src/signals/root_login.go)]
4. 💣 **[Privileged Path](docs/signals/privileged_path.md)** - Detects current directory (`.`) in PATH [[code](src/signals/privileged_path.go)]
5. 🐴 **[Trojan Horse](docs/signals/ld_preload.md)** - Checks for LD_PRELOAD/DYLD_INSERT_LIBRARIES (rootkit vector) [[code](src/signals/ld_preload.go)]
6. 🕶️ **[Blind Spot](docs/signals/history_disabled.md)** - Detects disabled shell history [[code](src/signals/history_disabled.go)]
7. 🚨 **[Prod Panic](docs/signals/prod_panic.md)** - Alerts when kubectl/AWS context points to production [[code](src/signals/prod_panic.go)]
8. 🕵️ **[Man in the Middle](docs/signals/proxy_active.md)** - Alerts on active proxy settings [[code](src/signals/proxy_active.go)]
9. 😷 **[Loose Cannon](docs/signals/permissive_umask.md)** - Checks for permissive umask (0000 or 0002) [[code](src/signals/permissive_umask.go)]
10. 🐳 **[Exposed Socket](docs/signals/docker_socket.md)** - Checks Docker socket permissions and orphaned DOCKER_HOST [[code](src/signals/docker_socket.go)]
11. 🐛 **[Debug Mode Enabled](docs/signals/debug_enabled.md)** - Detects debug/trace/verbose environment variables [[code](src/signals/debug_enabled.go)]
12. 🔐 **[History Permissions](docs/signals/history_permissions.md)** - Checks shell history files for world-readable permissions [[code](src/signals/history_permissions.go)]
13. 🔑 **[SSH Agent Key Bloat](docs/signals/ssh_agent_bloat.md)** - Detects too many keys in SSH agent (causes MaxAuthTries lockouts) [[code](src/signals/ssh_agent_bloat.go)]
14. 🔑 **[Open Door](docs/signals/ssh_keys.md)** - Detects SSH private keys with incorrect permissions [[code](src/signals/ssh_keys.go)]

## Repository Hygiene

15. 📝 **[Unignored Secret](docs/signals/env_not_ignored.md)** - Checks if .env files exist but aren't in .gitignore [[code](src/signals/env_not_ignored.go)]
16. 👑 **[Root-Owned Home Files](docs/signals/root_owned_home.md)** - Finds files in $HOME owned by root [[code](src/signals/root_owned_home.go)]
17. 🖊️ **[World-Writable Configs](docs/signals/world_writable_config.md)** - Detects config files with dangerous permissions [[code](src/signals/world_writable_config.go)]
18. 🗝️ **[Dead Letter](docs/signals/untracked_crypto_keys.md)** - Finds cryptographic keys not in .gitignore [[code](src/signals/untracked_crypto_keys.go)]
19. 🔄 **[Go Replace Directive](docs/signals/go_replace.md)** - Detects replace directives in go.mod (breaks builds) [[code](src/signals/go_replace.go)]
20. 🐍 **[PyCache Pollution](docs/signals/pycache_pollution.md)** - Checks for __pycache__ directories not properly ignored [[code](src/signals/pycache_pollution.go)]
21. 📦 **[NPM RC Tokens](docs/signals/npmrc_tokens.md)** - Detects auth tokens in project .npmrc (should be in ~/.npmrc) [[code](src/signals/npmrc_tokens.go)]
22. 🦀 **[Cargo Path Dependencies](docs/signals/cargo_path_deps.md)** - Checks for path dependencies in Cargo.toml [[code](src/signals/cargo_path_deps.go)]
23. 📁 **[Missing __init__.py](docs/signals/missing_init_py.md)** - Detects Python packages missing __init__.py files [[code](src/signals/missing_init_py.go)]
24. ☕ **[Snapshot Dependency](docs/signals/snapshot_dependency.md)** - Checks for SNAPSHOT dependencies on release branches (Java/Maven) [[code](src/signals/snapshot_dependency.go)]
25. 🎣 **[Pwn Request Risk](docs/signals/pwnrequest.md)** - Detects dangerous GitHub Actions pull_request_target patterns [[code](src/signals/pwnrequest.go)]
26. ⚓ **[Missing Git Hooks](docs/signals/missing_git_hooks.md)** - Detects when hook manager config exists but hooks aren't installed [[code](src/signals/missing_git_hooks.go)]

## System Health

27. 💾 **[Full Tank](docs/signals/disk_space.md)** - Alerts when disk usage exceeds 90% [[code](src/signals/disk_space.go)]
28. ♻️ **[Reboot Pending](docs/signals/reboot_pending.md)** - Detects pending system reboot (Linux) [[code](src/signals/reboot_pending.go)]
29. 🧟 **[Zombie Processes](docs/signals/zombie_processes.md)** - Alerts on excessive zombie processes [[code](src/signals/zombie_processes.go)]
30. 💔 **[Dangling Symlinks](docs/signals/dangling_symlinks.md)** - Detects symlinks pointing to non-existent targets [[code](src/signals/dangling_symlinks.go)]
31. ⏰ **[Time Drift Detected](docs/signals/time_drift.md)** - Detects drift between system time and filesystem time [[code](src/signals/time_drift.go)]

## Infrastructure Security (InfraSec)

32. 🏗️ **[Local Terraform State](docs/signals/terraform_state_local.md)** - Checks for local terraform.tfstate files (should use remote state) [[code](src/signals/terraform_state_local.go)]
33. ☸️ **[Root Kube Context](docs/signals/root_kube_context.md)** - Alerts when Kubernetes context uses kube-system namespace [[code](src/signals/root_kube_context.go)]
34. 🔐 **[Dangerous TF_VAR](docs/signals/dangerous_tf_var.md)** - Checks for dangerous Terraform variables in environment (secrets in shell history) [[code](src/signals/dangerous_tf_var.go)]

