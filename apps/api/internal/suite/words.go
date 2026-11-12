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
// names an app when it starts with one of them, so plurals match too.
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
	AppVision:   {"vision", "analytics", "statistique", "stats", "traf", "visite", "visitor"},
	AppEcho:     {"echo", "visio", "video", "transcript", "transcription", "appel"},
	AppToile:    {"toile", "whiteboard", "board", "tableau", "dessin"},
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
		if slices.ContainsFunc(words, func(name string) bool { return strings.HasPrefix(word, name) }) {
			apps = append(apps, app)
		}
	}
	slices.Sort(apps)
	return apps
}
