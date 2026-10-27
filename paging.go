// SPDX-License-Identifier: BSD-2-Clause

package ensembleweather

import "context"

// paginate fetches pages from cursor "" until the next cursor is nil or
// empty, yielding each item. A fetch error is yielded once and ends the walk.
func paginate[T any](ctx context.Context, cursor string, fetch func(ctx context.Context, cursor string) ([]T, *string, error), yield func(T, error) bool) {
	for {
		items, next, err := fetch(ctx, cursor)
		if err != nil {
			var zero T
			yield(zero, err)
			return
		}
		for _, it := range items {
			if !yield(it, nil) {
				return
			}
		}
		if next == nil || *next == "" {
			return
		}
		cursor = *next
	}
}

func (c *Client) walkLocations(ctx context.Context, params *LocationsParams, yield func(LocationSummary, error) bool) {
	p := LocationsParams{}
	if params != nil {
		p = *params
	}
	paginate(ctx, p.Cursor, func(ctx context.Context, cursor string) ([]LocationSummary, *string, error) {
		p.Cursor = cursor
		page, err := c.Locations(ctx, &p)
		if err != nil {
			return nil, nil, err
		}
		return page.Locations, page.NextCursor, nil
	}, yield)
}

func (c *Client) walkCatalogLocations(ctx context.Context, params *CatalogLocationsParams, yield func(CatalogLocation, error) bool) {
	p := CatalogLocationsParams{}
	if params != nil {
		p = *params
	}
	paginate(ctx, p.Cursor, func(ctx context.Context, cursor string) ([]CatalogLocation, *string, error) {
		p.Cursor = cursor
		page, err := c.CatalogLocations(ctx, &p)
		if err != nil {
			return nil, nil, err
		}
		return page.Items, page.NextCursor, nil
	}, yield)
}

func (c *Client) walkCatalogStations(ctx context.Context, params *CatalogStationsParams, yield func(CatalogStation, error) bool) {
	p := CatalogStationsParams{}
	if params != nil {
		p = *params
	}
	paginate(ctx, p.Cursor, func(ctx context.Context, cursor string) ([]CatalogStation, *string, error) {
		p.Cursor = cursor
		page, err := c.CatalogStations(ctx, &p)
		if err != nil {
			return nil, nil, err
		}
		return page.Items, page.NextCursor, nil
	}, yield)
}
