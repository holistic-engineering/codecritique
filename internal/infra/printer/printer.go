package printer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"

	"github.com/holistic-engineering/codecritique/config"
	"github.com/holistic-engineering/codecritique/internal/critique/model"
)

type Kind string

const (
	KindJSON Kind = "json"
	KindHTML Kind = "html"
)

type print func(*model.Review) error

type Printer struct {
	print print
}

func New(cfg *config.PrinterConfig) (*Printer, error) {
	switch Kind(cfg.Kind) {
	case KindJSON:
		return &Printer{
			print: printJSON,
		}, nil
	case KindHTML:
		return &Printer{
			print: printHTML,
		}, nil
	default:
		return nil, fmt.Errorf("printer kind %s not available", cfg.Kind)
	}
}

func (p *Printer) Print(review *model.Review) error {
	if err := p.print(review); err != nil {
		return fmt.Errorf("could not print review: %w", err)
	}

	return nil
}

func printJSON(review *model.Review) error {
	reviewJSON, err := json.MarshalIndent(&review, "", "    ")
	if err != nil {
		return fmt.Errorf("json.MarshalIndent: %w", err)
	}

	fmt.Print(string(reviewJSON))

	return nil
}

func printHTML(review *model.Review) error {
	tmpl, err := template.New("review").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, review); err != nil {
		return fmt.Errorf("failed to execute HTML template: %w", err)
	}

	fmt.Print(buf.String())

	return nil
}

const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CodeCritique Review</title>
    <script src="https://cdn.tailwindcss.com"></script>
</head>
<body class="bg-gray-100 font-sans">
    <div class="container mx-auto px-4 py-8">
        <h1 class="text-3xl font-bold mb-6">CodeCritique Review</h1>
        
        <div class="bg-white shadow-md rounded-lg p-6 mb-6">
            <h2 class="text-2xl font-semibold mb-4">Pull Request Details</h2>
            <p class="mb-2"><span class="font-semibold">Title:</span> {{.PullRequest.Title}}</p>
            <p class="mb-2"><span class="font-semibold">Branch:</span> {{.PullRequest.Branch}}</p>
            <p><span class="font-semibold">Description:</span> {{.PullRequest.Description}}</p>
        </div>

        <div class="bg-white shadow-md rounded-lg p-6 mb-6">
            <h2 class="text-2xl font-semibold mb-4">Review Summary</h2>
            <p class="mb-2"><span class="font-semibold">Summary:</span> {{.Summary}}</p>
            <p class="mb-2"><span class="font-semibold">Overall Impression:</span> {{.OverallImpression}}</p>
            <p class="mb-2"><span class="font-semibold">Estimated Effort:</span> {{.EstimatedEffort}}</p>
        </div>

        <div class="bg-white shadow-md rounded-lg p-6 mb-6">
            <h2 class="text-2xl font-semibold mb-4">Code Quality</h2>
            <div class="mb-4">
                <h3 class="text-xl font-semibold mb-2">Strengths</h3>
                <ul class="list-disc pl-6">
                    {{range .CodeQuality.Strengths}}
                    <li>{{.}}</li>
                    {{end}}
                </ul>
            </div>
            <div>
                <h3 class="text-xl font-semibold mb-2">Areas for Improvement</h3>
                <ul class="list-disc pl-6">
                    {{range .CodeQuality.AreasForImprovement}}
                    <li>{{.}}</li>
                    {{end}}
                </ul>
            </div>
        </div>

        <div class="bg-white shadow-md rounded-lg p-6 mb-6">
            <h2 class="text-2xl font-semibold mb-4">Potential Issues</h2>
            <ul class="list-disc pl-6">
                {{range .PotentialIssues}}
                <li>{{.}}</li>
                {{end}}
            </ul>
        </div>

        <div class="bg-white shadow-md rounded-lg p-6 mb-6">
            <h2 class="text-2xl font-semibold mb-4">Suggestions</h2>
            <ul class="list-disc pl-6">
                {{range .Suggestions}}
                <li>{{.}}</li>
                {{end}}
            </ul>
        </div>

        <div class="bg-white shadow-md rounded-lg p-6 mb-6">
            <h2 class="text-2xl font-semibold mb-4">Security Concerns</h2>
            <p>{{.SecurityConcerns}}</p>
        </div>

        <div class="bg-white shadow-md rounded-lg p-6 mb-6">
            <h2 class="text-2xl font-semibold mb-4">Testing</h2>
            <p>{{.Testing}}</p>
        </div>

        <div class="bg-white shadow-md rounded-lg p-6">
            <h2 class="text-2xl font-semibold mb-4">Code Feedback</h2>
            {{range .CodeFeedback}}
            <div class="mb-4 p-4 bg-gray-50 rounded-lg">
                <p class="mb-2"><span class="font-semibold">File:</span> {{.File}}</p>
                {{if .Line}}<p class="mb-2"><span class="font-semibold">Line:</span> {{.Line}}</p>{{end}}
                <p><span class="font-semibold">Suggestion:</span> {{.Suggestion}}</p>
            </div>
            {{end}}
        </div>
    </div>
</body>
</html>
`