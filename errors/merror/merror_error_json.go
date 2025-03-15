package merror

// MarshalJSON implements the json.Marshaler interface.
func (err Error) MarshalJSON() ([]byte, error) {
	return []byte(`"` + err.Error() + `"`), nil
}
