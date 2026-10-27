// SPDX-License-Identifier: BSD-2-Clause

//go:build go1.23

package ensembleweather

import "context"

// collectCatalogLocations drains the range-over-func iterator.
func collectCatalogLocations(ctx context.Context, c *Client, each func(CatalogLocation)) error {
	for loc, err := range c.AllCatalogLocations(ctx, nil) {
		if err != nil {
			return err
		}
		each(loc)
	}
	return nil
}
