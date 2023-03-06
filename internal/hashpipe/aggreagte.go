package hashpipe

import "context"

func PipeToSlice[T any](ctx context.Context, expectedSize int, input <-chan T) ([]T, error) {
	result := make([]T, 0, expectedSize)
	for {
		select {
		case obj, ok := <-input:
			if ok {
				result = append(result, obj)
			} else {
				return result, nil
			}
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}
