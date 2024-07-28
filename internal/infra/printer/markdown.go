package printer

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/holistic-engineering/codecritique/internal/critique/model"
)

type markdownPrinter struct{}

func (p *markdownPrinter) Kind() Kind {
	return KindMarkdown
}

func (p *markdownPrinter) Print(review *model.Review) error {
	tmpl, err := template.New("review").Parse(markdownTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse Markdown template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, review); err != nil {
		return fmt.Errorf("failed to execute Markdown template: %w", err)
	}

	fmt.Print(buf.String())
	return nil
}

const markdownTemplate = `
# CodeCritique Review

## Pull Request Details

- **Title:** {{.PullRequest.Title}}
- **Branch:** {{.PullRequest.Branch}}
- **Description:** {{.PullRequest.Description}}

## Review Summary

- **Summary:** {{.Summary}}
- **Overall Impression:** {{.OverallImpression}}
- **Estimated Effort:** {{.EstimatedEffort}}

## Code Quality

### Strengths

{{range .CodeQuality.Strengths}}
- {{.}}
{{end}}

### Areas for Improvement

{{range .CodeQuality.AreasForImprovement}}
- {{.}}
{{end}}

## Potential Issues

{{range .PotentialIssues}}
- {{.}}
{{end}}

## Suggestions

{{range .Suggestions}}
- {{.}}
{{end}}

## Security Concerns

{{.SecurityConcerns}}

## Testing

{{.Testing}}

## Code Feedback

{{range .CodeFeedback}}
### File: {{.File}}
{{if .Line}}**Line:** {{.Line}}{{end}}
**Suggestion:** {{.Suggestion}}

{{end}}
`
