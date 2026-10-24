package settings

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/ini.v1"
)

func TestPersistSecretKey_WritesKeyToFile(t *testing.T) {
	// Create a temp config file with an empty security section
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.ini")

	initialContent := `[global]
mode = development

[security]
secret_key =
`
	err := os.WriteFile(cfgPath, []byte(initialContent), 0644)
	assert.NoError(t, err)

	// Persist a new secret key
	testKey := "abc123def456"
	err = PersistSecretKey(cfgPath, testKey)
	assert.NoError(t, err)

	// Re-read the file and verify the key was written
	cfgFile, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, cfgPath)
	assert.NoError(t, err)

	actual := cfgFile.Section("security").Key("secret_key").String()
	assert.Equal(t, testKey, actual)

	if runtime.GOOS != "windows" {
		fileInfo, err := os.Stat(cfgPath)
		assert.NoError(t, err)
		assert.Equal(t, os.FileMode(0600), fileInfo.Mode().Perm())
	}
}

func TestPersistSecretKey_CreatesSecuritySection(t *testing.T) {
	// Config file with NO security section
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.ini")

	initialContent := `[global]
mode = development
`
	err := os.WriteFile(cfgPath, []byte(initialContent), 0644)
	assert.NoError(t, err)

	testKey := "newkey789"
	err = PersistSecretKey(cfgPath, testKey)
	assert.NoError(t, err)

	cfgFile, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, cfgPath)
	assert.NoError(t, err)

	actual := cfgFile.Section("security").Key("secret_key").String()
	assert.Equal(t, testKey, actual)
}

func TestPersistSecretKey_ErrorOnMissingFile(t *testing.T) {
	err := PersistSecretKey("/nonexistent/path/config.ini", "somekey")
	assert.Error(t, err)
}

func TestPersistSecretKey_PreservesOtherSettings(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, "test.ini")

	initialContent := `[global]
mode = production

[security]
secret_key = old_key
token_expired_time = 3600
`
	err := os.WriteFile(cfgPath, []byte(initialContent), 0644)
	assert.NoError(t, err)

	err = PersistSecretKey(cfgPath, "new_secret")
	assert.NoError(t, err)

	cfgFile, err := ini.LoadSources(ini.LoadOptions{
		IgnoreInlineComment: true,
	}, cfgPath)
	assert.NoError(t, err)

	// Secret key updated
	assert.Equal(t, "new_secret", cfgFile.Section("security").Key("secret_key").String())
	// Other settings preserved
	assert.Equal(t, "3600", cfgFile.Section("security").Key("token_expired_time").String())
	assert.Equal(t, "production", cfgFile.Section("global").Key("mode").String())
}

func TestGetWorkingPath_LegacyEnvFallback(t *testing.T) {
	legacyPath := t.TempDir()
	t.Setenv(ratelWorkDirEnvName, "")
	t.Setenv(legacyWorkDirEnvName, legacyPath)

	actual, err := getWorkingPath()
	assert.NoError(t, err)
	assert.Equal(t, legacyPath, actual)
}

func TestGetWorkingPath_PrefersCurrentEnv(t *testing.T) {
	currentPath := t.TempDir()
	legacyPath := t.TempDir()
	t.Setenv(ratelWorkDirEnvName, currentPath)
	t.Setenv(legacyWorkDirEnvName, legacyPath)

	actual, err := getWorkingPath()
	assert.NoError(t, err)
	assert.Equal(t, currentPath, actual)
}

func TestLoadSecurityConfiguration_EnvironmentSecretIsNotPersisted(t *testing.T) {
	t.Setenv("RATEL_SECURITY_SECRET_KEY", "environment-secret")

	cfgFile, err := ini.LoadSources(ini.LoadOptions{IgnoreInlineComment: true}, []byte(`[security]
secret_key =
`))
	assert.NoError(t, err)

	config := &Config{}
	err = loadSecurityConfiguration(config, cfgFile, "security")
	assert.NoError(t, err)
	assert.False(t, config.SecretKeyNoSet)
	assert.Equal(t, "environment-secret", config.SecretKey)
}
