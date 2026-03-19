package install

// OverrideGOOS temporarily overrides the runtimeGOOS variable used for
// platform detection. It returns a restore function that should be called
// (typically deferred) to restore the original value.
//
// This is only intended for use in tests to force the RC-file code path
// on Windows CI runners.
func OverrideGOOS(goos string) (restore func()) {
	old := runtimeGOOS
	runtimeGOOS = goos
	return func() { runtimeGOOS = old }
}
