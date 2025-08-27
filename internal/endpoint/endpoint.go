package endpoint

type HTTPRequest[T any] struct {
	Body    T
	Headers map[string]string
}

type HTTPResponse[T any] struct {
	Body    T
	Status  int
	Headers map[string]string
}
