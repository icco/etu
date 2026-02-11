package client

import "testing"

// setTestHome isolates config/cache paths to a temp directory.
// On Linux, os.UserConfigDir checks XDG_CONFIG_HOME before HOME,
// so we must clear it to ensure $HOME/.config is used.
func setTestHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv("XDG_CONFIG_HOME", "")
}
