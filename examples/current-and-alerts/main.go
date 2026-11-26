// SPDX-License-Identifier: BSD-2-Clause

// Current and alerts: current conditions and the weather alerts in effect at
// a named location. No API key is needed.
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

	now, err := client.Current(ctx, "or/mt-hood-meadows", nil)
	if err != nil {
		log.Fatal(err)
	}
	if c := now.Now; c != nil {
		fmt.Printf("%s at %s: %s, %.1f %s (from %s)\n", now.Location.Name, c.Time.Format("15:04 MST"),
			c.Text, c.Temperature, now.Units.For(ew.Temperature).Symbol(), c.Source)
	} else {
		fmt.Println("no current conditions right now")
	}
	for _, s := range now.Sensors {
		fmt.Printf("sensor %s (%s) reported at %s\n", s.Name, s.Source, s.ObservedAt.Format("15:04 MST"))
	}

	alerts, err := client.Alerts(ctx, "or/mt-hood-meadows")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%d active alert(s)\n", alerts.Count)
	for _, a := range alerts.Items {
		fmt.Printf("  %s [%s]: %s (expires %s)\n", a.Event, a.Severity, a.Headline, a.Expires.Format("Mon 15:04 MST"))
	}
}
