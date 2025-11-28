package signals

// GetAllSignals returns all security signal checks
// Includes checks that complete in <10ms total when run concurrently
func GetAllSignals() []Signal {
	return []Signal{
		// IAM signals (1-5)
		NewSSHKeysSignal(), // 6ms - File stat checks
		// NewSudoCachedSignal(),      // DISABLED: 12ms - Runs sudo command
		NewNakedCredentialsSignal(), // <1ms - Env var scan
		NewPrivilegedPathSignal(),   // <1ms - PATH parsing
		NewAWSAliasHijackSignal(),   // <1ms - File read and parse

		// OpSec signals (6-10)
		NewLDPreloadSignal(),       // <1ms - Env var check
		NewHistoryDisabledSignal(), // <1ms - Env var check
		NewProdPanicSignal(),       // 4ms - File reads (kubectl/aws config)
		NewProxyActiveSignal(),     // <1ms - Env var check
		NewPermissiveUmaskSignal(), // <1ms - Single syscall

		// Repository hygiene signals (11-15)
		NewEnvNotIgnoredSignal(),       // 4ms - Reads .gitignore
		NewRootOwnedHomeSignal(),       // 5ms - File stat checks
		NewWorldWritableConfigSignal(), // 5ms - File stat checks
		NewUntrackedCryptoKeysSignal(), // 5ms - Directory scan

		// System health signals (16-20)
		NewDiskSpaceSignal(),       // 4ms - Syscall
		NewRebootPendingSignal(),   // 5ms - File stat
		NewZombieProcessesSignal(), // 5ms - /proc scan
		NewDockerSocketSignal(),    // 5ms - File stat
	}
}
