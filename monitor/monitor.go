package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yushasama/tori/config"
	"github.com/yushasama/tori/dispatcher"
	"github.com/yushasama/tori/types"
)

func fetchProducts(siteName string, mon config.MonitorConfig) (*ProductsWrapper, error) {
	resp, err := http.Get(mon.EndpointURL)

	if err != nil {
		fmt.Printf("HTTP GET FAILED FOR %s: %v\n", mon.EndpointURL, err)
		return nil, err
	}

	defer resp.Body.Close()

	var wrapper ProductsWrapper

	if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
		fmt.Printf("FAILED TO DECODE JSON FROM %s: %v\n", mon.EndpointURL, err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("NON-200 RESPONSE: %d\n", resp.StatusCode)
		return nil, fmt.Errorf("BAD STATUS: %d", resp.StatusCode)
	}

	fmt.Printf("FETCHED %d PRODUCTS FOR MONITOR: %s - %s\n", len(wrapper.Products), siteName, mon.Name)

	return &wrapper, nil
}

func logDuration(start time.Time, label string) {
	elapsed := time.Since(start)
	fmt.Printf("%s: %d ns (%0.9f seconds)\n", label, elapsed.Nanoseconds(), elapsed.Seconds())
}

func runMonitor(ctx context.Context, site config.SiteConfig, mon config.MonitorConfig, pollInterval time.Duration, retry_interval time.Duration, max_retries int, disp *dispatcher.Dispatcher) error {
	watchlist := make(map[int64]struct{})

	for _, pid := range mon.ProductIDs {
		watchlist[pid.ID] = struct{}{}
	}

	seen := map[int64]bool{}

	runOnce := func() {
		start := time.Now()

		var wrapper *ProductsWrapper
		var err error

		for attempts := 0; attempts < max_retries; attempts++ {
			wrapper, err = fetchProducts(site.Name, mon)

			if err == nil {
				break
			}

			time.Sleep(retry_interval)
		}

		if err != nil {
			fmt.Printf("FAILED TO FETCH AFTER %d ATTEMPTS: %v\n", max_retries, err)
			return
		}

		for _, product := range wrapper.Products {

			for _, variant := range product.Variants {
				_, match := watchlist[variant.ID]

				if match {
					fmt.Printf("MATCHED VARIANT %s - ID: %d)\n", variant.Title, variant.ID)
				}

				if match && variant.Available && !seen[variant.ID] {
					seen[variant.ID] = true

					logDuration(start, fmt.Sprintf("MATCH FOUND FOR VARIANT %d", variant.ID))

					img := ""

					if variant.Image != nil {
						img = variant.Image.Src
					}

					disp.Submit(&types.Job{
						Site:     site.URL,
						Monitor:  mon.Name,
						Product:  product.Title,
						Variant:  variant.Title,
						ID:       fmt.Sprint(variant.ID),
						ImageURL: img,
						Price:    variant.Price,
					})
				}
			}
		}
	}

	runOnce()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():

			return nil

		case <-ticker.C:
			runOnce()
		}
	}
}
