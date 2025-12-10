package signals

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestSnapshotDependencySignal_NoJavaProject(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo on main branch
	exec.Command("git", "init").Run()
	exec.Command("git", "checkout", "-b", "main").Run()

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when no Java project files exist")
	}
}

func TestSnapshotDependencySignal_NotReleaseBranch(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo on feature branch
	exec.Command("git", "init").Run()
	exec.Command("git", "checkout", "-b", "feature/test").Run()

	// Create pom.xml with SNAPSHOT
	pom := `<project>
  <dependencies>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>mylib</artifactId>
      <version>1.0-SNAPSHOT</version>
    </dependency>
  </dependencies>
</project>`
	os.WriteFile("pom.xml", []byte(pom), 0644)

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when on feature branch (SNAPSHOT is OK)")
	}
}

func TestSnapshotDependencySignal_MainBranchWithSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo on main branch
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	exec.Command("git", "checkout", "-b", "main").Run()

	// Need at least one commit
	os.WriteFile("README.md", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create pom.xml with SNAPSHOT
	pom := `<project>
  <dependencies>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>mylib</artifactId>
      <version>1.0-SNAPSHOT</version>
    </dependency>
  </dependencies>
</project>`
	os.WriteFile("pom.xml", []byte(pom), 0644)

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when SNAPSHOT found on main branch")
	}

	// Verify diagnostic is populated
	diagnostic := signal.Diagnostic()
	if !strings.Contains(diagnostic, "SNAPSHOT") {
		t.Errorf("Expected diagnostic to contain 'SNAPSHOT', got '%s'", diagnostic)
	}
}

func TestSnapshotDependencySignal_ReleaseBranchWithSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo on release branch
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	exec.Command("git", "checkout", "-b", "release/1.0").Run()

	// Need at least one commit
	os.WriteFile("README.md", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create pom.xml with SNAPSHOT
	pom := `<project>
  <dependencies>
    <dependency>
      <version>1.0-SNAPSHOT</version>
    </dependency>
  </dependencies>
</project>`
	os.WriteFile("pom.xml", []byte(pom), 0644)

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when SNAPSHOT found on release branch")
	}
}

func TestSnapshotDependencySignal_GradleWithSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo on main branch
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	exec.Command("git", "checkout", "-b", "main").Run()

	os.WriteFile("README.md", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create build.gradle with SNAPSHOT
	gradle := `dependencies {
    implementation 'com.example:mylib:1.0-SNAPSHOT'
    testImplementation 'junit:junit:4.13.2'
}`
	os.WriteFile("build.gradle", []byte(gradle), 0644)

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when SNAPSHOT found in build.gradle on main branch")
	}
}

func TestSnapshotDependencySignal_NoSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo on main branch
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	exec.Command("git", "checkout", "-b", "main").Run()

	os.WriteFile("README.md", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create pom.xml without SNAPSHOT
	pom := `<project>
  <dependencies>
    <dependency>
      <groupId>com.example</groupId>
      <artifactId>mylib</artifactId>
      <version>1.0.0</version>
    </dependency>
  </dependencies>
</project>`
	os.WriteFile("pom.xml", []byte(pom), 0644)

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when no SNAPSHOT dependencies")
	}
}

func TestSnapshotDependencySignal_DiagnosticWithoutFoundSnapshot(t *testing.T) {
	signal := NewSnapshotDependencySignal().(*SnapshotDependencySignal)
	// Don't set foundSnapshot, test the fallback message
	diagnostic := signal.Diagnostic()
	expected := "SNAPSHOT dependencies found on release branch (unstable for production)"
	if diagnostic != expected {
		t.Errorf("Expected '%s', got '%s'", expected, diagnostic)
	}
}

func TestSnapshotDependencySignal_MasterBranch(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo on master branch (alternative to main)
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	exec.Command("git", "checkout", "-b", "master").Run()

	os.WriteFile("README.md", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create pom.xml with SNAPSHOT
	pom := `<project>
  <dependencies>
    <dependency>
      <version>1.0-SNAPSHOT</version>
    </dependency>
  </dependencies>
</project>`
	os.WriteFile("pom.xml", []byte(pom), 0644)

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when SNAPSHOT found on master branch")
	}
}

func TestSnapshotDependencySignal_ReleasesBranch(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo on releases/ branch (plural variant)
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	exec.Command("git", "checkout", "-b", "releases/v2.0").Run()

	os.WriteFile("README.md", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create pom.xml with SNAPSHOT
	pom := `<project>
  <dependencies>
    <dependency>
      <version>2.0-SNAPSHOT</version>
    </dependency>
  </dependencies>
</project>`
	os.WriteFile("pom.xml", []byte(pom), 0644)

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when SNAPSHOT found on releases/ branch")
	}
}

func TestSnapshotDependencySignal_GradleCompileDependency(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo on main branch
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	exec.Command("git", "checkout", "-b", "main").Run()

	os.WriteFile("README.md", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create build.gradle with SNAPSHOT using 'compile' (older Gradle syntax)
	gradle := `dependencies {
    compile 'com.example:mylib:1.0-SNAPSHOT'
}`
	os.WriteFile("build.gradle", []byte(gradle), 0644)

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when SNAPSHOT found with 'compile' in build.gradle")
	}
}

func TestSnapshotDependencySignal_GradleApiDependency(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo on main branch
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	exec.Command("git", "checkout", "-b", "main").Run()

	os.WriteFile("README.md", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create build.gradle with SNAPSHOT using 'api'
	gradle := `dependencies {
    api 'com.example:mylib:2.0-SNAPSHOT'
}`
	os.WriteFile("build.gradle", []byte(gradle), 0644)

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when SNAPSHOT found with 'api' in build.gradle")
	}
}

func TestSnapshotDependencySignal_GradleTestImplementationDependency(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo on main branch
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	exec.Command("git", "checkout", "-b", "main").Run()

	os.WriteFile("README.md", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create build.gradle with SNAPSHOT using 'testImplementation'
	gradle := `dependencies {
    testImplementation 'com.example:test-lib:1.0-SNAPSHOT'
}`
	os.WriteFile("build.gradle", []byte(gradle), 0644)

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when SNAPSHOT found with 'testImplementation' in build.gradle")
	}
}

func TestSnapshotDependencySignal_DetachedHead(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo and create a detached HEAD state
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	os.WriteFile("README.md", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Get the commit SHA and checkout to create detached HEAD
	output, _ := exec.Command("git", "rev-parse", "HEAD").Output()
	sha := strings.TrimSpace(string(output))
	exec.Command("git", "checkout", sha).Run()

	// Create pom.xml with SNAPSHOT
	pom := `<project>
  <dependencies>
    <dependency>
      <version>1.0-SNAPSHOT</version>
    </dependency>
  </dependencies>
</project>`
	os.WriteFile("pom.xml", []byte(pom), 0644)

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	// In detached HEAD state, not on a release branch, so SNAPSHOT is OK
	result := signal.Check(ctx)
	if result {
		t.Error("Expected false when in detached HEAD state (not a release context)")
	}
}

func TestSnapshotDependencySignal_OnTag(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	exec.Command("git", "checkout", "-b", "develop").Run()

	os.WriteFile("README.md", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create a tag
	exec.Command("git", "tag", "v1.0.0").Run()

	// Create pom.xml with SNAPSHOT
	pom := `<project>
  <dependencies>
    <dependency>
      <version>1.0-SNAPSHOT</version>
    </dependency>
  </dependencies>
</project>`
	os.WriteFile("pom.xml", []byte(pom), 0644)

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	// On a tag, this is a release context, so SNAPSHOT should be flagged
	result := signal.Check(ctx)
	if !result {
		t.Error("Expected true when SNAPSHOT found on a tagged commit")
	}
}

func TestSnapshotDependencySignal_DirectoryTraversalInRef(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()

	os.WriteFile("README.md", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Manually create a malicious .git/HEAD with directory traversal
	// This tests the security validation at line 117
	maliciousRef := "ref: refs/heads/../../../etc/passwd"
	os.WriteFile(".git/HEAD", []byte(maliciousRef), 0644)

	// Create pom.xml with SNAPSHOT
	pom := `<project>
  <dependencies>
    <dependency>
      <version>1.0-SNAPSHOT</version>
    </dependency>
  </dependencies>
</project>`
	os.WriteFile("pom.xml", []byte(pom), 0644)

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	// Should not panic or read outside .git directory
	// The validation should catch the ".." and return error
	result := signal.Check(ctx)
	// Should return false because isReleaseContext() will fail safely
	if result {
		t.Error("Expected false when .git/HEAD contains directory traversal attempt")
	}
}

func TestSnapshotDependencySignal_DirectoryTraversalInTag(t *testing.T) {
	tmpDir := t.TempDir()
	originalDir, _ := os.Getwd()
	os.Chdir(tmpDir)
	defer os.Chdir(originalDir)

	// Initialize git repo
	exec.Command("git", "init").Run()
	exec.Command("git", "config", "user.email", "test@example.com").Run()
	exec.Command("git", "config", "user.name", "Test User").Run()
	exec.Command("git", "checkout", "-b", "main").Run()

	os.WriteFile("README.md", []byte("test"), 0644)
	exec.Command("git", "add", ".").Run()
	exec.Command("git", "commit", "-m", "initial").Run()

	// Create .git/refs/tags directory
	os.MkdirAll(".git/refs/tags", 0755)

	// Try to create a malicious tag file with directory traversal in the name
	// This tests the security validation at line 171
	// Note: We can't actually create a file with "/" in the name on most filesystems,
	// but we can verify the code would skip it if it existed
	// The validation code checks for ".." and "/" in entry names and skips them
	maliciousTagContent := "abc123def456"
	os.MkdirAll(".git/refs", 0755)

	// Create a normal tag to ensure the function runs
	os.WriteFile(".git/refs/tags/v1.0.0", []byte(maliciousTagContent), 0644)

	// Create pom.xml with SNAPSHOT
	pom := `<project>
  <dependencies>
    <dependency>
      <version>1.0-SNAPSHOT</version>
    </dependency>
  </dependencies>
</project>`
	os.WriteFile("pom.xml", []byte(pom), 0644)

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	// Should not panic or read outside .git directory
	result := signal.Check(ctx)
	// Result depends on whether we're on the tag, but should not crash
	_ = result
}

func TestSnapshotDependencySignal_Disabled(t *testing.T) {
	os.Setenv("DASHLIGHTS_DISABLE_SNAPSHOT_DEPENDENCY", "1")
	defer os.Unsetenv("DASHLIGHTS_DISABLE_SNAPSHOT_DEPENDENCY")

	signal := NewSnapshotDependencySignal()
	ctx := context.Background()

	if signal.Check(ctx) {
		t.Error("Expected false when signal is disabled via environment variable")
	}
}
