package catalog

type Message struct {
	Key      string `json:"key"`
	Language string `json:"language"`
	Text     string `json:"text"`
}

type Catalog struct {
	values map[string]Message
}

func New() *Catalog {
	return &Catalog{values: map[string]Message{
		"arrival": {Key: "arrival", Language: "zh-CN", Text: "愿你在烛光中停留片刻"},
		"light":   {Key: "light", Language: "zh-CN", Text: "点亮一盏灯，留下一句思念"},
		"quiet":   {Key: "quiet", Language: "zh-CN", Text: "星点归于安静，记忆仍在"},
	}}
}

func (c *Catalog) Lookup(key string) (Message, bool) {
	value, ok := c.values[key]
	return value, ok
}

func (c *Catalog) Add(message Message) bool {
	if message.Key == "" || message.Language == "" || message.Text == "" {
		return false
	}
	c.values[message.Key] = message
	return true
}

func (c *Catalog) Keys() []string {
	result := make([]string, 0, len(c.values))
	for key := range c.values {
		result = append(result, key)
	}
	return result
}
