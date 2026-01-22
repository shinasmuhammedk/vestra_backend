package response

type APIResponse struct {
	StatusCode int         `json:"status"`           // HTTP status code
	// Success    bool        `json:"success"`          // success/failure
	Message    string      `json:"message"`          // human-readable message
	Data       interface{} `json:"data,omitempty"`   // optional data
	Error      interface{} `json:"error,omitempty"`  // optional error
	Code       string      `json:"code,omitempty"`   // only included if set
}
