package tags

type Tag interface {
	AddTo([2][]string) [2][]string
}

type Dual [2]string

func (d Dual) AddTo(tags [2][]string) [2][]string {
	tags[0] = append(tags[0], d[0])
	tags[1] = append(tags[1], d[1])
	return tags
}

type High string

func (h High) AddTo(tags [2][]string) [2][]string {
	tags[1] = append(tags[1], string(h))
	return tags
}

func (h High) Low() Low {
	return Low(h)
}

func (h High) WithLow(l string) Dual {
	return Dual{l, string(h)}
}

type Low string

func (l Low) AddTo(tags [2][]string) [2][]string {
	tags[0] = append(tags[0], string(l))
	return tags
}

func (l Low) High() High {
	return High(l)
}

func (l Low) WithHigh(h string) Dual {
	return Dual{string(l), h}
}
