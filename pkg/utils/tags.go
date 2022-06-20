package utils

func WithNameTag(tags map[string]string, name string) map[string]string {
	tags["Name"] = name
	return tags
}
