package dangerJs

import "time"

type GitHubIssue struct {
	Labels []GitHubIssueLabel `json:"labels"`
}

type GitHubIssueLabel struct {
	ID    int64  `json:"id"`
	URL   string `json:"url"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type GitHubUser struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Type      string `json:"type"` // "User" | "Organization" | "Bot"
	AvatarURL string `json:"avatar_url"`
	HRef      string `json:"href"`
}

type GitHubPR struct {
	Number            int            `json:"number"`
	State             string         `json:"state"` // "closed" | "open" | "merged" | "locked"
	Locked            bool           `json:"locked"`
	Title             string         `json:"title"`
	Body              string         `json:"body"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	ClosedAt          time.Time      `json:"closed_at,omitempty"`
	MergedAt          time.Time      `json:"merged_at,omitempty"`
	Head              GitHubMergeRef `json:"head"`
	Base              GitHubMergeRef `json:"base"`
	User              GitHubUser     `json:"user"`
	Assignee          GitHubUser     `json:"assignee"`
	Assignees         []GitHubUser   `json:"assignees"`
	Draft             bool           `json:"draft"`
	Merged            bool           `json:"merged"`
	Comments          int            `json:"comments"`
	ReviewComments    int            `json:"review_comments"`
	Commits           int            `json:"commits"`
	Additions         int            `json:"additions"`
	Deletions         int            `json:"deletions"`
	ChangedFiles      int            `json:"changed_files"`
	HTMLURL           string         `json:"html_url"`
	AuthorAssociation string         `json:"author_association"` // "COLLABORATOR", "CONTRIBUTOR", "FIRST_TIMER", "FIRST_TIME_CONTRIBUTOR", "MEMBER", "NONE", "OWNER"
}
type GitHubMergeRef struct {
	Label string     `json:"label"`
	Ref   string     `json:"ref"`
	SHA   string     `json:"sha"`
	User  GitHubUser `json:"user"`
	Repo  GitHubRepo `json:"repo"`
}

type GitHubRepo struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	IsPrivate   bool   `json:"private"`
	Description string `json:"description,omitempty"`
	IsFork      bool   `json:"fork"`
	HTMLURL     string `json:"html_url"`
}

type GitHubCommit struct {
	Commit    GitCommit  `json:"commit"`
	SHA       string     `json:"sha"`
	URL       string     `json:"url"`
	Author    GitHubUser `json:"author"`
	Committer GitHubUser `json:"committer"`
	Parents   []any      `json:"parents"`
}

type GitHubReview struct {
	User     GitHubUser `json:"user"`
	ID       int64      `json:"id,omitempty"`
	Body     string     `json:"body,omitempty"`
	CommitID string     `json:"commit_id,omitempty"`
	State    string     `json:"state,omitempty"` // APPROVED | CHANGES_REQUESTED | COMMENTED | PENDING | DISMISSED
}

type GitHubReviewers struct {
	Users []GitHubUser `json:"users"`
	Teams []any        `json:"teams"`
}

type GitHubMilestone struct {
	ID           int64      `json:"id"`
	Number       int64      `json:"number"`
	State        string     `json:"state"` // "closed" | "open" | "all"
	Title        string     `json:"title"`
	Description  string     `json:"description,omitempty"`
	Creator      GitHubUser `json:"creator"`
	OpenIssues   int        `json:"open_issues"`
	ClosedIssues int        `json:"closed_issues"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ClosedAt     time.Time  `json:"closed_at"`
	DueOn        time.Time  `json:"due_on"`
}

type GitHubAPIPR struct {
	Owner  string `json:"owner"`
	Repo   string `json:"repo"`
	Number int    `json:"number"`
}
