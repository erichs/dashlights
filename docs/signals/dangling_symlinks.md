# Dangling Symlinks

## What this is

This signal detects symbolic links (symlinks) in the current directory that point to non-existent targets. A dangling symlink is a link that references a file or directory that has been deleted or moved, leaving the symlink pointing to nothing.

## Why this matters

**Build Failures**: Dangling symlinks can cause:
- **Build errors** when tools try to access the linked files
- **Deployment failures** if the symlinks are part of your application
- **Test failures** when test suites encounter broken links

**Security Concerns**:
- **Time-of-check-time-of-use (TOCTOU) vulnerabilities**: If code checks for a symlink's existence but the target is missing, race conditions can occur
- **Privilege escalation**: Attackers can replace dangling symlinks with malicious targets
- **Information disclosure**: Dangling symlinks can reveal information about deleted files or directory structures

**Operational Issues**:
- **Confusing errors**: Applications may fail with cryptic "file not found" errors
- **Backup problems**: Backup tools may fail or skip directories with dangling symlinks
- **Disk space**: While symlinks themselves are tiny, they can prevent cleanup of directories

## How to remediate

### Find all dangling symlinks

**In current directory**:
```bash
find . -type l ! -exec test -e {} \; -print
```

**Recursively in project**:
```bash
find . -xtype l
```

**With details**:
```bash
find . -type l -ls | while read line; do
  link=$(echo "$line" | awk '{print $NF}')
  if [ ! -e "$link" ]; then
    echo "Dangling: $link"
  fi
done
```

### Fix the symlinks

**Option 1: Remove dangling symlinks**
```bash
# Remove all dangling symlinks in current directory
find . -xtype l -delete

# Or remove specific symlink
rm broken-link
```

**Option 2: Fix the symlink target**
```bash
# If you know where the file moved to
ln -sf /new/path/to/target symlink-name

# Or recreate the missing target
touch missing-file
```

**Option 3: Recreate from backup**
```bash
# If the target was accidentally deleted
git checkout HEAD -- path/to/deleted/file
```

### Platform-specific commands

**macOS**:
```bash
# Find dangling symlinks
find . -type l -exec sh -c 'test ! -e "$1"' _ {} \; -print

# Remove them
find . -type l -exec sh -c 'test ! -e "$1" && rm "$1"' _ {} \;
```

**Linux**:
```bash
# Find dangling symlinks
find . -xtype l

# Remove them
find . -xtype l -delete
```

**Windows (Git Bash/WSL)**:
```bash
# Find dangling symlinks
find . -type l ! -exec test -e {} \; -print

# Remove them (be careful!)
find . -type l ! -exec test -e {} \; -delete
```

### Prevent future dangling symlinks

1. **Use relative symlinks** when possible:
   ```bash
   # Instead of absolute paths
   ln -s /absolute/path/to/file link
   
   # Use relative paths
   ln -s ../relative/path/to/file link
   ```

2. **Check before deleting** files that might be symlink targets:
   ```bash
   # Find all symlinks pointing to a file before deleting it
   find . -lname "*filename*"
   ```

3. **Use version control** to track symlinks:
   ```bash
   # Git tracks symlinks, so you can restore them
   git ls-files -s | grep '^120000'
   ```

4. **Add a pre-commit hook** to detect dangling symlinks:
   ```bash
   #!/bin/bash
   # .git/hooks/pre-commit
   if find . -xtype l | grep -q .; then
     echo "Error: Dangling symlinks detected"
     find . -xtype l
     exit 1
   fi
   ```

