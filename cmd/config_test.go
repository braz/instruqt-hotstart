package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// newFlags mirrors the persistent flags relevant to config resolution.
func newFlags() *pflag.FlagSet {
	f := pflag.NewFlagSet("test", pflag.ContinueOnError)
	f.String("team", "", "")
	f.String("api-key", "", "")
	f.String("config", "", "")
	return f
}

func writeConfig(t *testing.T, team string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("team: "+team+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestConfigPrecedence(t *testing.T) {
	cfgPath := writeConfig(t, "fileteam")

	t.Run("config file only", func(t *testing.T) {
		f := newFlags()
		_ = f.Set("config", cfgPath)
		vp := viper.New()
		configureViper(vp, f)
		if got := vp.GetString("team"); got != "fileteam" {
			t.Errorf("team = %q, want fileteam", got)
		}
	})

	t.Run("env overrides file", func(t *testing.T) {
		t.Setenv("INSTRUQT_TEAM", "envteam")
		f := newFlags()
		_ = f.Set("config", cfgPath)
		vp := viper.New()
		configureViper(vp, f)
		if got := vp.GetString("team"); got != "envteam" {
			t.Errorf("team = %q, want envteam", got)
		}
	})

	t.Run("flag overrides env and file", func(t *testing.T) {
		t.Setenv("INSTRUQT_TEAM", "envteam")
		f := newFlags()
		_ = f.Set("config", cfgPath)
		_ = f.Set("team", "flagteam")
		vp := viper.New()
		configureViper(vp, f)
		if got := vp.GetString("team"); got != "flagteam" {
			t.Errorf("team = %q, want flagteam", got)
		}
	})

	// Regression: a hyphenated config key must resolve from its underscored
	// env var (api-key -> INSTRUQT_API_KEY), not INSTRUQT_API-KEY.
	t.Run("hyphenated key from underscored env var", func(t *testing.T) {
		t.Setenv("INSTRUQT_API_KEY", "secret-key")
		f := newFlags()
		vp := viper.New()
		configureViper(vp, f)
		if got := vp.GetString("api-key"); got != "secret-key" {
			t.Errorf("api-key = %q, want secret-key", got)
		}
	})
}
