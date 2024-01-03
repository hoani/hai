package aiconfig

type OpenAI struct {
	Key string `json:"key"`
}

type Config struct {
	OpenAI OpenAI `json:"openai"`
}
