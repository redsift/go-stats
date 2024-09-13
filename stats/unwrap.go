package stats

func Unwrap(c Collector) Collector {
	if t, ok := c.(interface {
		Unwrap() Collector
	}); ok {
		next := Unwrap(t.Unwrap())
		if next != c {
			return Unwrap(next)
		}
		return c
	}
	return c
}

func UnwrapH(c HighCardinalityCollector) HighCardinalityCollector {
	if t, ok := c.(interface {
		Unwrap() HighCardinalityCollector
	}); ok {
		next := UnwrapH(t.Unwrap())
		if next != c {
			return UnwrapH(next)
		}
	}
	return c
}
