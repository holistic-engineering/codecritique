package printer

import (
	"encoding/json"
	"fmt"

	"github.com/holistic-engineering/codecritique/internal/critique/model"
)

type jsonPrinter struct{}

func (p *jsonPrinter) Kind() Kind {
	return KindJSON
}

func (p *jsonPrinter) Print(review *model.Review) error {
	reviewJSON, err := json.MarshalIndent(review, "", "    ")
	if err != nil {
		return fmt.Errorf("json.MarshalIndent: %w", err)
	}

	fmt.Print(string(reviewJSON))
	return nil
}
