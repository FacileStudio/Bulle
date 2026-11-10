package encre_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/FacileStudio/Bulle/apps/api/internal/encre"
	"github.com/FacileStudio/Bulle/apps/api/internal/encre/encretest"
)

const placed = "invalid document: line 3, column 1: content[0]: text is empty or holds a forbidden character"

func TestValidateMarkdownReadsBothRefusalShapes(t *testing.T) {
	encretest.Install(t)
	ctx := context.Background()
	if err := encre.ValidateMarkdown(ctx, []byte("# Fine\n")); err != nil {
		t.Fatalf("valid body: %v", err)
	}
	var invalid *encre.ValidationError
	err := encre.ValidateMarkdown(ctx, []byte("# INVALID\n"))
	if !errors.As(err, &invalid) || err.Error() != placed {
		t.Fatalf("verdict on stdout must become a ValidationError naming its place, got %v", err)
	}
	if p := invalid.Problems[0]; p.Path != "doc.content[1]" || p.Line != 3 || p.Column != 1 {
		t.Fatalf("the place is lost: %+v", invalid.Problems)
	}
	err = encre.ValidateMarkdown(ctx, []byte("OVERSIZE"))
	if !errors.As(err, &invalid) || !strings.Contains(err.Error(), "larger than") || strings.Contains(err.Error(), "line ") {
		t.Fatalf("an input error on stderr must become a ValidationError without a place, got %v", err)
	}
	err = encre.ValidateMarkdown(ctx, []byte("CRASH"))
	if errors.As(err, &invalid) || err == nil || !strings.Contains(err.Error(), "exit 3") {
		t.Fatalf("an internal failure is not a refusal, got %v", err)
	}
}

func TestValidateMarkdownCapsTheBodyBeforeEncre(t *testing.T) {
	log := encretest.Install(t)
	ctx := context.Background()
	body := bytes.Repeat([]byte("a"), encre.MaxBytes+1)
	if err := encre.ValidateMarkdown(ctx, body); !errors.Is(err, encre.ErrTooLarge) {
		t.Fatalf("got %v, want ErrTooLarge", err)
	}
	if inputs := encretest.Inputs(t, log); len(inputs) != 0 {
		t.Fatalf("encre must not see a body over the cap, got %d calls", len(inputs))
	}
	if err := encre.ValidateMarkdown(ctx, body[:encre.MaxBytes]); err != nil {
		t.Fatalf("a body at the cap is validated: %v", err)
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

const schema2 = "# Hi\n\n````columns\n```column title=\"Pros\" tone=\"success\"\n- fast\n```\n```column title=\"Cons\" tone=\"danger\"\n- costly\n```\n````\n\n" +
	"Ship [badge:success Shipped] and [status:pending] with ==mark==.\n\n```mermaid\nflowchart LR\n  A-->B\n```\n\n" +
	"```stat\n{\"label\": \"Users\", \"value\": \"42\", \"sub\": \"this week\"}\n```\n\n> [!TIP]\n> Read it\n"

func TestRealEncreNamesThePlaceOfARefusal(t *testing.T) {
	if !encre.Available() {
		t.Skip("encre is not installed")
	}
	ctx := context.Background()
	if err := encre.ValidateMarkdown(ctx, []byte(schema2)); err != nil {
		t.Fatalf("every schema 2 construct is valid: %v", err)
	}
	var invalid *encre.ValidationError
	err := encre.ValidateMarkdown(ctx, []byte("# Hi\n\na\x01b\n"))
	if !errors.As(err, &invalid) || !strings.HasPrefix(err.Error(), "invalid document: line 3, column 1: ") {
		t.Fatalf("a control character must be refused at its place, got %v", err)
	}
}
