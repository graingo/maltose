package mvar

func (v *Var) Set(value any) (old any) {
	if v.safe {
		v.mu.Lock()
		defer v.mu.Unlock()
	}
	old = v.value
	v.value = value
	return
}
