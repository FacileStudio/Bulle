package runner

import (
	"context"

	"github.com/FacileStudio/Bulle/apps/api/internal/credentials"
)

// resolveSearchKeys answers the key each search engine would use for the user
// and which of those engines run on an instance key. A stored key that equals
// the instance key counts as the instance's.
func resolveSearchKeys(ctx context.Context, creds credentials.Store, userID int64) (map[string]string, []string, error) {
	own, err := ownSearchKeys(ctx, creds, userID)
	if err != nil {
		return nil, nil, err
	}
	keys := map[string]string{}
	var server []string
	for _, engine := range credentials.SearchEngines() {
		instance := credentials.EnvKey(engine)
		if key := own[engine]; key != "" && key != instance {
			keys[engine] = key
			continue
		}
		if instance != "" {
			keys[engine] = instance
			server = append(server, engine)
		}
	}
	return keys, server, nil
}

func ownSearchKeys(ctx context.Context, creds credentials.Store, userID int64) (map[string]string, error) {
	own := map[string]string{}
	if creds == nil {
		return own, nil
	}
	entries, err := creds.List(ctx, userID)
	if err != nil {
		return nil, err
	}
	search := map[string]bool{}
	for _, engine := range credentials.SearchEngines() {
		search[engine] = true
	}
	for _, entry := range entries {
		if !search[entry.Backend] {
			continue
		}
		key, err := creds.Resolve(ctx, userID, entry.Backend)
		if err != nil {
			return nil, err
		}
		own[entry.Backend] = key
	}
	return own, nil
}
