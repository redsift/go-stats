package tags

type List []Tag

func (t List) Get() [2][]string {
	out := [2][]string{}
	for _, tag := range t {
		out = tag.AddTo(out)
	}
	return out
}

func (t List) Low() []string {
	return t.Get()[0]
}

func (t List) High() []string {
	return t.Get()[1]
}
