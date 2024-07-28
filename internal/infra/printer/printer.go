package printer

import (
	"encoding/json"
	"fmt"

	"github.com/holistic-engineering/codecritique/config"
	"github.com/holistic-engineering/codecritique/internal/critique/model"
)

type Kind string

const (
	KindJSON     Kind = "json"
	KindHTML     Kind = "html"
	KindMarkdown Kind = "markdown"
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
	case KindHTML, KindMarkdown:
		return nil, fmt.Errorf("printer kind %s not implemented", cfg.Kind)
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
