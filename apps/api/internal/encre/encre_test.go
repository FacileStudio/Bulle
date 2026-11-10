package encre_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/encre"
	"github.com/FacileStudio/Bulle/apps/api/internal/encre/encretest"
)

func TestValidateMarkdownReadsBothRefusalShapes(t *testing.T) {
	encretest.Install(t)
	ctx := context.Background()
	if err := encre.ValidateMarkdown(ctx, []byte("# Fine\n")); err != nil {
		t.Fatalf("valid body: %v", err)
	}
	var invalid *encre.ValidationError
	err := encre.ValidateMarkdown(ctx, []byte("# INVALID\n"))
	if !errors.As(err, &invalid) || !strings.Contains(err.Error(), "forbidden character") {
		t.Fatalf("verdict on stdout must become a ValidationError, got %v", err)
	}
	if invalid.Problems[0].Path != "doc.content[0]" {
		t.Fatalf("the path is lost: %+v", invalid.Problems)
	}
	err = encre.ValidateMarkdown(ctx, []byte("OVERSIZE"))
	if !errors.As(err, &invalid) || !strings.Contains(err.Error(), "larger than") {
		t.Fatalf("an input error on stderr must become a ValidationError, got %v", err)
	}
	err = encre.ValidateMarkdown(ctx, []byte("CRASH"))
	if errors.As(err, &invalid) || err == nil || !strings.Contains(err.Error(), "exit 3") {
		t.Fatalf("an internal failure is not a refusal, got %v", err)
	}
}

func TestValidateMarkdownWithoutEncre(t *testing.T) {
	t.Setenv(encre.BinaryEnv, "/nonexistent/encre")
	if encre.Available() {
		t.Fatal("a missing binary must not be available")
	}
	if err := encre.ValidateMarkdown(context.Background(), []byte("x")); !errors.Is(err, encre.ErrUnavailable) {
		t.Fatalf("got %v, want ErrUnavailable", err)
	}
}

func TestRealEncreRefusesControlCharacters(t *testing.T) {
	if !encre.Available() {
		t.Skip("encre is not installed")
	}
	ctx := context.Background()
	if err := encre.ValidateMarkdown(ctx, []byte("# Hi\n\n```chart\n{\"kind\": \"bar\", \"labels\": [\"a\"], \"series\": [{\"name\": \"s\", \"data\": [1]}]}\n```\n")); err != nil {
		t.Fatalf("a chart block is valid: %v", err)
	}
	var invalid *encre.ValidationError
	if err := encre.ValidateMarkdown(ctx, []byte("a\x01b\n")); !errors.As(err, &invalid) {
		t.Fatalf("a control character must be refused, got %v", err)
	}
}
