package endpoint

// HTTPRequest represents a generic HTTP request wrapper for service method inputs
type HTTPRequest[T any] struct {
	// Body contains the request payload
	Body T `json:"body"`
	
	// Headers contains HTTP headers (for future extensibility)
	Headers map[string]string `json:"headers,omitempty"`
	
	// Metadata contains additional request context (for future extensibility)  
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// HTTPResponse represents a generic HTTP response wrapper for service method outputs
type HTTPResponse[T any] struct {
	// Data contains the response payload
	Data T `json:"data"`
	
	// StatusCode contains the HTTP status code (for future extensibility)
	StatusCode int `json:"statusCode,omitempty"`
	
	// Headers contains HTTP response headers (for future extensibility)
	Headers map[string]string `json:"headers,omitempty"`
	
	// Metadata contains additional response context (for future extensibility)
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// NewHTTPRequest creates a new HTTPRequest with the given body
func NewHTTPRequest[T any](body T) *HTTPRequest[T] {
	return &HTTPRequest[T]{
		Body:     body,
		Headers:  make(map[string]string),
		Metadata: make(map[string]interface{}),
	}
}

// NewHTTPResponse creates a new HTTPResponse with the given data
func NewHTTPResponse[T any](data T) *HTTPResponse[T] {
	return &HTTPResponse[T]{
		Data:     data,
		Headers:  make(map[string]string),
		Metadata: make(map[string]interface{}),
	}
}
