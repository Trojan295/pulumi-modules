package utils

func WithNameTag(tags map[string]string, name string) map[string]string {
	if tags == nil {
		tags = make(map[string]string)
	}

	tags["Name"] = name
	return tags
}
