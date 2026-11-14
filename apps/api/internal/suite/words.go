package suite

import (
	"slices"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// appWords are the words, in English and French and folded to plain
// lowercase letters, a person or a model uses for each app's job. A word
// names an app when it is one of them or its plural, so "log" names Journal
// and "login" does not.
var appWords = map[string][]string{
	AppAgenda:   {"agenda", "calendar", "calendrier", "event", "evenement", "rdv", "rendezvous", "appointment", "reunion", "schedule", "planning"},
	AppCourrier: {"courrier", "courriel", "mail", "email", "inbox", "boite"},
	AppNuage:    {"nuage", "file", "fichier", "folder", "dossier", "drive", "storage", "stockage"},
	AppOpus:     {"opus", "task", "tache", "todo", "project", "projet"},
	AppSablier:  {"sablier", "time", "temps", "timer", "chrono", "timesheet", "heure", "hour"},
	AppGlouton:  {"glouton", "contact", "prospect", "crm", "company", "entreprise", "societe", "lead"},
	AppSonar:    {"sonar", "web", "internet", "google"},
	AppPlume:    {"plume", "signature", "sign", "contrat", "contract"},
	AppCapsule:  {"capsule", "encrypted", "chiffre"},
	AppJournal:  {"journal", "log"},
	AppAntenne:  {"antenne", "alert", "alerte"},
	AppVision:   {"vision", "analytics", "statistique", "stats", "traffic", "trafic", "visite", "visitor"},
	AppEcho:     {"echo", "visio", "video", "transcript", "transcription", "appel"},
	AppToile:    {"toile", "whiteboard", "board", "tableau", "dessin"},
	AppDocs:     {"docs", "document", "note", "page", "doc"},
	AppFeuilles: {"feuilles", "tableur", "spreadsheet", "sheet", "classeur", "workbook"},
}

// Words splits text into folded lowercase words: accents dropped, and
// hyphens and apostrophes inside a word removed, so "rendez-vous" is one word.
func Words(text string) []string {
	folded, _, err := transform.String(transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC), strings.ToLower(text))
	if err != nil {
		folded = strings.ToLower(text)
	}
	folded = strings.NewReplacer("-", "", "'", "", "’", "").Replace(folded)
	return strings.FieldsFunc(folded, func(r rune) bool { return !unicode.IsLetter(r) && !unicode.IsDigit(r) })
}

// AppsNamed is the apps a folded word names, sorted.
func AppsNamed(word string) []string {
	var apps []string
	for app, words := range appWords {
		if slices.ContainsFunc(words, func(name string) bool { return SameWord(word, name) }) {
			apps = append(apps, app)
		}
	}
	slices.Sort(apps)
	return apps
}

// SameWord reports whether two folded words are one word, or one is the
// other's plural in English or French.
func SameWord(a, b string) bool {
	return a == b || a == b+"s" || a == b+"x" || b == a+"s" || b == a+"x"
}
