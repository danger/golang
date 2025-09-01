package dangerJs

import (
	"bytes"
	"os/exec"
	"regexp"
	"strings"
)

type DSL struct {
	Git    Git    `json:"git"`
	GitHub GitHub `json:"github,omitempty"`
	GitLab GitLab `json:"gitlab,omitempty"`
	// TODO: bitbucket_server
	// TODO: bitbucket_cloud
	Settings Settings `json:"settings"`
}

type FilePath = string

type Git struct {
	ModifiedFiles []FilePath  `json:"modified_files"`
	CreateFiles   []FilePath  `json:"created_files"`
	DeletedFiles  []FilePath  `json:"deleted_files"`
	Commits       []GitCommit `json:"commits"`
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
func (g Git) DiffForFile(filePath string) (FileDiff, error) {
	cmd := exec.Command("git", "diff", "--unified=0", "HEAD^", "HEAD", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return FileDiff{}, err
	}

	diffContent := out.String()
	var fileDiff FileDiff
	// Only match lines that start with + or - but not +++ or --- (file headers)
	addedRe := regexp.MustCompile(`^\+([^+].*|$)`)
	removedRe := regexp.MustCompile(`^-([^-].*|$)`)

	lines := strings.Split(diffContent, "\n")
	for _, line := range lines {
		if matches := addedRe.FindStringSubmatch(line); len(matches) > 1 {
			fileDiff.AddedLines = append(fileDiff.AddedLines, DiffLine{Content: matches[1]})
		} else if matches := removedRe.FindStringSubmatch(line); len(matches) > 1 {
			fileDiff.RemovedLines = append(fileDiff.RemovedLines, DiffLine{Content: matches[1]})
		}
	}

	return fileDiff, nil
}

type Settings struct {
	GitHub struct {
		AccessToken       string `json:"accessToken"`
		BaseURL           string `json:"baseURL"`
		AdditionalHeaders any    `json:"additionalHeaders"`
	} `json:"github"`
	CLIArgs CLIArgs `json:"cliArgs"`
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
