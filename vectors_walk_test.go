// SPDX-License-Identifier: BSD-2-Clause

//go:build !go1.23

package ensembleweather

import "context"

// collectCatalogLocations drives the walker behind AllCatalogLocations
// directly: before Go 1.23 the iterator itself does not compile, and its
// body is exactly this call.
func collectCatalogLocations(ctx context.Context, c *Client, each func(CatalogLocation)) error {
	var failed error
	c.walkCatalogLocations(ctx, nil, func(loc CatalogLocation, err error) bool {
		if err != nil {
			failed = err
			return false
		}
		each(loc)
		return true
	})
	return failed
}
