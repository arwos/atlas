package app

import "encoding/json"

func decode[T any](p json.RawMessage, out *T) error {
	if err := json.Unmarshal(p, out); err != nil {
		return err
	}
	return nil
}
