package signals

// GetAllSignals returns all security signal checks
// Includes checks that complete in <10ms total when run concurrently
func GetAllSignals() []Signal {
	return []Signal{
		// IAM signals
		// NewSudoCachedSignal(),      // DISABLED: 12ms - Runs sudo command
		NewNakedCredentialsSignal(), // <1ms - Env var scan
		NewPrivilegedPathSignal(),   // <1ms - PATH parsing
		NewAWSAliasHijackSignal(),   // <1ms - File read and parse

		// OpSec signals
		NewLDPreloadSignal(),          // <1ms - Env var check
		NewHistoryDisabledSignal(),    // <1ms - Env var check
		NewProdPanicSignal(),          // 4ms - File reads (kubectl/aws config)
		NewProxyActiveSignal(),        // <1ms - Env var check
		NewPermissiveUmaskSignal(),    // <1ms - Single syscall
		NewDockerSocketSignal(),       // 5ms - File stat
		NewDebugEnabledSignal(),       // <1ms - Env var check
		NewHistoryPermissionsSignal(), // <1ms - File stat checks
		NewSSHAgentBloatSignal(),      // <1ms - Unix socket query

		// Repository hygiene signals
		NewEnvNotIgnoredSignal(),       // 4ms - Reads .gitignore
		NewRootOwnedHomeSignal(),       // 5ms - File stat checks
		NewWorldWritableConfigSignal(), // 5ms - File stat checks
		NewUntrackedCryptoKeysSignal(), // 5ms - Directory scan

		// System health signals
		NewDiskSpaceSignal(),        // 4ms - Syscall
		NewRebootPendingSignal(),    // 5ms - File stat
		NewZombieProcessesSignal(),  // 5ms - /proc scan
		NewDanglingSymlinksSignal(), // <1ms - Directory scan
		NewTimeDriftSignal(),        // <1ms - File create/stat/delete
	}
}
