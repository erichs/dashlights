package signals

import (
	"context"
	"os"
	"os/exec"
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
