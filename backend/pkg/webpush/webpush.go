// Package webpush sends Web Push notifications (RFC 8291/8292) to browser
// push endpoints, so a subscribed device shows a notification on its
// lock screen / notification tray even while the app/tab is closed.
package webpush

import (
	"encoding/json"
	"fmt"

	webpushgo "github.com/SherClockHolmes/webpush-go"
)

// Config holds the VAPID key pair used to authenticate this server to the
// browser push services (Chrome/FCM, Firefox/Mozilla, etc). Generate a pair
// once with webpush-go's GenerateVAPIDKeys and keep the private key secret.
type Config struct {
	Enabled    bool
	PublicKey  string
	PrivateKey string
	// Subject identifies the sender to the push service, e.g. "mailto:ops@example.com".
	Subject string
}

// Sender sends a push payload to one subscription endpoint.
type Sender interface {
	Send(endpoint, p256dh, auth string, payload Payload) error
}

// Payload is the JSON body delivered to the service worker's `push` event
// (see src/app/sw.ts on the frontend).
type Payload struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Link  string `json:"link,omitempty"`
	Tag   string `json:"tag,omitempty"`
}

type sender struct {
	cfg Config
}

func NewSender(cfg Config) Sender {
	return &sender{cfg: cfg}
}

func (s *sender) Send(endpoint, p256dh, auth string, payload Payload) error {
	if !s.cfg.Enabled {
		return fmt.Errorf("web push is disabled: VAPID keys are not configured")
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	sub := &webpushgo.Subscription{
		Endpoint: endpoint,
		Keys:     webpushgo.Keys{P256dh: p256dh, Auth: auth},
	}

	resp, err := webpushgo.SendNotification(body, sub, &webpushgo.Options{
		Subscriber:      s.cfg.Subject,
		VAPIDPublicKey:  s.cfg.PublicKey,
		VAPIDPrivateKey: s.cfg.PrivateKey,
		TTL:             60 * 60 * 24, // 1 day: drop the push if the device stays offline longer than this
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// 404/410 mean the subscription is gone (browser data cleared, uninstalled,
	// etc) — the caller should delete it so it stops trying.
	if resp.StatusCode == 404 || resp.StatusCode == 410 {
		return ErrSubscriptionExpired
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("push service returned status %d", resp.StatusCode)
	}
	return nil
}

// ErrSubscriptionExpired signals the endpoint is no longer valid and should
// be removed from storage.
var ErrSubscriptionExpired = fmt.Errorf("push subscription expired or invalid")
