package main

type Webhook struct {
	Action       string                 `json:"action"`
	CheckSuite   *WebhookCheckSuite     `json:"check_suite,omitempty"`
	CheckRun     *WebhookCheckRun       `json:"check_run,omitempty"`
	Installation WebhookInstallation    `json:"installation"`
	Repository   WebhookRepository      `json:"repository"`
	PullRequest  *WebhookPullRequest    `json:"pull_request,omitempty"`
}

type WebhookCheckSuite struct {
	Id             int                     `json:"id"`
	HeadSha        string                  `json:"head_sha"`
	PullRequests   *[]WebhookPullRequest   `json:"pull_requests"`
	Before         *string                 `json:"before,omitempty"`
	After          *string                 `json:"after,omitempty"`
}

type WebhookCheckRun struct {
	Id             int               `json:"id"`
	HeadSha        string            `json:"head_sha"`
	Url            string            `json:"url"`
	CheckSuite     WebhookCheckSuite `json:"check_suite"`
}

type WebhookPullRequest struct {
	Number        int                    `json:"number"`
	Url           string                 `json:"url"`
	Head          WebhookPullRequestHead `json:"head"`
	Base          WebhookPullRequestHead `json:"base"`
}

type WebhookPullRequestHead struct {
	Sha          string  `json:"sha"`
	Repo         WebhookRepository `json:"repo"`
}

type WebhookInstallation struct {
	Id        int      `json:"id"`
	NodeId    string   `json:"node_id"`
}

type WebhookRepository struct {
	Id             int      `json:"id"`
	RepositoryURL  string   `json:"url"`
	CompareURL     string   `json:"compare_url"`
	CommitsURL     string   `json:"commits_url"`
}


type InstallationToken struct {
	Token       string `json:"token"`
	ExpiresAt   string `json:"expires_at"`
}

type File struct {
	Sha          string  `json:"sha"`
	Filename     string  `json:"filename"`
	Status       string  `json:"status"`
	ContentsURL  string  `json:"contents_url"`
}

type CheckRun struct {
	Name         string          `json:"name"`
	HeadSha      *string         `json:"head_sha,omitempty"`
	Status       string          `json:"status"`
	Conclusion   *string         `json:"conclusion,omitempty"`
	Url          *string         `json:"url,omitempty"`
	Output       CheckRunOutput  `json:"output"`
}

type CheckRunOutput struct {
	Title        string                  `json:"title"`
	Summary      string                  `json:"summary"`
	Annotations  *[]CheckRunAnnotation   `json:"annotations,omitempty"`
}

type CheckRunAnnotation struct {
	Path         string                  `json:"path"`
	StartLine    int                     `json:"start_line"`
	EndLine      int                     `json:"end_line"`
	Level        string                  `json:"annotation_level"`
	Message      string                  `json:"message"`
}

type Files struct {
	Files   *[]File `json:"files"`
}