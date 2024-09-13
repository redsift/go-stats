package tags

type List []Tag

func (t List) Get() [2][]string {
	out := [2][]string{}
	for _, tag := range t {
		out = tag.AddTo(out)
	}
	return out
}

func (t *List) Add(ts ...Tag) {
	*t = append(*t, ts...)
}

func (t *List) AddHigh(s ...string) {
	t.H(s...)
}

func (t *List) AddLow(s ...string) {
	t.L(s...)
}

func (t *List) H(s ...string) {
	*t = append(*t, HighList(s...)...)
}

func (t *List) L(s ...string) {
	*t = append(*t, LowList(s...)...)
}

func (t List) All() []string {
	all := t.Get()
	return append(all[0], all[1]...)
}

func (t List) Low() []string {
	return t.Get()[0]
}

func (t List) High() []string {
	return t.Get()[1]
}

// AddTo implements Tag
func (t List) AddTo(tags [2][]string) [2][]string {
	for _, e := range t {
		e.AddTo(tags)
	}
	return tags
}
