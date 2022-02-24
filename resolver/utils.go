package resolver

import (
	"encoding/base64"
	"kilogram-api/model"
)

func appendMeta(metas []*model.Meta, key, val string) []*model.Meta {
	found := false

	for _, meta := range metas {
		if meta.Key == key {
			meta.Val = val
			found = true
		}
	}

	if !found {
		metas = append(metas, &model.Meta{Key: key, Val: val})
	}

	return metas
}

func validateBase64(s string) error {
	if _, err := base64.StdEncoding.DecodeString(s); err != nil {
		return ErrInvalidBase64
	}

	return nil
}
