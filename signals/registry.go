package signals

// GetAllSignals returns all security signal checks
// Includes checks that complete in <10ms total when run concurrently
func GetAllSignals() []Signal {
	return []Signal{
		// IAM signals
		NewNakedCredentialsSignal(), // Env var check
		NewPrivilegedPathSignal(),   // Env var check
		NewAWSAliasHijackSignal(),   // File read (with permissions check)

		// OpSec signals
		NewLDPreloadSignal(),          // Env var check
		NewHistoryDisabledSignal(),    // Env var check
		NewProdPanicSignal(),          // File reads (kubectl/aws config)
		NewProxyActiveSignal(),        // Env var check
		NewPermissiveUmaskSignal(),    // Single syscall
		NewDockerSocketSignal(),       // File stat
		NewDebugEnabledSignal(),       // Env var check
		NewHistoryPermissionsSignal(), // File stat checks
		NewSSHAgentBloatSignal(),      // Unix socket query
		NewSSHKeysSignal(),            // File stat checks

		// Repository hygiene signals
		NewEnvNotIgnoredSignal(),       // Reads .gitignore
		NewRootOwnedHomeSignal(),       // File stat checks
		NewWorldWritableConfigSignal(), // File stat checks
		NewUntrackedCryptoKeysSignal(), // Directory scan
		NewGoReplaceSignal(),           // File read
		NewPyCachePollutionSignal(),    // Directory walk (performance)
		NewNpmrcTokensSignal(),         // File read
		NewCargoPathDepsSignal(),       // File read
		NewMissingInitPySignal(),       // Directory walk - checks for missing __init__.py
		NewSnapshotDependencySignal(),  // Optimized .git file reads (was 4.5ms with shell, now ~34μs)

		// System health signals
		NewDiskSpaceSignal(),        // Syscall
		NewRebootPendingSignal(),    // File stat
		NewZombieProcessesSignal(),  // /proc scan
		NewDanglingSymlinksSignal(), // Directory scan
		NewTimeDriftSignal(),        // File create/stat/delete

		// InfraSec signals
		NewTerraformStateLocalSignal(), // File stat
		NewRootKubeContextSignal(),     // File read (.kube/config)
		NewDangerousTFVarSignal(),      // Env var check
	}
}
