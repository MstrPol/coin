package store

import (
	"encoding/json"
	"errors"
	"fmt"

	"coin.local/coin-api/internal/componentpackage"
)

var ErrInvalidContentRef = errors.New("invalid content_ref")

func validateContentRefOnWrite(contentRef json.RawMessage) error {
	return validateContentRefOnWriteForType("", contentRef)
}

func validateContentRefOnWriteForType(componentType string, contentRef json.RawMessage) error {
	if componentType == "agent" {
		return nil
	}
	if err := componentpackage.ValidateContentRefOnWrite(contentRef); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidContentRef, err)
	}
	return nil
}
