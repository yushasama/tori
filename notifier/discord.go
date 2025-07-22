package notifier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/yushasama/tori/types"
)

type DiscordNotifier struct {
	WebhookURL string
	Username   string
	AvatarURL  string
}

func NewDiscordNotifier(webhookURL, username, avatarURL string) *DiscordNotifier {
	return &DiscordNotifier{
		WebhookURL: webhookURL,
		Username:   username,
		AvatarURL:  avatarURL,
	}
}

var _ Notifier = (*DiscordNotifier)(nil)

func (d *DiscordNotifier) Notify(job *types.Job) {
	embed := map[string]interface{}{
		"title": "Product Found",
		"url":   fmt.Sprintf("%s/cart/add?id=%s", job.Site, job.ID),
		"color": 16777214,
		"fields": []map[string]string{
			{"name": "Price", "value": job.Price},
			{"name": "Product", "value": job.Product},
			{"name": "Variant", "value": job.Variant},
			{"name": "ID", "value": job.ID},
		},
		"footer": map[string]string{
			"text":     "Tori",
			"icon_url": d.AvatarURL,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if job.ImageURL != "" {
		embed["image"] = map[string]string{"url": job.ImageURL}
	}

	payload := map[string]interface{}{
		"username":   d.Username,
		"avatar_url": d.AvatarURL,
		"embeds":     []interface{}{embed},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("FAILED TO ENCODE PAYLOAD: %v\n", err)
		return
	}

	resp, err := http.Post(d.WebhookURL, "application/json", bytes.NewBuffer(body))

	if err != nil {
		fmt.Printf("FAILED TO SEND WEBHOOK: %v\n", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		fmt.Printf("DISCORD WEBHOOK RETURNED %d\n", resp.StatusCode)
	}
}
