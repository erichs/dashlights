# Docker Socket Exposure

## What this is

This signal detects two Docker-related security issues:

1. **Overly permissive permissions** on the Docker socket (`/var/run/docker.sock`) - when the socket is world-readable or world-writable
2. **Orphaned DOCKER_HOST** environment variable - when `DOCKER_HOST` points to a non-existent socket, causing Docker commands to hang

## Why this matters

**Security Risk - Permissive Socket**:
- **Root-equivalent access**: Anyone who can access the Docker socket has effective root access to the host system
- **Container escape**: Attackers can create privileged containers and escape to the host
- **Data exfiltration**: Access to all container filesystems and volumes
- **Credential theft**: Can extract secrets from running containers

**Example attack**:
```bash
# If socket is world-accessible, any user can:
docker run -v /:/host -it ubuntu chroot /host /bin/bash
# Now they have root access to the host filesystem
```

**Operational Risk - Orphaned DOCKER_HOST**:
- **Hanging commands**: Docker commands will hang indefinitely waiting for the non-existent socket
- **CI/CD failures**: Build pipelines fail with timeout errors
- **Confusing errors**: Users get cryptic "cannot connect to Docker daemon" errors

## How to remediate

### Fix Docker socket permissions

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

### Add users to docker group (instead of opening permissions)

**Add your user to the docker group**:
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

### Fix orphaned DOCKER_HOST

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

### Platform-specific fixes

**macOS (Docker Desktop)**:
```bash
# Docker Desktop manages the socket automatically
# If you have issues, restart Docker Desktop

# Check Docker Desktop is running
docker ps

# If DOCKER_HOST is set, unset it (not needed on macOS)
unset DOCKER_HOST
```

**Linux (systemd)**:
```bash
# Restart Docker daemon to fix socket permissions
sudo systemctl restart docker

# Check Docker service status
sudo systemctl status docker

# Ensure Docker starts on boot
sudo systemctl enable docker
```

**Windows (Docker Desktop)**:
```powershell
# Docker Desktop uses named pipes, not Unix sockets
# If you have DOCKER_HOST set, remove it

# Check environment variable
$env:DOCKER_HOST

# Remove it
Remove-Item Env:DOCKER_HOST
```

### Security best practices

1. **Never make the socket world-accessible**:
   ```bash
   # NEVER do this:
   sudo chmod 666 /var/run/docker.sock  # ❌ DANGEROUS
   ```

2. **Use Docker contexts** for remote hosts:
   ```bash
   # Instead of DOCKER_HOST, use contexts
   docker context create remote --docker "host=tcp://remote:2376"
   docker context use remote
   ```

3. **Use rootless Docker** for better security:
   ```bash
   # Install rootless Docker
   curl -fsSL https://get.docker.com/rootless | sh
   ```

4. **Audit Docker socket access**:
   ```bash
   # See who's in the docker group
   getent group docker
   
   # Check for suspicious processes accessing the socket
   sudo lsof /var/run/docker.sock
   ```

5. **Use Docker socket proxy** for limited access:
   ```bash
   # Run a proxy that restricts which Docker API calls are allowed
   docker run -v /var/run/docker.sock:/var/run/docker.sock \
     -p 2375:2375 \
     tecnativa/docker-socket-proxy
   ```

6. **Monitor socket access**:
   ```bash
   # Use auditd to monitor socket access
   sudo auditctl -w /var/run/docker.sock -p war -k docker_socket
   ```

