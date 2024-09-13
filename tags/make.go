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

func L(s string) Low {
	return Low(s)
}

func H(s string) High {
	return High(s)
}
