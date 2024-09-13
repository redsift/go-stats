package tags

func Bool(key string, value bool) Low {
	if value {
		return Low(key + ":true")
	}
	return Low(key + ":false")
}

func BoolP(key string, value *bool, defaultValue bool) Low {
	if value != nil {
		return Bool(key, *value)
	}
	return Bool(key, defaultValue)
}

func Error(key string, err error) Tag {
	if err != nil {
		return High(key + ":" + err.Error())
	}
	return Empty{}
}

func String(key, value string) Low {
	return Low(key + ":" + value)
}
