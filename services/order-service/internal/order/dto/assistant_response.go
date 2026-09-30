package dto

type AssistantResponse struct {
	Answer  string `json:"answer"`
	RawData []any  `json:"raw_data"`
}
