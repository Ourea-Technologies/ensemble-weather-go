// SPDX-License-Identifier: BSD-2-Clause

// Search and catalog: find a location by name, then page through the
// location catalog by following its cursor. (On Go 1.23+,
// client.AllCatalogLocations does the paging in a range loop.)
package main

import (
	"context"
	"fmt"
	"log"

	ew "oureawx.com/go/ensemble-weather"
)

func main() {
	ctx := context.Background()
	client := ew.NewClient()

	hits, err := client.Search(ctx, "hood meadows", &ew.SearchParams{State: "or", Limit: 5})
	if err != nil {
		log.Fatal(err)
	}
	for _, h := range hits.Items {
		fmt.Printf("found %s (%s, %s County) at %.0f m\n", h.Name, h.Key, h.County, h.ElevationM)
	}

	params := &ew.CatalogLocationsParams{Admin1: "US-OR", Limit: 100}
	for pages := 1; ; pages++ {
		page, err := client.CatalogLocations(ctx, params)
		if err != nil {
			log.Fatal(err)
		}
		for _, loc := range page.Items {
			fmt.Printf("page %d: %s (%s)\n", pages, loc.Name, loc.Type)
		}
		if page.NextCursor == nil {
			break
		}
		params.Cursor = *page.NextCursor
	}
}
