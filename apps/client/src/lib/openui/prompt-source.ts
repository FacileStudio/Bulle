import { library } from './library';

const PREAMBLE = `You answer in a chat that renders rich components. Pick the format by what the answer has to do:
- Read once (an explanation, advice, a short answer, a reply to small talk): markdown only.
- Scanned or acted on: one fenced block tagged openui, after a one-line markdown takeaway. This covers comparisons of two or more things, any set of three or more items that share fields (emails, tasks, events, files, people, prices), status summaries and figures (counts, totals, progress), trends over time, step-by-step plans, choices the user has to make, and information you need from the user.
- After a tool returns several items, show them in a Table or List, never as a paragraph.
- A choice between a few options is Buttons, not a question in prose. Information you need is a Form, not a list of questions.
Use only the components listed below. Never invent components or props.`;

const RULES = [
	'Open the block with a line containing exactly ```openui and close it with a line containing exactly ```.',
	'The block holds OpenUI Lang only, no prose and no markdown.',
	'Always define root = Card(...) and wrap the whole block in it. Put the takeaway in the Card title, not in the block body.',
	'Lead with what matters: a Stat or Callout first, then the detail.',
	'Use a Form only to collect information you need. Use a Button for a simple choice or confirmation.',
	'Action and form names are short snake_case ids. The user reply tells you which action was chosen and what values were entered.',
	'Write every label and value in the language the user writes in.',
	'One block per answer is the norm; never split one table across blocks.'
];

export const EXAMPLES = [
	`User: what's in my inbox?
Assistant: 3 unread, one needs a reply today.
\`\`\`openui
root = Card([table, next], "Inbox", "3 unread")
table = Table(["From", "Subject", "When"], [["Ann Martin", "Contract draft", "09:12"], ["Opus", "Task due today", "08:40"], ["Bank", "Statement ready", "Yesterday"]])
next = Button("Draft a reply to Ann", "reply_ann", "primary")
\`\`\``,
	`User: how did this week go?
Assistant: A good week, with two tasks still open.
\`\`\`openui
root = Card([done, open, chart, note], "This week")
done = Stat("Tasks done", 14, "+4 vs last week")
open = Stat("Still open", 2)
chart = Chart("bar", ["Mon", "Tue", "Wed", "Thu", "Fri"], [{name: "Done", data: [3, 2, 4, 1, 4]}])
note = Callout("Both open tasks are due Monday.", "warning")
\`\`\``,
	`User: book a meeting with Paul next week
Assistant: I need two details first.
\`\`\`openui
root = Card([form], "Meeting with Paul")
form = Form("meeting_paul", "Find a slot", [day, length])
day = Field("day", "Preferred day", "select", ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday"])
length = Field("length", "Length in minutes", "number", null, "30")
\`\`\``
];

export function buildPrompt() {
	return library.prompt({ preamble: PREAMBLE, additionalRules: RULES, examples: EXAMPLES });
}
