package twapi

import "context"

// Ptr returns a pointer to the value v.
func Ptr[T any](v T) *T {
	return &v
}

// Iterate allows scanning through paginated results from the Teamwork API.
func Iterate[T HTTPRequester, R interface {
	HTTPResponser
	Iterate() *T
}](ctx context.Context, e *Engine, req T) (next func() (R, bool, error), err error) {
	next = func() (R, bool, error) {
		response, err := Execute[T, R](ctx, e, req)
		if err != nil {
			return response, false, err
		}

		nextRequest := response.Iterate()
		if nextRequest == nil {
			return response, false, nil
		}

		req = *nextRequest
		return response, true, nil
	}
	return next, nil
}
