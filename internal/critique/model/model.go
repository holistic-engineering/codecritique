package model

type PullRequest struct {
	Title       string
	Branch      string
	Description string
	Diff        string
}

type Review struct {
	PullRequest       *PullRequest `json:"-"`
	Summary           string       `json:"summary"`
	OverallImpression string       `json:"overall_impression"`
	CodeQuality       CodeQuality  `json:"code_quality"`
	PotentialIssues   []string     `json:"potential_issues"`
	Suggestions       []string     `json:"suggestions"`
	SecurityConcerns  string       `json:"security_concerns"`
	Testing           string       `json:"testing"`
	EstimatedEffort   string       `json:"estimated_effort_to_review"`
	CodeFeedback      []Feedback   `json:"code_feedback"`
}

type CodeQuality struct {
	Strengths           []string `json:"strengths"`
	AreasForImprovement []string `json:"areas_for_improvement"`
}

type Feedback struct {
	File       string `json:"file"`
	Line       *int   `json:"line,omitempty"`
	Suggestion string `json:"suggestion"`
}
