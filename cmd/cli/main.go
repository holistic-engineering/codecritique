package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/holistic-engineering/codecritique/config"
	"github.com/holistic-engineering/codecritique/internal/critique"
	"github.com/holistic-engineering/codecritique/internal/infra/ai"
	"github.com/holistic-engineering/codecritique/internal/infra/git"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: codecritique <owner/repo> <pr_number>")
	}

	cfg, err := config.LoadConfig("settings/settings.toml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}

	repoPath, prNumber := os.Args[1], os.Args[2]
	parts := strings.Split(repoPath, "/")
	if len(parts) != 2 {
		log.Fatal("Invalid repository path. Use the format: owner/repo")
	}

	owner, repo := parts[0], parts[1]

	git, err := git.New(&cfg.Git)
	if err != nil {
		log.Fatalf("could not initialize git client: %s", err)
	}

	ai := ai.New(&cfg.AI)
	critique := critique.New(git, ai)
	if err := critique.Criticize(context.Background(), owner, repo, prNumber); err != nil {
		log.Fatalf("could not criticize pull request: %s", err)
	}
}