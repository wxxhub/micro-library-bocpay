package library

import (
	"encoding/json"
	"errors"
)

type MixInt int

func (u *MixInt) UnmarshalJSON(bs []byte) error {
	var i int
	if err := json.Unmarshal(bs, &i); err == nil {
		*u = MixInt(i)
		return nil
	}
	var s string
	if err := json.Unmarshal(bs, &s); err != nil {
		return errors.New("expected a string or an integer")
	}
	if err := json.Unmarshal([]byte(s), &i); err != nil {
		return err
	}
	*u = MixInt(i)
	return nil
}
