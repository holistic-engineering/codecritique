package main

import (
	"context"
	"log"
	"os"
	"strings"

	"github.com/holistic-engineering/codecritique/internal/critique"
	"github.com/holistic-engineering/codecritique/internal/infra/ai"
	"github.com/holistic-engineering/codecritique/internal/infra/git"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("Usage: codecritique <owner/repo> <pr_number>")
	}

	repoPath, prNumber := os.Args[1], os.Args[2]
	parts := strings.Split(repoPath, "/")
	if len(parts) != 2 {
		log.Fatal("Invalid repository path. Use the format: owner/repo")
	}

	owner, repo := parts[0], parts[1]
	token := os.Getenv("GIT_TOKEN")

	fetcher, err := git.New(git.GitHub, token)
	if err != nil {
		log.Fatalf("could not inititlize git client: %s", err)
	}

	reviwer := ai.New(ai.KindOllama)
	critique := critique.New(fetcher, reviwer)
	if err := critique.Criticize(context.Background(), owner, repo, prNumber); err != nil {
		log.Fatalf("could not criticize pull request: %s", err)
	}
}
