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
	if len(os.Args) < 4 {
		log.Fatal("Usage: codecritique <owner/repo> <pr_number> <ai_provider>")
	}

	repoPath, prNumber, aiProvider := os.Args[1], os.Args[2], os.Args[3]
	parts := strings.Split(repoPath, "/")
	if len(parts) != 2 {
		log.Fatal("Invalid repository path. Use the format: owner/repo")
	}

	owner, repo := parts[0], parts[1]
	token := os.Getenv("GIT_TOKEN")

	fetcher, err := git.New(git.GitHub, token)
	if err != nil {
		log.Fatalf("could not initialize git client: %s", err)
	}

	var reviewer ai.Provider
	switch strings.ToLower(aiProvider) {
	case "ollama":
		reviewer = ai.ProviderOllama
	case "groq":
		reviewer = ai.ProviderGroq
	default:
		log.Fatalf("unsupported AI provider: %s", aiProvider)
	}

	aiClient := ai.New(reviewer)
	critique := critique.New(fetcher, aiClient)
	if err := critique.Criticize(context.Background(), owner, repo, prNumber); err != nil {
		log.Fatalf("could not criticize pull request: %s", err)
	}
}
