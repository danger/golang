package dangerJs

import (
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type GitHub interface {
	GetIssue() GitHubIssue
	GetPR() GitHubPR
	GetThisPR() GitHubAPIPR
	GetCommits() []GitHubCommit
	GetReviews() []GitHubReview
	GetRequestedReviewers() GitHubReviewers
}

type GitLab interface {
	GetMetadata() RepoMetaData
	GetMR() GitLabMR
	GetCommits() []GitLabMRCommit
	GetApprovals() GitLabApproval
}

type Settings interface {
	GetGitHubAccessToken() string
	GetGitHubBaseURL() string
	GetGitHubAdditionalHeaders() any
	GetCLIArgs() CLIArgs
}

type Git interface {
	GetModifiedFiles() []FilePath
	GetCreatedFiles() []FilePath
	GetDeletedFiles() []FilePath
	GetCommits() []GitCommit
	DiffForFile(filePath string) (FileDiff, error)
	DiffForFileWithRefs(filePath, baseRef, headRef string) (FileDiff, error)
}

// DSL is the main Danger context, with all fields as interfaces for testability.
type DSL struct {
	Git      Git      `json:"git"`
	GitHub   GitHub   `json:"github,omitempty"`
	GitLab   GitLab   `json:"gitlab,omitempty"`
	Settings Settings `json:"settings"`
}

type FilePath = string

// gitImpl is the internal implementation of the Git interface
type gitImpl struct {
	ModifiedFiles []FilePath  `json:"modified_files"`
	CreatedFiles  []FilePath  `json:"created_files"`
	DeletedFiles  []FilePath  `json:"deleted_files"`
	Commits       []GitCommit `json:"commits"`
}

func (g gitImpl) GetModifiedFiles() []FilePath {
	return g.ModifiedFiles
}

func (g gitImpl) GetCreatedFiles() []FilePath {
	return g.CreatedFiles
}

func (g gitImpl) GetDeletedFiles() []FilePath {
	return g.DeletedFiles
}

func (g gitImpl) GetCommits() []GitCommit {
	return g.Commits
}

// FileDiff represents the changes in a file.
type FileDiff struct {
	AddedLines   []DiffLine
	RemovedLines []DiffLine
}

// DiffLine represents a single line in a file diff.
type DiffLine struct {
	Content string
	Line    int
}

// DiffForFile executes a git diff command for a specific file and parses its output.
// Uses HEAD^ and HEAD as the base and head references by default.
func (g gitImpl) DiffForFile(filePath string) (FileDiff, error) {
	return g.DiffForFileWithRefs(filePath, "HEAD^", "HEAD")
}

// DiffForFileWithRefs executes a git diff command for a specific file with configurable references.
func (g gitImpl) DiffForFileWithRefs(filePath, baseRef, headRef string) (FileDiff, error) {
	cmd := exec.Command("git", "diff", "--unified=0", baseRef, headRef, filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return FileDiff{}, err
	}

	return parseDiffContent(out.String()), nil
}

// parseDiffContent parses git diff output and extracts added and removed lines with line numbers
func parseDiffContent(diffContent string) FileDiff {
	var fileDiff FileDiff
	// Only match lines that start with + or - but not +++ or --- (file headers)
	addedRe := regexp.MustCompile(`^\+([^+].*|$)`)
	removedRe := regexp.MustCompile(`^-([^-].*|$)`)
	// Match hunk headers like @@ -1,3 +1,4 @@
	hunkRe := regexp.MustCompile(`^@@\s+-(\d+)(?:,(\d+))?\s+\+(\d+)(?:,(\d+))?\s+@@`)

	lines := strings.Split(diffContent, "\n")
	var currentRemovedLine, currentAddedLine int

	for _, line := range lines {
		// Check for hunk header to track line numbers
		if hunkMatches := hunkRe.FindStringSubmatch(line); len(hunkMatches) > 0 {
			// Parse starting line numbers from hunk header
			if removedStart, err := strconv.Atoi(hunkMatches[1]); err == nil {
				currentRemovedLine = removedStart
			}
			if addedStart, err := strconv.Atoi(hunkMatches[3]); err == nil {
				currentAddedLine = addedStart
			}
		} else if matches := addedRe.FindStringSubmatch(line); len(matches) > 1 {
			fileDiff.AddedLines = append(fileDiff.AddedLines, DiffLine{
				Content: matches[1],
				Line:    currentAddedLine,
			})
			currentAddedLine++
		} else if matches := removedRe.FindStringSubmatch(line); len(matches) > 1 {
			fileDiff.RemovedLines = append(fileDiff.RemovedLines, DiffLine{
				Content: matches[1],
				Line:    currentRemovedLine,
			})
			currentRemovedLine++
		}
	}

	return fileDiff
}

// settingsImpl is the internal implementation of the Settings interface
type settingsImpl struct {
	GitHub struct {
		AccessToken       string `json:"accessToken"`
		BaseURL           string `json:"baseURL"`
		AdditionalHeaders any    `json:"additionalHeaders"`
	} `json:"github"`
	CLIArgs CLIArgs `json:"cliArgs"`
}

// GetGitHubAccessToken returns the GitHub access token
func (s settingsImpl) GetGitHubAccessToken() string {
	return s.GitHub.AccessToken
}

func (s settingsImpl) GetGitHubBaseURL() string {
	return s.GitHub.BaseURL
}

func (s settingsImpl) GetGitHubAdditionalHeaders() any {
	return s.GitHub.AdditionalHeaders
}

func (s settingsImpl) GetCLIArgs() CLIArgs {
	return s.CLIArgs
}

// gitHubImpl is the internal implementation of the GitHub interface
type gitHubImpl struct {
	Issue              GitHubIssue     `json:"issue"`
	PR                 GitHubPR        `json:"pr"`
	ThisPR             GitHubAPIPR     `json:"thisPR"`
	Commits            []GitHubCommit  `json:"commits"`
	Reviews            []GitHubReview  `json:"reviews"`
	RequestedReviewers GitHubReviewers `json:"requested_reviewers"`
}

func (g gitHubImpl) GetIssue() GitHubIssue {
	return g.Issue
}

func (g gitHubImpl) GetPR() GitHubPR {
	return g.PR
}

func (g gitHubImpl) GetThisPR() GitHubAPIPR {
	return g.ThisPR
}

func (g gitHubImpl) GetCommits() []GitHubCommit {
	return g.Commits
}

func (g gitHubImpl) GetReviews() []GitHubReview {
	return g.Reviews
}

func (g gitHubImpl) GetRequestedReviewers() GitHubReviewers {
	return g.RequestedReviewers
}

// gitLabImpl is the internal implementation of the GitLab interface
type gitLabImpl struct {
	Metadata  RepoMetaData     `json:"Metadata"`
	MR        GitLabMR         `json:"mr"`
	Commits   []GitLabMRCommit `json:"commits"`
	Approvals GitLabApproval   `json:"approvals"`
}

func (g gitLabImpl) GetMetadata() RepoMetaData {
	return g.Metadata
}

func (g gitLabImpl) GetMR() GitLabMR {
	return g.MR
}

func (g gitLabImpl) GetCommits() []GitLabMRCommit {
	return g.Commits
}

func (g gitLabImpl) GetApprovals() GitLabApproval {
	return g.Approvals
}

// DSLData is used for JSON unmarshaling, with concrete types
type DSLData struct {
	Git      gitImpl      `json:"git"`
	GitHub   gitHubImpl   `json:"github,omitempty"`
	GitLab   gitLabImpl   `json:"gitlab,omitempty"`
	Settings settingsImpl `json:"settings"`
}

// ToInterface converts DSLData to DSL with interfaces
func (d DSLData) ToInterface() DSL {
	return DSL{
		Git:      d.Git,
		GitHub:   d.GitHub,
		GitLab:   d.GitLab,
		Settings: d.Settings,
	}
}

type CLIArgs struct {
	Base               string `json:"base"`
	Verbose            string `json:"verbose"`
	ExternalCIProvider string `json:"externalCiProvider"`
	TextOnly           bool   `json:"textOnly"` // JS has this as string
	Dangerfile         string `json:"dangerfile"`
	ID                 string `json:"id"`
	Staging            bool   `json:"staging"`
}
