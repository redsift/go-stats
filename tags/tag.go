package tags

type Tag interface {
	AddTo([2][]string) [2][]string
}

type High string

func (h High) AddTo(tags [2][]string) [2][]string {
	tags[1] = append(tags[1], string(h))
	return tags
}

func (h High) Low() Low {
	return Low(h)
}

type Low string

func (l Low) AddTo(tags [2][]string) [2][]string {
	tags[0] = append(tags[0], string(l))
	return tags
}

func (l Low) High() High {
	return High(l)
}
