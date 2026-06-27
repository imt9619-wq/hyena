package movements

// Represent the minic of keyBoard input
type Inputs struct{
	W, A, S, D           KeyPress
    Space, Shift, Sprint KeyPress
}

type KeyPress struct{
	Pressed   bool
    PressOnce bool
}

func (in Inputs) NextTickInputs() Inputs{
	nextIn := Inputs{}
	if in.W.Pressed && !in.W.PressOnce{nextIn.W = in.W}
	if in.A.Pressed && !in.A.PressOnce{nextIn.A = in.A}
	if in.S.Pressed && !in.S.PressOnce{nextIn.S = in.S}
	if in.D.Pressed && !in.D.PressOnce{nextIn.D = in.D}

	if in.Space.Pressed && !in.Space.PressOnce{nextIn.Space = in.Space}
	if in.Shift.Pressed && !in.Shift.PressOnce{nextIn.Shift = in.Shift}
	if in.Sprint.Pressed && !in.Sprint.PressOnce{nextIn.Sprint = in.Sprint}
	return nextIn
}