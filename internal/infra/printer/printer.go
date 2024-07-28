package printer

import (
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

type printer interface {
	Print(*model.Review) error
	Kind() Kind
}

type Printer struct {
	printer printer
}

func New(cfg *config.PrinterConfig) (*Printer, error) {
	switch Kind(cfg.Kind) {
	case KindJSON:
		return &Printer{
			printer: &jsonPrinter{},
		}, nil
	case KindHTML:
		return &Printer{
			printer: &htmlPrinter{},
		}, nil
	case KindMarkdown:
		return &Printer{
			printer: &markdownPrinter{},
		}, nil
	default:
		return nil, fmt.Errorf("printer kind %s not available", cfg.Kind)
	}
}

func (p *Printer) Print(review *model.Review) error {
	if err := p.printer.Print(review); err != nil {
		return fmt.Errorf("could not print for kind %s: %w", p.printer.Kind(), err)
	}
	return nil
}
