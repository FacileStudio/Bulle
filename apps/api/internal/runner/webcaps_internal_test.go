package runner

import (
	"context"
	"testing"

	"github.com/FacileStudio/Bulle/core/nacelletest"
	"github.com/FacileStudio/nacelle"
)

type pageInput struct {
	URL string `json:"url"`
}

func TestSonarFetchIsCappedPerRun(t *testing.T) {
	fetch, err := nacelle.NewToolWithOptions("sonar_fetch", "fetch", func(context.Context, pageInput) (string, error) {
		return "page", nil
	}, nacelle.ToolOptions{ReadOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	tool := capWeb([]nacelle.Tool{fetch})[0]
	if !tool.Spec().ReadOnly {
		t.Fatal("capping dropped ReadOnly")
	}
	for i := range webCalls["sonar_fetch"] {
		if _, err := nacelletest.Run(context.Background(), tool, []byte(`{"url":"https://example.com"}`)); err != nil {
			t.Fatalf("call %d: %v", i, err)
		}
	}
	if _, err := nacelletest.Run(context.Background(), tool, []byte(`{"url":"https://example.com"}`)); err == nil {
		t.Fatal("a call past the per-run cap ran")
	}
}
