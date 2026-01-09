package init

import (
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"

	"github.com/ankitpokhrel/jira-cli/api"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

func TestUpdateNetrc(t *testing.T) {
	defer gock.Off()
	gock.InterceptClient(http.DefaultClient)

	// Inject a client that uses http.DefaultTransport which gock intercepts
	api.SetJiraClient(jira.NewClient(jira.Config{
		Server: "https://test.jira.com",
		Login:  "test-user",
	}, jira.WithTransport(http.DefaultClient.Transport)))
	defer api.SetJiraClient(nil)

	gock.New("https://test.jira.com").
		Get("/rest/api/2/myself").
		Reply(200).
		JSON(map[string]string{"key": "test-user", "name": "test-user"})

	gock.New("https://test.jira.com").
		Get("/rest/api/2/project").
		Reply(200).
		JSON([]map[string]string{{"key": "TEST", "name": "TEST"}})

	gock.New("https://test.jira.com").
		Get("/rest/agile/1.0/board").
		Reply(200).
		JSON(map[string][]map[string]interface{}{"values": {{"id": 1, "name": "TEST"}}})

	gock.New("https://test.jira.com").
		Get("/rest/api/2/issue/createmeta").
		Reply(200).
		JSON(map[string][]map[string]interface{}{"projects": {{"key": "TEST", "issuetypes": []map[string]string{{"name": "Story"}}}}})

	gock.New("https://test.jira.com").
		Get("/rest/api/2/field").
		Reply(200).
		JSON([]map[string]string{})

	tempDir, err := ioutil.TempDir("", "jira-cli-test")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Setenv("HOME", tempDir)
	t.Setenv("JIRA_API_TOKEN", "test-token")

	netrcPath := filepath.Join(tempDir, ".netrc")


	cmd := NewCmdInit()
	cmd.SetArgs([]string{
		"--update-netrc",
		"--server", "https://test.jira.com",
		"--login", "test-user",
		"--force",
		"--installation", "cloud",
		"--project", "TEST",
		"--board", "TEST",
	})

	assert.NoError(t, cmd.Execute())

	data, err := ioutil.ReadFile(netrcPath)
	assert.NoError(t, err)

	expected := "machine test.jira.com\n  login test-user\n  password test-token\n"
	assert.Equal(t, expected, string(data))
}
