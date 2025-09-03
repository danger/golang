package dangerJs

import "time"

type RepoMetaData struct {
	RepoSlug      string `json:"repoSlug"`
	PullRequestID string `json:"pullRequestID"`
}

type GitLabMRBase struct {
	ID           int64      `json:"id"`
	IID          int64      `json:"iid"`
	ProjectID    int64      `json:"project_id"`
	Title        string     `json:"title"`
	Description  string     `json:"description"`
	State        string     `json:"state"` // "closed" | "open" | "locked" | "merged"
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	TargetBranch string     `json:"target_branch"`
	SourceBranch string     `json:"source_branch"`
	Upvotes      int        `json:"upvotes"`
	Downvotes    int        `json:"downvotes"`
	Author       GitLabUser `json:"author"`
	User         struct {
		CanMerge bool `json:"can_merge"`
	} `json:"user"`
	Assignee                  GitLabUser      `json:"assignee,omitempty"`
	Assignees                 []GitLabUser    `json:"assignees"`
	Reviewers                 []GitLabUser    `json:"reviewers"`
	SourceProjectID           int64           `json:"source_project_id"`
	TargetProjectID           int64           `json:"target_project_id"`
	Labels                    []string        `json:"labels"`
	WorkInProgress            bool            `json:"work_in_progress"`
	Milestone                 GitLabMileStone `json:"milestone"`
	MergeWhenPipelineSucceeds bool            `json:"merge_when_pipeline_succeeds"`
	MergeStatus               string          `json:"merge_status"` // "can_be_merged"
	MergeError                any             `json:"merge_error"`
	SHA                       string          `json:"sha"`
	MergeCommitSHA            string          `json:"merge_commit_sha,omitempty"`
	UserNotesCount            int             `json:"UserNotesCount"`
	DiscussionLocked          any             `json:"DiscussionLocked"`
	ShouldRemoveSourceBranch  bool            `json:"should_remove_source_branch"`
	ForceRemoveSourceBranch   bool            `json:"force_remove_source_branch"`
	AllowCollaboration        bool            `json:"allow_collaboration"`
	AllowMaintainerToPush     bool            `json:"allow_maintainer_to_push"`
	WebURL                    string          `json:"web_url"`
	TimeStats                 GitLabTimeStats `json:"time_stats"`
}

type GitLabUser struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Username  string `json:"username"`
	State     string `json:"state"` // "active" | "blocked"
	AvatarURL string `json:"avatar_url,omitempty"`
	WebURL    string `json:"web_url"`
}

type GitLabMileStone struct {
	ID          int64     `json:"id"`
	IID         int64     `json:"iid"`
	ProjectID   int64     `json:"project_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	State       string    `json:"state"` // "closed" | "active"
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DueDate     time.Time `json:"due_date"`
	StartDate   time.Time `json:"start_date"`
	WebURL      string    `json:"web_url"`
}

type GitLabTimeStats struct {
	TimeEstimate        int `json:"time_estimate"`
	TotalTimeSpent      int `json:"total_time_spent"`
	HumanTimeEstimate   int `json:"human_time_estimate,omitempty"`
	HumanTotalTimeSpent int `json:"human_total_time_spent,omitempty"`
}

type GitLabMR struct {
	GitLabMRBase

	Squash       bool       `json:"squash"`
	Subscribed   bool       `json:"subscribed"`
	ChangesCount string     `json:"changes_count"`
	MergedBy     GitLabUser `json:"merged_by"`
	MergedAt     string     `json:"merged_at"`
	ClosedBy     GitLabUser `json:"closed_by,omitempty"`
	ClosedAt     string     `json:"closed_at,omitempty"`

	LatestBuildStartedAt        string           `json:"latest_build_started_at"`
	LatestBuildFinishedAt       string           `json:"latest_build_finished_at"`
	FirstDeployedToProductionAt string           `json:"first_deployed_to_production_at,omitempty"`
	Pipeline                    gitLabMRPipeline `json:"pipeline"`
	DiffRefs                    gitLabMRDiffRefs `json:"diff_refs"`
	DivergedCommitsCount        int              `json:"diverged_commits_count"`
	RebaseInProgress            bool             `json:"rebase_in_progress"`
	ApprovalsBeforeMerge        any              `json:"approvals_before_merge"`
}

type gitLabMRPipeline struct {
	ID     int64  `json:"id"`
	SHA    string `json:"sha"`
	Ref    string `json:"ref"`
	Status string `json:"status"` // "canceled" | "failed" | "pending" | "running" | "skipped" | "success"
	WebURL string `json:"web_url"`
}
type gitLabMRDiffRefs struct {
	BaseSHA  string `json:"base_sha"`
	HeadSHA  string `json:"head_sha"`
	StartSHA string `json:"start_sha"`
}

type GitLabMRCommit struct {
	ID             string   `json:"id"`
	ShortID        string   `json:"short_id"`
	CreatedAt      string   `json:"created_at"`
	ParentIDs      []string `json:"parent_ids"`
	Title          string   `json:"title"`
	Message        string   `json:"message"`
	AuthorName     string   `json:"author_name"`
	AuthorEmail    string   `json:"author_email"`
	AuthoredDate   string   `json:"authored_date"`
	CommitterName  string   `json:"committer_name"`
	CommitterEmail string   `json:"committer_email"`
	CommittedDate  string   `json:"committed_date"`
}

type GitLabApproval struct {
	ID                int64     `json:"id"`
	IID               int64     `json:"iid"`
	ProjectID         int64     `json:"project_id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	State             string    `json:"state"` // "closed" | "open" | "locked" | "merged"
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	MergeStatus       string    `json:"merge_status"` // "can_be_merged"
	ApprovalsRequired int       `json:"approvals_required"`
	ApprovalsLeft     int       `json:"approvals_left"`
	ApprovedBy        any       `json:"approved_by"` // {user:GitLabUser} | []GitLabUser
}
