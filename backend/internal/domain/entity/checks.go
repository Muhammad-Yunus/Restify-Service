package entity

import "fmt"

// validateStruct runs struct validation and wraps any error with context.
func validateStruct(v any) error {
	if err := validate.Struct(v); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}
