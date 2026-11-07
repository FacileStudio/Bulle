package runner

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/store"
)

func TestJobToolsAdvertiseJobIDAndDeliveryEnum(t *testing.T) {
	tools := buildJobTools(store.NewMemory(), nil, 1, "", store.ChannelWeb)
	for _, name := range []string{JobsUpdateTool, JobsDeleteTool} {
		schema, _ := json.Marshal(findTool(tools, name).Schema())
		if !strings.Contains(string(schema), `"job_id"`) || strings.Contains(string(schema), `"id"`) {
			t.Errorf("%s must advertise job_id only: %s", name, schema)
		}
	}
	create := findTool(tools, JobsCreateTool)
	schema, _ := json.Marshal(create.Schema())
	if !strings.Contains(string(schema), `"enum":["none","push"]`) || !strings.Contains(create.Description(), "CRON_TZ=") {
		t.Fatalf("jobs_create must give the delivery enum and the time zone rule: %s | %s", schema, create.Description())
	}
}

func TestJobDeliveryRefusesUnknownTargets(t *testing.T) {
	tool := findTool(buildJobTools(store.NewMemory(), nil, 1, "", store.ChannelWeb), JobsCreateTool)
	_, err := tool.Run(context.Background(), json.RawMessage(`{"name":"n","schedule":"0 9 * * *","prompt":"p","delivery":"matrix"}`))
	if err == nil || !strings.Contains(err.Error(), "none or push") {
		t.Fatalf("an unknown delivery must name the valid ones, got %v", err)
	}
	out, err := tool.Run(context.Background(), json.RawMessage(`{"name":"n","schedule":"CRON_TZ=Europe/Paris 0 9 * * 1-5","prompt":"p","delivery":"none"}`))
	if err != nil || !strings.Contains(out, `"delivery":""`) {
		t.Fatalf("none must store an empty delivery and CRON_TZ must parse: %s, %v", out, err)
	}
}
