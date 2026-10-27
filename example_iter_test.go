// SPDX-License-Identifier: BSD-2-Clause

//go:build go1.23

package ensembleweather_test

import (
	"context"
	"fmt"
	"log"

	ensembleweather "oureawx.com/go/ensemble-weather"
)

// Walk the whole catalog; the iterator follows the cursor.
func ExampleClient_AllCatalogLocations() {
	client := ensembleweather.NewClient()
	for loc, err := range client.AllCatalogLocations(context.Background(), &ensembleweather.CatalogLocationsParams{Admin1: "US-OR"}) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(loc.Key(), loc.Name)
	}
}
