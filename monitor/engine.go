package monitor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/yushasama/tori/config"
	"github.com/yushasama/tori/dispatcher"
	"github.com/yushasama/tori/notifier"
)

func Start(ctx context.Context, cfg *config.Config) error {
	fmt.Println("Starting monitors...")

	dispatchers := make(map[string]*dispatcher.Dispatcher)

	for _, site := range cfg.Sites {
		if site.URL == "" || !strings.HasPrefix(site.URL, "https://") {
			panic(fmt.Sprintf("INVALID SITE URL: %s", site.URL))
		}

		if !site.Enabled {
			continue
		}

		rateLimit := 30 // Discord Rate Limit
		interval := 60 * time.Second

		d := notifier.NewDiscordNotifier(site.Notifier.WebhookURL, site.Notifier.Username, site.Notifier.AvatarURL)
		disp := dispatcher.NewDispatcher(rateLimit, interval, d)

		go disp.Run(ctx)
		dispatchers[site.Name] = disp

		for _, mon := range site.Monitors {

			go func(site config.SiteConfig, mon config.MonitorConfig) {
				fmt.Printf("LAUNCHING MONITOR: %s -> %s\n", site.Name, mon.Name)

				err := runMonitor(ctx, site, mon, cfg.GlobalPollInterval, cfg.GlobalRetryInterval, cfg.GlobalMaxRetries, disp)

				if err != nil {
					fmt.Printf("Monitor error: %v\n", err)
				}
			}(site, mon)
		}
	}

	<-ctx.Done()
	return nil
}
