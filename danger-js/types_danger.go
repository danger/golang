package dangerJs

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var (
	// Compiled regex patterns for diff parsing
	addedLineRe   = regexp.MustCompile(`^\+([^+].*|$)`)
	removedLineRe = regexp.MustCompile(`^-([^-].*|$)`)
	hunkHeaderRe  = regexp.MustCompile(`^@@\s+-(\d+)(?:,(\d+))?\s+\+(\d+)(?:,(\d+))?\s+@@`)

	// Shell metacharacters that could be used for command injection
	shellMetaChars = []string{";", "|", "&", "$", "`", "(", ")", "{", "}", "[", "]", "*", "?", "<", ">", "'", "\""}

	// Whitespace characters that should be rejected in git refs
	whitespaceChars = []string{" ", "\t", "\n", "\r"}
)

type GitHub interface {
	Issue() GitHubIssue
	PR() GitHubPR
	ThisPR() GitHubAPIPR
	Commits() []GitHubCommit
	Reviews() []GitHubReview
	RequestedReviewers() GitHubReviewers
}

type GitLab interface {
	Metadata() RepoMetaData
	MR() GitLabMR
	Commits() []GitLabMRCommit
	Approvals() GitLabApproval
}

type Settings interface {
	GitHubAccessToken() string
	GitHubBaseURL() string
	GitHubAdditionalHeaders() any
	CLIArgs() CLIArgs
}

type Git interface {
	ModifiedFiles() []FilePath
	CreatedFiles() []FilePath
	DeletedFiles() []FilePath
	Commits() []GitCommit
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
	ModifiedFilesList []FilePath  `json:"modified_files"`
	CreatedFilesList  []FilePath  `json:"created_files"`
	DeletedFilesList  []FilePath  `json:"deleted_files"`
	CommitsList       []GitCommit `json:"commits"`
}

func (g gitImpl) ModifiedFiles() []FilePath {
	return g.ModifiedFilesList
}

func (g gitImpl) CreatedFiles() []FilePath {
	return g.CreatedFilesList
}

func (g gitImpl) DeletedFiles() []FilePath {
	return g.DeletedFilesList
}

func (g gitImpl) Commits() []GitCommit {
	return g.CommitsList
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

// validateFilePath validates that the file path doesn't contain dangerous characters
func validateFilePath(path string) bool {
	// Empty paths are invalid
	if path == "" {
		return false
	}

	// Clean the path and check for dangerous patterns
	cleaned := filepath.Clean(path)

	// Reject paths that try to escape the repository
	if strings.Contains(cleaned, "..") {
		return false
	}

	// Reject absolute paths as they could access files outside the repository
	if filepath.IsAbs(cleaned) {
		return false
	}

	// Reject paths with shell metacharacters that could be used for command injection
	for _, char := range shellMetaChars {
		if strings.Contains(path, char) {
			return false
		}
	}

	// Reject paths with whitespace characters that could cause issues in shell commands
	for _, char := range whitespaceChars {
		if strings.Contains(path, char) {
			return false
		}
	}

	return true
}

// validateGitRef validates that the git ref name doesn't contain dangerous characters
func validateGitRef(ref string) bool {
	// Git refs must not contain certain characters and must not be empty
	if ref == "" {
		return false
	}
	// Disallow shell metacharacters
	for _, char := range shellMetaChars {
		if strings.Contains(ref, char) {
			return false
		}
	}
	// Disallow whitespace characters
	for _, char := range whitespaceChars {
		if strings.Contains(ref, char) {
			return false
		}
	}
	// Disallow path traversal
	if strings.Contains(ref, "..") {
		return false
	}
	// Disallow slashes at the start or end, or consecutive slashes
	if strings.HasPrefix(ref, "/") || strings.HasSuffix(ref, "/") || strings.Contains(ref, "//") {
		return false
	}
	// Disallow ref names with ASCII control characters
	for _, r := range ref {
		if r < 32 || r == 127 {
			return false
		}
	}
	return true
}

// DiffForFileWithRefs executes a git diff command for a specific file with configurable references.
func (g gitImpl) DiffForFileWithRefs(filePath, baseRef, headRef string) (FileDiff, error) {
	// Validate file path to prevent command injection
	if !validateFilePath(filePath) {
		return FileDiff{}, fmt.Errorf("invalid file path: %s", filePath)
	}
	// Validate baseRef and headRef to prevent command injection
	if !validateGitRef(baseRef) {
		return FileDiff{}, fmt.Errorf("invalid base ref: %s", baseRef)
	}
	if !validateGitRef(headRef) {
		return FileDiff{}, fmt.Errorf("invalid head ref: %s", headRef)
	}

	cmd := exec.Command("git", "diff", "--unified=0", baseRef, headRef, filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return FileDiff{}, err
	}

	return parseDiffContent(out.String()), nil
}

// parseHunkHeader extracts line number information from a hunk header
func parseHunkHeader(line string) (removedStart, addedStart int, isHunkHeader bool) {
	if matches := hunkHeaderRe.FindStringSubmatch(line); len(matches) > 3 {
		var err error
		removedStart, err = strconv.Atoi(matches[1])
		if err != nil {
			return 0, 0, false
		}
		addedStart, err = strconv.Atoi(matches[3])
		if err != nil {
			return 0, 0, false
		}
		return removedStart, addedStart, true
	}
	return 0, 0, false
}

// parseAddedLine extracts content from an added line and returns whether it's an added line
func parseAddedLine(line string) (content string, isAdded bool) {
	if matches := addedLineRe.FindStringSubmatch(line); len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}

// parseRemovedLine extracts content from a removed line and returns whether it's a removed line
func parseRemovedLine(line string) (content string, isRemoved bool) {
	if matches := removedLineRe.FindStringSubmatch(line); len(matches) > 1 {
		return matches[1], true
	}
	return "", false
}

// parseDiffContent parses git diff output and extracts added and removed lines with line numbers
func parseDiffContent(diffContent string) FileDiff {
	var fileDiff FileDiff

	lines := strings.Split(diffContent, "\n")
	// Initialize line numbers to -1 to indicate no hunk header has been found yet
	currentRemovedLine := -1
	currentAddedLine := -1

	for _, line := range lines {
		// Check for hunk header to track line numbers
		if removedStart, addedStart, isHunk := parseHunkHeader(line); isHunk {
			currentRemovedLine = removedStart
			currentAddedLine = addedStart
		} else if content, isAdded := parseAddedLine(line); isAdded {
			// Only add line if we have a valid line number from a hunk header
			if currentAddedLine >= 0 {
				fileDiff.AddedLines = append(fileDiff.AddedLines, DiffLine{
					Content: content,
					Line:    currentAddedLine,
				})
				currentAddedLine++
			}
		} else if content, isRemoved := parseRemovedLine(line); isRemoved {
			// Only add line if we have a valid line number from a hunk header
			if currentRemovedLine >= 0 {
				fileDiff.RemovedLines = append(fileDiff.RemovedLines, DiffLine{
					Content: content,
					Line:    currentRemovedLine,
				})
				currentRemovedLine++
			}
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
	CLIArgsData CLIArgs `json:"cliArgs"`
}

// GitHubAccessToken returns the GitHub access token
func (s settingsImpl) GitHubAccessToken() string {
	return s.GitHub.AccessToken
}

func (s settingsImpl) GitHubBaseURL() string {
	return s.GitHub.BaseURL
}

func (s settingsImpl) GitHubAdditionalHeaders() any {
	return s.GitHub.AdditionalHeaders
}

func (s settingsImpl) CLIArgs() CLIArgs {
	return s.CLIArgsData
}

// gitHubImpl is the internal implementation of the GitHub interface
type gitHubImpl struct {
	IssueData              GitHubIssue     `json:"issue"`
	PRData                 GitHubPR        `json:"pr"`
	ThisPRData             GitHubAPIPR     `json:"thisPR"`
	CommitsList            []GitHubCommit  `json:"commits"`
	ReviewsList            []GitHubReview  `json:"reviews"`
	RequestedReviewersData GitHubReviewers `json:"requested_reviewers"`
}

func (g gitHubImpl) Issue() GitHubIssue {
	return g.IssueData
}

func (g gitHubImpl) PR() GitHubPR {
	return g.PRData
}

func (g gitHubImpl) ThisPR() GitHubAPIPR {
	return g.ThisPRData
}

func (g gitHubImpl) Commits() []GitHubCommit {
	return g.CommitsList
}

func (g gitHubImpl) Reviews() []GitHubReview {
	return g.ReviewsList
}

func (g gitHubImpl) RequestedReviewers() GitHubReviewers {
	return g.RequestedReviewersData
}

// gitLabImpl is the internal implementation of the GitLab interface
type gitLabImpl struct {
	MetadataData  RepoMetaData     `json:"Metadata"`
	MRData        GitLabMR         `json:"mr"`
	CommitsList   []GitLabMRCommit `json:"commits"`
	ApprovalsData GitLabApproval   `json:"approvals"`
}

func (g gitLabImpl) Metadata() RepoMetaData {
	return g.MetadataData
}

func (g gitLabImpl) MR() GitLabMR {
	return g.MRData
}

func (g gitLabImpl) Commits() []GitLabMRCommit {
	return g.CommitsList
}

func (g gitLabImpl) Approvals() GitLabApproval {
	return g.ApprovalsData
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
