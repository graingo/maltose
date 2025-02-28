package merror

// MarshalJSON 实现了 json.Marshaler 接口
func (err Error) MarshalJSON() ([]byte, error) {
	return []byte(`"` + err.Error() + `"`), nil
}
