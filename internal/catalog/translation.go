package catalog

import (
	"sort"
	"strings"
)

type Locale string

const (
	LocaleChinese Locale = "zh-CN"
	LocaleEnglish Locale = "en"
)

type Bundle struct {
	Locale   Locale             `json:"locale"`
	Messages map[string]Message `json:"messages"`
}

type Renderer struct {
	primary  *Catalog
	fallback *Catalog
}

func NewRenderer(primary, fallback *Catalog) *Renderer {
	if primary == nil {
		primary = New()
	}
	if fallback == nil {
		fallback = New()
	}
	return &Renderer{primary: primary, fallback: fallback}
}

func (r *Renderer) Render(key string) string {
	if r == nil {
		return ""
	}
	if message, ok := r.primary.Lookup(key); ok {
		return message.Text
	}
	if message, ok := r.fallback.Lookup(key); ok {
		return message.Text
	}
	return ""
}

func (r *Renderer) RenderOr(key, fallback string) string {
	value := r.Render(key)
	if value == "" {
		return fallback
	}
	return value
}

func (r *Renderer) Has(key string) bool { return r.Render(key) != "" }

func (c *Catalog) Bundle(locale Locale) Bundle {
	result := Bundle{Locale: locale, Messages: map[string]Message{}}
	for _, key := range c.Keys() {
		message, ok := c.Lookup(key)
		if !ok {
			continue
		}
		if locale == "" || Locale(message.Language) == locale {
			result.Messages[key] = message
		}
	}
	return result
}

func (b Bundle) Keys() []string {
	result := make([]string, 0, len(b.Messages))
	for key := range b.Messages {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func (b Bundle) Valid() bool {
	if b.Locale == "" || len(b.Messages) == 0 {
		return false
	}
	for key, message := range b.Messages {
		if key == "" || message.Key != key || strings.TrimSpace(message.Text) == "" {
			return false
		}
	}
	return true
}

func (b Bundle) Merge(other Bundle) Bundle {
	result := Bundle{Locale: b.Locale, Messages: map[string]Message{}}
	for key, message := range b.Messages {
		result.Messages[key] = message
	}
	if result.Locale == "" {
		result.Locale = other.Locale
	}
	for key, message := range other.Messages {
		if _, exists := result.Messages[key]; !exists {
			result.Messages[key] = message
		}
	}
	return result
}

func NormalizeLocale(value string) Locale {
	value = strings.TrimSpace(value)
	if value == "" {
		return LocaleChinese
	}
	value = strings.ReplaceAll(value, "_", "-")
	if strings.EqualFold(value, "zh") || strings.EqualFold(value, "zh-cn") {
		return LocaleChinese
	}
	if strings.EqualFold(value, "en") || strings.EqualFold(value, "en-us") {
		return LocaleEnglish
	}
	return Locale(value)
}

func (c *Catalog) Localized(key string, locale Locale) (Message, bool) {
	message, ok := c.Lookup(key)
	if !ok {
		return Message{}, false
	}
	if locale != "" && NormalizeLocale(message.Language) != NormalizeLocale(string(locale)) {
		return Message{}, false
	}
	return message, true
}

func (c *Catalog) Clone() *Catalog {
	clone := &Catalog{values: map[string]Message{}}
	for _, key := range c.Keys() {
		if message, ok := c.Lookup(key); ok {
			clone.values[key] = message
		}
	}
	return clone
}

func (c *Catalog) Remove(key string) bool {
	if _, ok := c.values[key]; !ok {
		return false
	}
	delete(c.values, key)
	return true
}

func (c *Catalog) Replace(messages []Message) error {
	for _, message := range messages {
		if err := ValidateMessage(message); err != nil {
			return err
		}
	}
	for _, message := range messages {
		c.values[message.Key] = message
	}
	return nil
}

func (c *Catalog) Missing(keys []string) []string {
	missing := []string{}
	for _, key := range keys {
		if _, ok := c.Lookup(key); !ok {
			missing = append(missing, key)
		}
	}
	sort.Strings(missing)
	return missing
}

func (c *Catalog) Ordered() []Message {
	keys := c.Keys()
	result := make([]Message, 0, len(keys))
	for _, key := range keys {
		if message, ok := c.Lookup(key); ok {
			result = append(result, message)
		}
	}
	return result
}
