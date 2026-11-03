package settings

import (
	"strings"
	"testing"
)

// keyFailureCases are the ways a command can fail to produce a key. Each is a
// load error naming the field it was written in, never an empty key: an empty
// key reaches the backend as "no credential" and reads exactly like a config
// that forgot one, which is the failure this setting exists to make impossible.
var keyFailureCases = []struct {
	name string
	file string
	want string
}{
	{
		name: "a judge command that prints nothing",
		file: "limits:\n  compaction:\n    judge:\n      api_key_command: \"true\"\n",
		want: "limits.compaction.judge.api_key_command",
	},
	{
		name: "a judge command that fails",
		file: "limits:\n  compaction:\n    judge:\n      api_key_command: \"exit 3\"\n",
		want: "limits.compaction.judge.api_key_command",
	},
}

// The judge's command fills its own field.
func TestJudgeKeyCommandReachesItsField(t *testing.T) {
	clearEnv(t, "TYPESAFE_API_KEY", "BULLE_COMPACTION_JUDGE_API_KEY", "NACELLE_COMPACTION_JUDGE_API_KEY")
	written(t, "limits:\n  compaction:\n    judge:\n      enabled: true\n      api_key_command: \"echo sk-judge\"\n")
	cfg, err := settings(Config{})
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	if err := resolveJudgeKey(&cfg); err != nil {
		t.Fatalf("resolveJudgeKey: %v", err)
	}
	if got := cfg.Compaction.Judge.APIKey; got != "sk-judge" {
		t.Error("the judge key is not the judge command's answer")
	}
}

// With no api_key anywhere, ResolveKeys fills the judge's key from the vendor
// variable the merged provider names — so a file that says clef needs no
// command at all on a machine that already exports OPENROUTER_API_KEY, and the
// command stays the last resort.
func TestResolveKeysFillsTheJudgeKeyFromTheProviderVendor(t *testing.T) {
	clearEnv(t, "TYPESAFE_API_KEY", "BULLE_COMPACTION_JUDGE_API_KEY", "NACELLE_COMPACTION_JUDGE_API_KEY")
	t.Setenv("OPENROUTER_API_KEY", "sk-openrouter")
	written(t, "limits:\n  compaction:\n    judge:\n      enabled: true\n      provider: clef\n")

	cfg, err := settings(Config{})
	if err != nil {
		t.Fatalf("settings: %v", err)
	}
	if err := resolveJudgeKey(&cfg); err != nil {
		t.Fatalf("resolveJudgeKey: %v", err)
	}
	if got := cfg.Compaction.Judge.APIKey; got != "sk-openrouter" {
		t.Errorf("judge key = %q, want the provider's vendor variable", got)
	}
}

// A command that cannot produce a key is refused at load, naming where it was
// written, rather than leaving a credential-shaped hole behind.
func TestAFailingKeyCommandNamesItsField(t *testing.T) {
	clearEnv(t, "TYPESAFE_API_KEY", "BULLE_COMPACTION_JUDGE_API_KEY", "NACELLE_COMPACTION_JUDGE_API_KEY")
	for _, tt := range keyFailureCases {
		t.Run(tt.name, func(t *testing.T) {
			written(t, tt.file)
			cfg, err := settings(Config{})
			if err != nil {
				t.Fatalf("settings: %v", err)
			}
			err = resolveJudgeKey(&cfg)
			if err == nil {
				t.Fatal("a command that produces no key was accepted, want a load error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %q, want it to name %s", err, tt.want)
			}
		})
	}
}

// The command is a shell program, not a name and an argument list: `tiroir
// unlock && tiroir get K` and a pipeline both have to work, and a `get` that
// prints its answer with a trailing newline is the ordinary case rather than one
// to special-case at every call site.
func TestKeyFromCommandRunsThroughAShell(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    string
	}{
		{name: "no command is no key", command: "", want: ""},
		{name: "the trailing newline is not part of the key", command: "echo sk-newline", want: "sk-newline"},
		{name: "a pipeline runs as written", command: "printf sk-piped | tr a-z A-Z", want: "SK-PIPED"},
		{name: "two commands in sequence", command: "true && echo sk-chained", want: "sk-chained"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := KeyFromCommand(tt.command)
			if err != nil {
				t.Fatalf("KeyFromCommand(%q): %v", tt.command, err)
			}
			if got != tt.want {
				t.Errorf("KeyFromCommand(%q) = %q, want %q", tt.command, got, tt.want)
			}
		})
	}
}
