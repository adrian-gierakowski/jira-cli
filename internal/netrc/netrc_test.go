package netrc

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "jira-cli-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Setenv("HOME", tempDir)

	netrcPath := filepath.Join(tempDir, ".netrc")

	// Test creating a new .netrc file
	err = Update("test.jira.com", "test-user", "test-token", false)
	assert.NoError(t, err)

	data, err := os.ReadFile(netrcPath)
	assert.NoError(t, err)

	expected := "machine test.jira.com\n  login test-user\n  password test-token\n"
	assert.Equal(t, expected, string(data))

	// Test adding a new entry to an existing .netrc file
	err = Update("another.jira.com", "another-user", "another-token", false)
	assert.NoError(t, err)

	data, err = os.ReadFile(netrcPath)
	assert.NoError(t, err)

	expected = "machine test.jira.com\n  login test-user\n  password test-token\nmachine another.jira.com\n  login another-user\n  password another-token\n"
	assert.Equal(t, expected, string(data))

	// Test updating an existing entry with force=true
	err = Update("test.jira.com", "updated-user", "updated-token", true)
	assert.NoError(t, err)

	data, err = os.ReadFile(netrcPath)
	assert.NoError(t, err)

	expected = "machine test.jira.com\n  login updated-user\n  password updated-token\nmachine another.jira.com\n  login another-user\n  password another-token\n"
	assert.Equal(t, expected, string(data))

	// Test updating an existing entry with force=false
	err = Update("test.jira.com", "should-fail", "should-fail", false)
	assert.Error(t, err)
}
