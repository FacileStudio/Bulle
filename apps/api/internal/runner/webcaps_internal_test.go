package runner

import (
	"context"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/core/nacelletest"
	"github.com/FacileStudio/nacelle"
)

type pageInput struct {
	URL string `json:"url"`
}

func TestWebFetchIsCappedPerRunAndTruncated(t *testing.T) {
	page := strings.Repeat("line of text\n", fetchTokenBudget)
	fetch, err := nacelle.NewToolWithOptions("web_fetch", "fetch", func(context.Context, pageInput) (string, error) {
		return page, nil
	}, nacelle.ToolOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	tool := capWeb([]nacelle.Tool{fetch})[0]
	if !tool.Spec().ReadOnly {
		t.Fatal("capping dropped ReadOnly")
	}
	for i := range maxFetchesPerRun {
		out, err := nacelletest.Run(context.Background(), tool, []byte(`{"url":"https://example.com"}`))
		if err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
		if len(out) > fetchTokenBudget*charsPerToken+100 || !strings.Contains(out, "[truncated") {
			t.Fatalf("call %d returned %d bytes", i, len(out))
		}
	}
	if _, err := nacelletest.Run(context.Background(), tool, []byte(`{"url":"https://example.com"}`)); err == nil {
		t.Fatal("a call past the per-run cap ran")
	}
}
