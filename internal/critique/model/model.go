package model

type PullRequest struct {
	Title       string
	Description string
	Files       []*File
}

type File struct {
	Name    string
	Content string
}

type Review struct {
	PullRequest *PullRequest
	Suggestions []*Suggestion
}

type Suggestion struct {
	File    *File
	Line    int
	Message string
}
