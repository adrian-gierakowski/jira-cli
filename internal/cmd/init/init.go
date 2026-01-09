package init

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ankitpokhrel/jira-cli/internal/cmdutil"
	jiraConfig "github.com/ankitpokhrel/jira-cli/internal/config"
	"github.com/ankitpokhrel/jira-cli/internal/netrc"
	"github.com/ankitpokhrel/jira-cli/internal/query"
	"github.com/ankitpokhrel/jira-cli/pkg/jira"
)

type initParams struct {
	installation string
	server       string
	login        string
	authType     string
	project      string
	board        string
	force        bool
	insecure     bool
	updateNetrc  bool
}

// NewCmdInit is an init command.
func NewCmdInit() *cobra.Command {
	cmd := cobra.Command{
		Use:     "init",
		Short:   "Init initializes jira config",
		Long:    "Init initializes jira configuration required for the tool to work properly.",
		Aliases: []string{"initialize", "configure", "config", "setup"},
		Run:     initialize,
	}

	cmd.Flags().SortFlags = false

	cmd.Flags().String("installation", "", "Is this a 'cloud' or 'local' jira installation?")
	cmd.Flags().String("server", "", "Link to your jira server")
	cmd.Flags().String("login", "", "Jira login username or email based on your setup")
	cmd.Flags().String("auth-type", "", "Authentication type can be basic, bearer or mtls")
	cmd.Flags().String("project", "", "Your default project key")
	cmd.Flags().String("board", "", "Name of your default board in the project")
	cmd.Flags().Bool("force", false, "Forcefully override existing config if it exists")
	cmd.Flags().Bool("insecure", false, `If set, the tool will skip TLS certificate verification.
This can be useful if your server is using self-signed certificates.`)
	cmd.Flags().Bool("update-netrc", false, `If set, the tool will update ~/.netrc with your jira credentials.
This is useful if you are using tools like git-remote-helpers.`)

	return &cmd
}

func parseFlags(flags query.FlagParser) *initParams {
	installation, err := flags.GetString("installation")
	cmdutil.ExitIfError(err)

	server, err := flags.GetString("server")
	cmdutil.ExitIfError(err)

	login, err := flags.GetString("login")
	cmdutil.ExitIfError(err)

	authType, err := flags.GetString("auth-type")
	cmdutil.ExitIfError(err)

	// If auth type is not provided, check if it's set in env.
	if authType == "" {
		authType = os.Getenv("JIRA_AUTH_TYPE")
	}
	authType = strings.ToLower(authType)

	project, err := flags.GetString("project")
	cmdutil.ExitIfError(err)

	board, err := flags.GetString("board")
	cmdutil.ExitIfError(err)

	force, err := flags.GetBool("force")
	cmdutil.ExitIfError(err)

	insecure, err := flags.GetBool("insecure")
	cmdutil.ExitIfError(err)

	updateNetrc, err := flags.GetBool("update-netrc")
	cmdutil.ExitIfError(err)

	return &initParams{
		installation: installation,
		server:       server,
		login:        login,
		authType:     authType,
		project:      project,
		board:        board,
		force:        force,
		insecure:     insecure,
		updateNetrc:  updateNetrc,
	}
}

func fetchTenantInfo(server string) (*jira.TenantInfo, error) {
	server = strings.TrimSuffix(server, "/")
	req, err := http.NewRequest(http.MethodGet, server+"/_edge/tenant_info", nil)
	if err != nil {
		return nil, err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, nil // It's okay if this fails.
	}

	var info jira.TenantInfo
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return nil, nil // Also okay if this fails.
	}
	if info.CloudID == "" {
		return nil, nil
	}

	return &info, nil
}

func initialize(cmd *cobra.Command, _ []string) {
	params := parseFlags(cmd.Flags())

	var browseServer string
	if strings.EqualFold(params.installation, jira.InstallationTypeCloud) {
		if info, err := fetchTenantInfo(params.server); err == nil && info != nil {
			browseServer = params.server
			params.server = "https://api.atlassian.com/ex/jira/" + info.CloudID
		}
	}

	c := jiraConfig.NewJiraCLIConfigGenerator(
		&jiraConfig.JiraCLIConfig{
			Installation: strings.ToLower(params.installation),
			Server:       params.server,
			BrowseServer: browseServer,
			Login:        params.login,
			AuthType:     params.authType,
			Project:      params.project,
			Board:        params.board,
			Force:        params.force,
			Insecure:     params.insecure,
		},
	)

	if params.insecure {
		cmdutil.Warn(`You are using --insecure option. In this mode, the client will NOT verify
server's certificate chain and host name in requests to the jira server.`)
		fmt.Println()
	}

	file, err := c.Generate()
	if err != nil {
		if e, ok := err.(*jira.ErrUnexpectedResponse); ok {
			fmt.Println()
			cmdutil.Failed("Received unexpected response '%s' from jira. Please try again.", e.Status)
		} else {
			switch err {
			case jiraConfig.ErrSkip:
				cmdutil.Success("Skipping config generation. Current config: %s", viper.ConfigFileUsed())
			case jiraConfig.ErrUnexpectedResponseFormat:
				fmt.Println()
				cmdutil.Failed("Got response in unexpected format when fetching metadata. Please try again.")
			default:
				fmt.Println()
				cmdutil.Failed("Unable to generate configuration: %s", err.Error())
			}
		}
		os.Exit(1)
	}

	cfg := c.GetConfig()

	if params.updateNetrc {
		if err := updateNetrc(cfg.Server, cfg.Login, params.force); err != nil {
			cmdutil.Failed("Unable to update .netrc file: %s", err.Error())
			os.Exit(1)
		}
		cmdutil.Success("Successfully updated .netrc file")
	}

	cmdutil.Success("Configuration generated: %s", file)
}

func updateNetrc(server, login string, force bool) error {
	apiToken := os.Getenv("JIRA_API_TOKEN")
	if apiToken == "" {
		return fmt.Errorf("JIRA_API_TOKEN environment variable not set")
	}

	u, err := url.Parse(server)
	if err != nil {
		return fmt.Errorf("invalid server URL: %w", err)
	}

	return netrc.Update(u.Hostname(), login, apiToken, force)
}
