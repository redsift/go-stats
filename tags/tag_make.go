package tags

type L = Low
type H = High

func D(low, high string) Dual {
	return Dual{low, high}
}

func E() Empty {
	return Empty{}
}
