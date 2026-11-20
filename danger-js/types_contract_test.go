package dangerJs

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

// testdata/github_pr_dsl.json is the DSL file danger-js 14.1.0 wrote for
// danger/golang#46, captured with
//
//	danger pr https://github.com/danger/golang/pull/46 --process <script> --passURLForDSL
//
// with settings.github.accessToken redacted. It is the exact payload the
// runner reads, so a danger-js release that renames or retypes a field fails
// here rather than silently producing an empty DSL.
func loadGitHubFixture(t *testing.T) DSL {
	t.Helper()
	b, err := os.ReadFile("testdata/github_pr_dsl.json")
	require.NoError(t, err)

	var data struct {
		Danger DSLData `json:"danger"`
	}
	require.NoError(t, json.Unmarshal(b, &data))
	return data.Danger.ToInterface()
}

func TestGitHubDSLContract_Git(t *testing.T) {
	dsl := loadGitHubFixture(t)

	require.Equal(t, []FilePath{".github/workflows/test.yml"}, dsl.Git.ModifiedFiles())
	require.Empty(t, dsl.Git.CreatedFiles())
	require.Empty(t, dsl.Git.DeletedFiles())

	commits := dsl.Git.Commits()
	require.Len(t, commits, 1)
	require.Equal(t, "e14e0e3f3323a2edcb555d4688a3a161b74b95cc", commits[0].SHA)
	require.Equal(t, "renovate[bot]", commits[0].Author.Name)
	require.Equal(t, "renovate: Update dependency node to v26", commits[0].Message)
	require.Equal(t, []string{"4a1306d194a6b165e888f41b2b87d4bf7e18e0ed"}, commits[0].Parents)
}

func TestGitHubDSLContract_GitHub(t *testing.T) {
	dsl := loadGitHubFixture(t)
	gh := dsl.GitHub

	pr := gh.PR()
	require.Equal(t, 46, pr.Number)
	require.Equal(t, "open", pr.State)
	require.Equal(t, "renovate: Update dependency node to v26", pr.Title)
	require.Contains(t, pr.Body, "This PR contains the following updates")
	require.Equal(t, "renovate[bot]", pr.User.Login)
	require.Equal(t, "Bot", pr.User.Type)
	require.Equal(t, "renovate/node-26.x", pr.Head.Ref)
	require.Equal(t, "main", pr.Base.Ref)
	require.Equal(t, "danger/golang", pr.Base.Repo.FullName)
	require.Equal(t, 1, pr.ChangedFiles)
	require.Equal(t, 1, pr.Commits)
	require.False(t, pr.CreatedAt.IsZero())
	require.True(t, pr.MergedAt.IsZero(), "null merged_at must decode to the zero time")

	require.Equal(t, GitHubAPIPR{Owner: "danger", Repo: "golang", Number: 46}, gh.ThisPR())

	labels := gh.Issue().Labels
	require.Len(t, labels, 1)
	require.Equal(t, "bot::renovate", labels[0].Name)

	commits := gh.Commits()
	require.Len(t, commits, 1)
	require.Equal(t, "e14e0e3f3323a2edcb555d4688a3a161b74b95cc", commits[0].SHA)
	require.Equal(t, "renovate: Update dependency node to v26", commits[0].Commit.Message)

	reviews := gh.Reviews()
	require.Len(t, reviews, 1)
	require.Equal(t, "APPROVED", reviews[0].State)
	require.Equal(t, "e14e0e3f3323a2edcb555d4688a3a161b74b95cc", reviews[0].CommitID)

	require.Empty(t, gh.RequestedReviewers().Users)
	require.Empty(t, gh.RequestedReviewers().Teams)
}

func TestGitHubDSLContract_Settings(t *testing.T) {
	dsl := loadGitHubFixture(t)

	require.Equal(t, "REDACTED", dsl.Settings.GitHubAccessToken())
	require.Equal(t, "https://api.github.com", dsl.Settings.GitHubBaseURL())
	require.Equal(t, map[string]any{}, dsl.Settings.GitHubAdditionalHeaders())
	require.True(t, dsl.Settings.CLIArgs().TextOnly)
}

// TestCLIArgsDecoding covers the cliArgs danger-js sends for each flag. The
// JSON shapes are copied from a danger-js 14.1.0 run with
// `--verbose --base main -d build/ci/dangerfile.go --id myid`.
func TestCLIArgsDecoding(t *testing.T) {
	cc := []struct {
		name string
		json string
		exp  CLIArgs
	}{
		{
			name: "defaults",
			json: `{"textOnly":true}`,
			exp:  CLIArgs{TextOnly: true},
		},
		{
			name: "dangerfile and id",
			json: `{"dangerfile":"build/ci/dangerfile.go","id":"myid"}`,
			exp:  CLIArgs{Dangerfile: "build/ci/dangerfile.go", ID: "myid"},
		},
		{
			name: "base",
			json: `{"base":"main"}`,
			exp:  CLIArgs{Base: "main"},
		},
		{
			name: "verbose",
			json: `{"verbose":true}`,
			exp:  CLIArgs{Verbose: true},
		},
		{
			name: "staging",
			json: `{"staging":true}`,
			exp:  CLIArgs{Staging: true},
		},
	}
	for _, c := range cc {
		t.Run(c.name, func(t *testing.T) {
			var s settingsImpl
			require.NoError(t, json.Unmarshal([]byte(`{"cliArgs":`+c.json+`}`), &s))
			require.Equal(t, c.exp, s.CLIArgs())
		})
	}
}

// TestGitLabDSLDecoding is hand-written from danger-js's GitLabDSL type (no
// real GitLab payload has been captured yet), so it guards the Go side of the
// mapping only.
func TestGitLabDSLDecoding(t *testing.T) {
	payload := `{
		"git": {"modified_files": ["a.go"], "created_files": [], "deleted_files": [], "commits": []},
		"gitlab": {
			"metadata": {"repoSlug": "group/project", "pullRequestID": "7"},
			"mr": {"iid": 7, "title": "Add a thing", "state": "opened", "author": {"username": "alice"}},
			"commits": [{"id": "abc123", "title": "Add a thing"}],
			"approvals": {"approvals_required": 2, "approvals_left": 1}
		},
		"settings": {"github": {}, "cliArgs": {}}
	}`

	var data DSLData
	require.NoError(t, json.Unmarshal([]byte(payload), &data))
	dsl := data.ToInterface()

	require.Equal(t, []FilePath{"a.go"}, dsl.Git.ModifiedFiles())
	require.Equal(t, RepoMetaData{RepoSlug: "group/project", PullRequestID: "7"}, dsl.GitLab.Metadata())
	require.Equal(t, int64(7), dsl.GitLab.MR().IID)
	require.Equal(t, "Add a thing", dsl.GitLab.MR().Title)
	require.Equal(t, "alice", dsl.GitLab.MR().Author.Username)
	require.Len(t, dsl.GitLab.Commits(), 1)
	require.Equal(t, "abc123", dsl.GitLab.Commits()[0].ID)
	require.Equal(t, 2, dsl.GitLab.Approvals().ApprovalsRequired)
	require.Equal(t, 1, dsl.GitLab.Approvals().ApprovalsLeft)
}
