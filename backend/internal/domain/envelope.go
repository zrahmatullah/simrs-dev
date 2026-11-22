package domain

// Meta holds pagination information
type Meta struct {
	Page      int `json:"page"`
	Limit     int `json:"limit"`
	TotalData int `json:"total_data"`
	TotalPage int `json:"total_page"`
}

// ResponseEnvelope is the standard JSON envelope specified in AGENTS.md
type ResponseEnvelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Meta    *Meta       `json:"meta,omitempty"`
	Errors  interface{} `json:"errors"`
}
