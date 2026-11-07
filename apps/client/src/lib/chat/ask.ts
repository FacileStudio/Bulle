export const ASK_TOOL = 'ask_user';

export type AskOption = { label: string; description: string };

export type AskQuestion = {
	question: string;
	header: string;
	options: AskOption[];
	multiSelect: boolean;
};

export type AskAnswer = { picked: string[]; typed: string };

function text(value: unknown): string {
	return typeof value === 'string' ? value.trim() : '';
}

/**
 * Reads the questions of an ask_user call, or null when the arguments do not
 * hold at least one question with options.
 */
export function askQuestions(args: unknown): AskQuestion[] | null {
	const raw = args && typeof args === 'object' ? (args as Record<string, unknown>).questions : null;
	if (!Array.isArray(raw)) return null;
	const questions = raw.flatMap((entry): AskQuestion[] => {
		if (!entry || typeof entry !== 'object') return [];
		const q = entry as Record<string, unknown>;
		const options = (Array.isArray(q.options) ? q.options : [])
			.map((o) => (o && typeof o === 'object' ? (o as Record<string, unknown>) : {}))
			.map((o) => ({ label: text(o.label), description: text(o.description) }))
			.filter((o) => o.label);
		const question = text(q.question);
		if (!question || !options.length) return [];
		return [{ question, header: text(q.header), options, multiSelect: q.multi_select === true }];
	});
	return questions.length ? questions : null;
}

/**
 * Turns the user's picks into the message sent back to the assistant. A
 * typed answer wins over picks for a single-choice question and joins them
 * for a multi-choice one. One question sends only its answer.
 */
export function askReply(questions: AskQuestion[], answers: AskAnswer[]): string {
	const lines = questions.map((q, i) => {
		const answer = answers[i] ?? { picked: [], typed: '' };
		const typed = answer.typed.trim();
		const parts = q.multiSelect ? [...answer.picked, typed] : [typed || answer.picked[0] || ''];
		return { q, value: parts.filter(Boolean).join(', ') };
	});
	if (lines.length === 1) return lines[0].value;
	return lines.map(({ q, value }) => `${q.question} ${value || '(no answer)'}`).join('\n');
}

/** Reports whether every question has a pick or a typed answer. */
export function askComplete(answers: AskAnswer[]): boolean {
	return answers.every((a) => a.picked.length > 0 || a.typed.trim() !== '');
}
