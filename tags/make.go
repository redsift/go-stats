package tags

func Bool(key string, value bool) Low {
	if value {
		return Low(key + ":true")
	}
	return Low(key + ":false")
}

func String(key, value string) Low {
	return Low(key + ":" + value)
}

type L = Low
type H = High

func LowSlice[S ~string](ins ...S) (o []Low) {
	o = make([]Low, len(ins))
	for i, in := range ins {
		o[i] = Low(in)
	}
	return
}

func LowList[S ~string](ins ...S) (o List) {
	o = make(List, len(ins))
	for i, in := range ins {
		o[i] = Low(in)
	}
	return
}

func HighSlice[S ~string](ins ...S) (o []High) {
	o = make([]High, len(ins))
	for i, in := range ins {
		o[i] = High(in)
	}
	return
}

func HighList[S ~string](ins ...S) (o List) {
	o = make(List, len(ins))
	for i, in := range ins {
		o[i] = High(in)
	}
	return
}

func ToList[T Tag](ins ...T) (o List) {
	o = make(List, len(ins))
	for i, in := range ins {
		o[i] = Tag(in)
	}
	return
}
