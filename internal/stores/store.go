package stores

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/pupload/pupload/internal/models"
	"github.com/pupload/pupload/internal/stores/s3"
)

func UnmarshalStore(input models.StoreInput) (models.Store, error) {

	switch input.Type {

	case "s3":
		var params s3.S3StoreInput
		if err := decodeParams(input.Params, &params); err != nil {
			return nil, fmt.Errorf("error decoding params for s3local, %w", err)
		}

		store, err := s3.NewS3Store(params)
		if err != nil {
			return nil, fmt.Errorf("Unable to create s3 store: %w", err)
		}

		return store, nil

	default:
		return nil, fmt.Errorf("invalid store tpye")
	}
}

func decodeParams(raw json.RawMessage, out any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(out); err != nil {
		return err
	}

	if dec.More() {
		return fmt.Errorf("Unexpected extra data in params")
	}

	return nil
}
