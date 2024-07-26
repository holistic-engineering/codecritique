package git

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v57/github"
	"github.com/holistic-engineering/codecritique/config"
	"github.com/holistic-engineering/codecritique/internal/critique/model"
	"github.com/xanzy/go-gitlab"
	"golang.org/x/oauth2"
)

type Provider string

const (
	GitHub Provider = "GitHub"
	GitLab Provider = "GitLab"
)

type Client struct {
	provider     Provider
	githubClient *github.Client
	gitlabClient *gitlab.Client
}

func New(cfg *config.GitConfig) (*Client, error) {
	switch Provider(cfg.Provider) {
	case GitHub:
		ctx := context.Background()
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: cfg.Token},
		)
		tc := oauth2.NewClient(ctx, ts)
		client := github.NewClient(tc)
		return &Client{
			provider:     GitHub,
			githubClient: client,
		}, nil
	case GitLab:
		client, err := gitlab.NewClient(cfg.Token)
		if err != nil {
			return nil, fmt.Errorf("failed to create GitLab client: %w", err)
		}
		return &Client{
			provider:     GitLab,
			gitlabClient: client,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported Git provider: %s", cfg.Provider)
	}
}

func (c *Client) FetchPullRequest(ctx context.Context, owner, repo, number string) (*model.PullRequest, error) {
	switch c.provider {
	case GitHub:
		return c.fetchGitHubPR(ctx, owner, repo, number)
	case GitLab:
		return c.fetchGitLabMR(owner, repo, number)
	default:
		return nil, fmt.Errorf("unsupported Git provider: %s", c.provider)
	}
}

func (c *Client) fetchGitHubPR(ctx context.Context, owner, repo, number string) (*model.PullRequest, error) {
	prNumber, err := strconv.Atoi(number)
	if err != nil {
		return nil, fmt.Errorf("invalid PR number: %w", err)
	}

	pr, _, err := c.githubClient.PullRequests.Get(ctx, owner, repo, prNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub PR: %w", err)
	}

	files, _, err := c.githubClient.PullRequests.ListFiles(ctx, owner, repo, prNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitHub PR files: %w", err)
	}

	modelFiles := make([]*model.File, len(files))
	for i, file := range files {
		content, _, _, err := c.githubClient.Repositories.GetContents(ctx, owner, repo, file.GetFilename(), &github.RepositoryContentGetOptions{
			Ref: pr.GetHead().GetSHA(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch file content: %w", err)
		}

		fileContent, err := content.GetContent()
		if err != nil {
			return nil, fmt.Errorf("failed to decode file content: %w", err)
		}

		modelFiles[i] = &model.File{
			Name:    file.GetFilename(),
			Content: fileContent,
		}
	}

	return &model.PullRequest{
		Title:       pr.GetTitle(),
		Description: pr.GetBody(),
		Files:       modelFiles,
	}, nil
}

func (c *Client) fetchGitLabMR(owner, repo, number string) (*model.PullRequest, error) {
	mrNumber, err := strconv.Atoi(number)
	if err != nil {
		return nil, fmt.Errorf("invalid MR number: %w", err)
	}

	mr, _, err := c.gitlabClient.MergeRequests.GetMergeRequest(owner+"/"+repo, mrNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitLab MR: %w", err)
	}

	changes, _, err := c.gitlabClient.MergeRequests.ListMergeRequestDiffs(owner+"/"+repo, mrNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch GitLab MR changes: %w", err)
	}

	modelFiles := make([]*model.File, len(changes))
	for i, change := range changes {
		fileContent, _, err := c.gitlabClient.RepositoryFiles.GetRawFile(owner+"/"+repo, change.NewPath, &gitlab.GetRawFileOptions{
			Ref: gitlab.Ptr(mr.SourceBranch),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch file content: %w", err)
		}

		modelFiles[i] = &model.File{
			Name:    change.NewPath,
			Content: string(fileContent),
		}
	}

	return &model.PullRequest{
		Title:       mr.Title,
		Description: mr.Description,
		Files:       modelFiles,
	}, nil
}
