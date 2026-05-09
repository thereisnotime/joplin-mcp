package joplin

import "context"

// FetchPage fetches a single page of T from a paginated endpoint.
type FetchPage[T any] func(ctx context.Context, page int) (Page[T], error)

// CollectAll walks every page returned by fetch and concatenates the items.
//
// Stops at the first error or when has_more is false. Use this when you actually
// want the entire result set; for selective access prefer calling FetchPage
// directly with a specific page number.
func CollectAll[T any](ctx context.Context, fetch FetchPage[T]) ([]T, error) {
	var all []T
	page := 1
	for {
		p, err := fetch(ctx, page)
		if err != nil {
			return nil, err
		}
		all = append(all, p.Items...)
		if !p.HasMore {
			return all, nil
		}
		page++
	}
}
