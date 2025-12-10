# Snapshot Dependency

## What this is

This signal detects snapshot dependencies in Maven `pom.xml` files. Snapshot versions (ending in `-SNAPSHOT`) are development versions that can change at any time, making builds non-reproducible and potentially unstable.

## Why this matters

**Build Reproducibility**:
- **Non-deterministic builds**: Snapshot dependencies can change between builds, causing different behavior
- **CI/CD failures**: Builds may pass locally but fail in CI, or vice versa
- **Debugging difficulty**: Hard to reproduce bugs when dependencies keep changing
- **Version confusion**: Unclear which actual version of code is running

**Security & Supply Chain**:
- **Dependency confusion**: Attackers can publish malicious snapshot versions
- **Unvetted code**: Snapshot versions may contain untested or malicious code
- **No audit trail**: Can't track which exact version was used in a build
- **Compliance issues**: Can't prove which code version was deployed

**Production Risk**:
- **Unexpected behavior**: Snapshot dependencies can introduce breaking changes without warning
- **Rollback difficulty**: Can't reliably roll back to a previous build
- **Support issues**: Can't reproduce customer issues if dependencies have changed

## How to remediate

### Replace snapshots with release versions

**Find snapshot dependencies**:
```bash
# Search for SNAPSHOT in pom.xml
grep -n "SNAPSHOT" pom.xml

# Or use Maven to show dependencies
mvn dependency:tree | grep SNAPSHOT
```

**Replace with release versions**:
```xml
<!-- Before (bad) -->
<dependency>
    <groupId>com.example</groupId>
    <artifactId>my-library</artifactId>
    <version>1.0.0-SNAPSHOT</version>
</dependency>

<!-- After (good) -->
<dependency>
    <groupId>com.example</groupId>
    <artifactId>my-library</artifactId>
    <version>1.0.0</version>
</dependency>
```

**Update to latest release**:
```bash
# Check for latest release version
mvn versions:display-dependency-updates

# Update to latest release
mvn versions:use-latest-releases
```

### Remove snapshot repositories

**Remove snapshot repositories from pom.xml**:
```xml
<!-- Remove this section -->
<repositories>
    <repository>
        <id>snapshots</id>
        <url>https://oss.sonatype.org/content/repositories/snapshots</url>
        <snapshots>
            <enabled>true</enabled>
        </snapshots>
    </repository>
</repositories>
```

**Or disable snapshots**:
```xml
<!-- If you must keep the repository, disable snapshots -->
<repositories>
    <repository>
        <id>central</id>
        <url>https://repo.maven.apache.org/maven2</url>
        <snapshots>
            <enabled>false</enabled>
        </snapshots>
    </repository>
</repositories>
```

### For internal dependencies

**If you control the dependency, release it**:
```bash
# Release the snapshot version
mvn release:prepare
mvn release:perform

# This creates a proper release version (e.g., 1.0.0)
# And updates to next snapshot (e.g., 1.0.1-SNAPSHOT)
```

**Use semantic versioning**:
```xml
<!-- Release versions follow semantic versioning -->
<version>1.0.0</version>  <!-- Major.Minor.Patch -->
<version>1.1.0</version>  <!-- New features -->
<version>2.0.0</version>  <!-- Breaking changes -->
```

### Temporary workaround (not recommended)

**If you must use snapshots temporarily**:
```xml
<!-- Pin to specific snapshot timestamp -->
<dependency>
    <groupId>com.example</groupId>
    <artifactId>my-library</artifactId>
    <version>1.0.0-20240101.120000-1</version>  <!-- Specific snapshot -->
</dependency>
```

**Or use dependency management to lock versions**:
```xml
<dependencyManagement>
    <dependencies>
        <dependency>
            <groupId>com.example</groupId>
            <artifactId>my-library</artifactId>
            <version>1.0.0-SNAPSHOT</version>
        </dependency>
    </dependencies>
</dependencyManagement>
```

### Gradle equivalent

**For Gradle projects**:
```groovy
// build.gradle

// Bad - snapshot dependency
dependencies {
    implementation 'com.example:my-library:1.0.0-SNAPSHOT'
}

// Good - release version
dependencies {
    implementation 'com.example:my-library:1.0.0'
}

// Disable snapshot repositories
repositories {
    mavenCentral()
    // Remove or disable snapshot repos
    // maven { url 'https://oss.sonatype.org/content/repositories/snapshots' }
}
```

### Best practices

1. **Never use snapshots in production**:
   ```xml
   <!-- Development only -->
   <profiles>
       <profile>
           <id>dev</id>
           <dependencies>
               <dependency>
                   <groupId>com.example</groupId>
                   <artifactId>my-library</artifactId>
                   <version>1.0.0-SNAPSHOT</version>
               </dependency>
           </dependencies>
       </profile>
   </profiles>
   ```

2. **Use release versions** for all dependencies:
   ```bash
   # Check for snapshots before releasing
   mvn dependency:tree | grep SNAPSHOT
   # Should return nothing
   ```

3. **Automate dependency updates**:
   ```bash
   # Use Dependabot, Renovate, or similar
   # To keep dependencies up to date with release versions
   ```

4. **Lock dependency versions**:
   ```bash
   # Use Maven Enforcer Plugin
   mvn enforcer:enforce
   ```

5. **Add CI check** to prevent snapshots:
   ```yaml
   # .github/workflows/ci.yml
   - name: Check for snapshot dependencies
     run: |
       if grep -q "SNAPSHOT" pom.xml; then
         echo "Error: Snapshot dependencies found"
         exit 1
       fi
   ```

6. **Use dependency management**:
   ```xml
   <!-- Centralize version management -->
   <dependencyManagement>
       <dependencies>
           <!-- All versions defined here -->
       </dependencies>
   </dependencyManagement>
   ```

7. **Document dependency policy**:
   ```markdown
   # Dependency Policy
   - Only use release versions in main branch
   - Snapshots allowed in feature branches for development
   - All dependencies must be from trusted repositories
   - Regular dependency updates via Dependabot
   ```

### Maven Enforcer Plugin

**Prevent snapshot dependencies**:
```xml
<build>
    <plugins>
        <plugin>
            <groupId>org.apache.maven.plugins</groupId>
            <artifactId>maven-enforcer-plugin</artifactId>
            <version>3.0.0</version>
            <executions>
                <execution>
                    <id>enforce-no-snapshots</id>
                    <goals>
                        <goal>enforce</goal>
                    </goals>
                    <configuration>
                        <rules>
                            <requireReleaseDeps>
                                <message>No Snapshots Allowed!</message>
                            </requireReleaseDeps>
                        </rules>
                        <fail>true</fail>
                    </configuration>
                </execution>
            </executions>
        </plugin>
    </plugins>
</build>
```

**Run enforcer**:
```bash
mvn enforcer:enforce
```

### When snapshots are acceptable

Snapshots are only acceptable in:
- **Local development** on feature branches
- **Integration testing** between unreleased components
- **Temporary workarounds** with a plan to replace with release

Even then, document why and when they'll be removed.


## Disabling This Signal

To disable this signal, set the environment variable:
```
export DASHLIGHTS_DISABLE_SNAPSHOT_DEPENDENCY=1
```

To disable permanently, add the above line to your shell configuration file (`~/.zshrc`, `~/.bashrc`, etc.).
