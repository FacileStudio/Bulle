package main

import (
	"context"
	"log/slog"

	"github.com/FacileStudio/Bulle/apps/api/internal/env"
	"github.com/FacileStudio/Bulle/apps/api/internal/push"
	"gorm.io/gorm"
)

// newPushHub wires ntfy, FCM and Web Push from the environment. A channel
// whose settings are missing or broken is left off with a warning, so a bad
// key never stops the server.
func newPushHub(db *gorm.DB, cfg env.Push, logger *slog.Logger) *push.Hub {
	hub := push.HubConfig{
		Ntfy:    push.New(push.Config{URL: cfg.URL, Token: cfg.Token, TopicSecret: cfg.TopicSecret}, logger),
		Web:     push.NewWebPush(cfg.VAPIDPublic, cfg.VAPIDPrivate, cfg.VAPIDSubject),
		Devices: push.NewPGDevices(db),
		Logger:  logger,
	}
	if cfg.FCMCredentials != "" {
		hub.FCM = newFCM(cfg.FCMCredentials, logger)
	} else {
		logger.Warn("FCM_SERVICE_ACCOUNT is not set: native Android notifications are off")
	}
	if hub.Web == nil {
		logger.Warn("VAPID_PUBLIC_KEY or VAPID_PRIVATE_KEY is not set: browser notifications are off")
	}
	return push.NewHub(hub)
}

func newFCM(credentials string, logger *slog.Logger) *push.FCM {
	raw, err := push.DecodeCredentials(credentials)
	if err == nil {
		var fcm *push.FCM
		if fcm, err = push.NewFCM(context.Background(), raw); err == nil {
			return fcm
		}
	}
	logger.Error("FCM_SERVICE_ACCOUNT is unusable: native Android notifications are off", slog.Any("error", err))
	return nil
}
