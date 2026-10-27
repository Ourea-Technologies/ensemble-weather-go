// SPDX-License-Identifier: BSD-2-Clause

//go:build go1.23

package ensembleweather

import (
	"context"
	"iter"
)

// AllLocations iterates over every named location that params select,
// following NextCursor page by page. A failed page ends the iteration with
// its error:
//
//	for loc, err := range client.AllLocations(ctx, &ensembleweather.LocationsParams{State: "or"}) {
//		if err != nil {
//			return err
//		}
//		fmt.Println(loc.Key())
//	}
//
// It needs Go 1.23; on older Go, call [Client.Locations] with the cursor.
func (c *Client) AllLocations(ctx context.Context, params *LocationsParams) iter.Seq2[LocationSummary, error] {
	return func(yield func(LocationSummary, error) bool) { c.walkLocations(ctx, params, yield) }
}

// AllCatalogLocations iterates over the location catalog, following
// NextCursor page by page. A failed page ends the iteration with its error.
// It needs Go 1.23; on older Go, call [Client.CatalogLocations] with the cursor.
func (c *Client) AllCatalogLocations(ctx context.Context, params *CatalogLocationsParams) iter.Seq2[CatalogLocation, error] {
	return func(yield func(CatalogLocation, error) bool) { c.walkCatalogLocations(ctx, params, yield) }
}

// AllCatalogStations iterates over the station catalog, following NextCursor
// page by page. A failed page ends the iteration with its error.
// It needs Go 1.23; on older Go, call [Client.CatalogStations] with the cursor.
func (c *Client) AllCatalogStations(ctx context.Context, params *CatalogStationsParams) iter.Seq2[CatalogStation, error] {
	return func(yield func(CatalogStation, error) bool) { c.walkCatalogStations(ctx, params, yield) }
}
