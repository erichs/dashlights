# Zombie Processes

## What this is

This signal detects zombie processes (defunct processes) on your system. A zombie process is a process that has completed execution but still has an entry in the process table because its parent process hasn't read its exit status yet.

While a few zombies are normal and harmless, many zombies indicate a bug in the parent process or a resource leak.

## Why this matters

**Resource Leaks**:
- **Process table exhaustion**: Too many zombies can fill the process table, preventing new processes from starting
- **PID exhaustion**: Each zombie consumes a process ID
- **Memory leaks**: Indicates parent process isn't properly cleaning up child processes
- **Application bugs**: Usually indicates a bug in process management

**Operational Impact**:
- **System instability**: Extreme zombie accumulation can make the system unstable
- **Cannot fork**: New processes fail with "Cannot fork" errors when process table is full
- **Monitoring alerts**: Zombie processes trigger monitoring alerts
- **Performance degradation**: Indicates underlying issues with application health

**What zombies look like**:
```bash
$ ps aux | grep defunct
user     12345  0.0  0.0      0     0 ?        Z    10:00   0:00 [process] <defunct>
```

## How to remediate

### Identify zombie processes

**Find zombie processes**:
```bash
# List all zombie processes
ps aux | grep 'Z'

# Or use ps with specific format
ps -eo pid,ppid,stat,comm | grep '^[0-9]* [0-9]* Z'

# Count zombies
ps aux | grep -c 'Z'

# Show zombie details
ps -eo pid,ppid,stat,comm,cmd | awk '$3 ~ /^Z/ {print}'
```

**Find parent of zombie processes**:
```bash
# Get parent PID of zombies
ps -eo pid,ppid,stat,comm | awk '$3 ~ /^Z/ {print "Zombie PID:", $1, "Parent PID:", $2, "Command:", $4}'

# See what the parent process is
ps -p <parent-pid> -o pid,ppid,comm,cmd
```

### Kill zombie processes (by fixing parent)

**You cannot kill zombies directly** - they're already dead. You must fix the parent process.

**Option 1: Restart the parent process**:
```bash
# Find parent PID
PARENT_PID=$(ps -eo pid,ppid,stat | awk '$3 ~ /^Z/ {print $2}' | head -1)

# Check what the parent is
ps -p $PARENT_PID -o comm,cmd

# Restart the parent (if it's a service)
sudo systemctl restart service-name

# Or kill and restart the parent
kill $PARENT_PID
# Parent will be restarted by init/systemd if it's a service
```

**Option 2: Send SIGCHLD to parent**:
```bash
# Tell parent to reap its zombie children
kill -s SIGCHLD $PARENT_PID

# Verify zombies are gone
ps aux | grep 'Z'
```

**Option 3: Kill parent process**:
```bash
# If parent is stuck and won't reap zombies
kill -9 $PARENT_PID

# Zombies will be adopted by init (PID 1) and cleaned up
# Check after a few seconds
sleep 5
ps aux | grep 'Z'
```

### Fix the root cause

**For your own applications**:

**In C/C++**:
```c
#include <signal.h>
#include <sys/wait.h>

// Set up signal handler to reap children
void sigchld_handler(int sig) {
    while (waitpid(-1, NULL, WNOHANG) > 0);
}

int main() {
    // Install signal handler
    signal(SIGCHLD, sigchld_handler);

    // Or use sigaction (better)
    struct sigaction sa;
    sa.sa_handler = sigchld_handler;
    sigemptyset(&sa.sa_mask);
    sa.sa_flags = SA_RESTART | SA_NOCLDSTOP;
    sigaction(SIGCHLD, &sa, NULL);

    // Your code that forks children...
}
```

**In Python**:
```python
import signal
import os

# Set up signal handler
def sigchld_handler(signum, frame):
    while True:
        try:
            pid, status = os.waitpid(-1, os.WNOHANG)
            if pid == 0:
                break
        except ChildProcessError:
            break

signal.signal(signal.SIGCHLD, sigchld_handler)

# Or use subprocess module which handles this automatically
import subprocess
subprocess.Popen(['command'], ...)
```

**In Go**:
```go
package main

import (
    "os/exec"
)

func main() {
    cmd := exec.Command("command")
    cmd.Start()

    // Always wait for child processes
    defer cmd.Wait()

    // Or use goroutine
    go func() {
        cmd.Wait()
    }()
}
```

**In Node.js**:
```javascript
const { spawn } = require('child_process');

const child = spawn('command', ['args']);

// Always handle exit event
child.on('exit', (code, signal) => {
    console.log(`Child exited with code ${code}`);
});

// Or use child.unref() for long-running children
child.unref();
```

### Prevent zombie accumulation

**Use process managers**:
```bash
# systemd (Linux)
# Automatically reaps zombies for services

# supervisord
[program:myapp]
command=/path/to/myapp
autorestart=true

# PM2 (Node.js)
pm2 start app.js

# Docker
# Docker automatically reaps zombies in containers
```

**Use init systems in containers**:
```dockerfile
# Use tini as init
FROM ubuntu
RUN apt-get update && apt-get install -y tini
ENTRYPOINT ["/usr/bin/tini", "--"]
CMD ["myapp"]

# Or use dumb-init
FROM ubuntu
RUN apt-get update && apt-get install -y dumb-init
ENTRYPOINT ["/usr/bin/dumb-init", "--"]
CMD ["myapp"]
```

### Monitor for zombies

**Add monitoring**:
```bash
# Check for zombies in monitoring script
ZOMBIE_COUNT=$(ps aux | grep -c 'Z')
if [ $ZOMBIE_COUNT -gt 10 ]; then
    echo "WARNING: $ZOMBIE_COUNT zombie processes detected"
    # Send alert
fi
```

**Add to cron**:
```bash
# Check every hour
0 * * * * /usr/local/bin/check-zombies.sh
```

**Use monitoring tools**:
- Nagios/Icinga: Check zombie process count
- Prometheus: `node_exporter` exposes zombie count
- Datadog/New Relic: Monitor process metrics

### Platform-specific considerations

**Linux**:
```bash
# systemd automatically reaps zombies for services
# Check service status
systemctl status service-name

# Restart service to clear zombies
sudo systemctl restart service-name
```

**macOS**:
```bash
# launchd handles zombie reaping
# Check for zombies
ps aux | grep 'Z'

# Restart service
sudo launchctl stop com.example.service
sudo launchctl start com.example.service
```

**Docker containers**:
```bash
# Use --init flag to use tini
docker run --init image

# Or in docker-compose.yml
services:
  app:
    image: myapp
    init: true
```

### Best practices

1. **Always wait for child processes**:
   ```bash
   # In shell scripts
   command &
   PID=$!
   wait $PID
   ```

2. **Use signal handlers** to reap children:
   ```bash
   # In shell scripts
   trap 'wait' EXIT
   ```

3. **Use process managers** instead of manual process management

4. **Use init systems** in containers (tini, dumb-init)

5. **Monitor zombie count** and alert if it exceeds threshold

6. **Fix bugs** in parent processes that don't reap children

7. **Use modern APIs** that handle process cleanup automatically:
   - Python: `subprocess` module
   - Node.js: `child_process` with event handlers
   - Go: `exec.Command` with `Wait()`

### Troubleshooting

**If zombies persist after killing parent**:
```bash
# Check if init (PID 1) is working
ps -p 1 -o comm,cmd

# Reboot if necessary (last resort)
sudo reboot
```

**If you can't create new processes**:
```bash
# Check process limit
ulimit -u

# Check current process count
ps aux | wc -l

# Kill zombie parents to free PIDs
# Find and kill problematic parent processes
```

**If zombies are from a critical service**:
```bash
# Don't kill the service immediately
# 1. Identify the bug in the service
# 2. Fix the bug
# 3. Deploy the fix
# 4. Restart the service during maintenance window
```

### When zombies are normal

A few zombies are normal and harmless:
- **Short-lived zombies**: Appear briefly between child exit and parent wait
- **During shutdown**: Processes being cleaned up
- **After crashes**: Parent crashed before reaping children

**Concern threshold**: More than 10-20 zombies is unusual and should be investigated.

### Emergency cleanup

**If system is unstable due to zombies**:
```bash
# Find all zombie parents
ps -eo pid,ppid,stat,comm | awk '$3 ~ /^Z/ {print $2}' | sort -u

# Kill all zombie parents (careful!)
ps -eo pid,ppid,stat,comm | awk '$3 ~ /^Z/ {print $2}' | sort -u | xargs kill

# If that doesn't work, force kill
ps -eo pid,ppid,stat,comm | awk '$3 ~ /^Z/ {print $2}' | sort -u | xargs kill -9

# Reboot if necessary
sudo reboot
```


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_ZOMBIE_PROCESSES=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
