# Docker Socket Exposure

## What this is

This signal detects Docker-related security and operational issues:

1. **Overly permissive permissions** on the Docker socket - when the socket is world-readable or world-writable
2. **Orphaned DOCKER_HOST** environment variable - when `DOCKER_HOST` points to a non-existent socket, causing Docker commands to hang

**Platform-specific behavior**:
- **Linux**: Checks `/var/run/docker.sock` for world-readable (0004) or world-writable (0002) permissions
- **macOS (Docker Desktop)**: Checks if `/var/run/docker.sock` symlink points to a valid user socket (`~/.docker/run/docker.sock`), only flags if world-writable
- **Windows**: Not applicable (Docker Desktop uses named pipes)

## Why this matters

### Security Risk - Permissive Socket (Linux)

On **Linux**, the Docker socket provides root-equivalent access:
- **Root-equivalent access**: Anyone who can access the Docker socket has effective root access to the host system
- **Container escape**: Attackers can create privileged containers and escape to the host
- **Data exfiltration**: Access to all container filesystems and volumes
- **Credential theft**: Can extract secrets from running containers

**Example attack on Linux**:
```bash
# If socket is world-accessible, any user can:
docker run -v /:/host -it ubuntu chroot /host /bin/bash
# Now they have root access to the host filesystem
```

### macOS Docker Desktop Behavior

On **macOS**, Docker Desktop uses a different architecture:
- Socket location: `/var/run/docker.sock` → symlink → `~/.docker/run/docker.sock`
- Ownership: User-owned socket in home directory (e.g., `erichs:staff`)
- Permissions: Typically `srwxr-xr-x` (0755) - world-readable is **safe** on macOS
- Security: Protected by macOS file system permissions on the user's home directory

**Why macOS is different**:
- The actual socket is in the user's home directory, not a system directory
- Only the user (and root) can access files in their home directory
- World-readable permissions on the socket don't grant access to other users
- Docker Desktop manages socket lifecycle automatically

**This signal on macOS**:
- ✅ Allows world-readable permissions (safe due to home directory protection)
- ❌ Flags world-writable permissions (still dangerous)
- ❌ Flags orphaned symlinks (Docker Desktop not running or misconfigured)

### Operational Risk - Orphaned DOCKER_HOST

On **all platforms**:
- **Hanging commands**: Docker commands will hang indefinitely waiting for the non-existent socket
- **CI/CD failures**: Build pipelines fail with timeout errors
- **Confusing errors**: Users get cryptic "cannot connect to Docker daemon" errors

## How to remediate

### macOS (Docker Desktop)

**Check socket configuration**:
```bash
# Check if symlink exists and where it points
ls -la /var/run/docker.sock
# Expected: lrwxr-xr-x ... /var/run/docker.sock -> /Users/<user>/.docker/run/docker.sock

# Check actual socket permissions
ls -la ~/.docker/run/docker.sock
# Expected: srwxr-xr-x ... <user> staff ... docker.sock
```

**If you see this signal on macOS**:

1. **Orphaned symlink** (Docker Desktop not running):
   ```bash
   # Start Docker Desktop
   open -a Docker

   # Wait for Docker to start, then verify
   docker ps
   ```

2. **World-writable socket** (unusual, but dangerous):
   ```bash
   # Fix permissions on user socket
   chmod 755 ~/.docker/run/docker.sock

   # Verify
   ls -la ~/.docker/run/docker.sock
   # Should show: srwxr-xr-x (755)
   ```

3. **DOCKER_HOST set incorrectly**:
   ```bash
   # On macOS with Docker Desktop, DOCKER_HOST should NOT be set
   echo $DOCKER_HOST

   # If set, unset it
   unset DOCKER_HOST

   # Remove from shell config
   grep -n DOCKER_HOST ~/.zshrc ~/.bash_profile
   # Edit and remove the export statement
   ```

**Note**: On macOS, world-readable permissions (0755) are **safe and normal** because the socket is in your home directory, protected by macOS file system permissions.

### Linux

**Check current permissions**:
```bash
ls -la /var/run/docker.sock
# Should show: srw-rw---- (660) with group 'docker'
```

**Fix permissions** (if world-readable or world-writable):
```bash
# Set correct permissions
sudo chmod 660 /var/run/docker.sock

# Ensure correct group ownership
sudo chown root:docker /var/run/docker.sock
```

**Verify**:
```bash
ls -la /var/run/docker.sock
# Should show: srw-rw---- 1 root docker ... /var/run/docker.sock
```

**Add users to docker group** (instead of opening permissions):
```bash
# Add user to docker group
sudo usermod -aG docker $USER

# Log out and back in for changes to take effect
# Or run:
newgrp docker
```

**Verify group membership**:
```bash
groups
# Should include 'docker'
```

### Common: Fix orphaned DOCKER_HOST

**Check if DOCKER_HOST is set**:
```bash
echo $DOCKER_HOST
```

**If it points to a non-existent socket**:
```bash
# Unset it
unset DOCKER_HOST

# Remove from shell configuration
grep -n DOCKER_HOST ~/.bashrc ~/.zshrc ~/.profile
# Edit the file and remove the export statement
```

**For remote Docker hosts**:
```bash
# If you intentionally use a remote Docker host, verify it's accessible
docker -H unix:///var/run/docker.sock ps

# Or use TCP
docker -H tcp://remote-host:2376 ps
```

### Linux: Restart Docker daemon

```bash
# Restart Docker daemon to fix socket permissions
sudo systemctl restart docker

# Check Docker service status
sudo systemctl status docker

# Ensure Docker starts on boot
sudo systemctl enable docker
```

### Security best practices

#### Linux-specific

1. **Never make the socket world-accessible on Linux**:
   ```bash
   # NEVER do this on Linux:
   sudo chmod 666 /var/run/docker.sock  # ❌ DANGEROUS - grants root access
   ```

2. **Audit Docker socket access**:
   ```bash
   # See who's in the docker group
   getent group docker

   # Check for suspicious processes accessing the socket
   sudo lsof /var/run/docker.sock
   ```

3. **Use rootless Docker** for better security:
   ```bash
   # Install rootless Docker (Linux only)
   curl -fsSL https://get.docker.com/rootless | sh
   ```

4. **Monitor socket access with auditd**:
   ```bash
   # Use auditd to monitor socket access (Linux only)
   sudo auditctl -w /var/run/docker.sock -p war -k docker_socket
   ```

#### All platforms

1. **Use Docker contexts** for remote hosts:
   ```bash
   # Instead of DOCKER_HOST, use contexts
   docker context create remote --docker "host=tcp://remote:2376"
   docker context use remote
   ```

2. **Use Docker socket proxy** for limited access:
   ```bash
   # Run a proxy that restricts which Docker API calls are allowed
   docker run -v /var/run/docker.sock:/var/run/docker.sock \
     -p 2375:2375 \
     tecnativa/docker-socket-proxy
   ```

3. **Understand platform differences**:
   - **Linux**: Socket in `/var/run/` provides root access - must be restricted to `docker` group
   - **macOS**: Socket in user's home directory - protected by macOS file permissions
   - **Windows**: Uses named pipes - different security model

