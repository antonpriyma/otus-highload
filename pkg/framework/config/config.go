package config

type Config interface {
	Unmarshal(rawVal interface{}) error
	UnmarshalKey(key string, rawVal interface{}) error
	GetString(string) string
	GetStringMap(key string) map[string]interface{}
}
