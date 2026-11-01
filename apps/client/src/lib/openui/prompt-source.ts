import { library } from './library';

const PREAMBLE =
	'When a structured answer is clearer than prose (a table, metrics, a chart, a form, a choice), put it in one fenced block tagged openui, and keep ordinary explanation as markdown outside the block. Use only the components listed below. Never invent components or props.';

const RULES = [
	'Open the block with a line containing exactly ```openui and close it with a line containing exactly ```.',
	'The block holds OpenUI Lang only, no prose and no markdown.',
	'Always define root = Card(...) and wrap the whole block in it.',
	'Use a Form only to collect information you need. Use a Button for a simple choice or confirmation.',
	'Action and form names are short snake_case ids. The user reply tells you which action was chosen and what values were entered.',
	'Prefer plain markdown when the answer is a short sentence or a single paragraph.'
];

export function buildPrompt() {
	return library.prompt({ preamble: PREAMBLE, additionalRules: RULES });
}
