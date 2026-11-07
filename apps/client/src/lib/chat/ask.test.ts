import { describe, expect, it } from 'vitest';
import { askComplete, askQuestions, askReply } from './ask';

const day = {
	question: 'Which day?',
	options: [{ label: 'Monday' }, { label: 'Tuesday', description: 'busier' }]
};

describe('ask_user', () => {
	it('parses questions and drops malformed ones', () => {
		const questions = askQuestions({ questions: [day, { question: 'Empty?', options: [] }, 'junk'] });
		expect(questions).toHaveLength(1);
		expect(questions?.[0].options[1]).toEqual({ label: 'Tuesday', description: 'busier' });
		expect(askQuestions({})).toBeNull();
	});

	it('sends a lone answer bare and several as lines, typed text first', () => {
		const tone = { question: 'Tone?', options: [{ label: 'Warm' }, { label: 'Dry' }], multi_select: true };
		const qs = askQuestions({ questions: [day] })!;
		expect(askReply(qs, [{ picked: ['Monday'], typed: '' }])).toBe('Monday');
		expect(askReply(qs, [{ picked: ['Monday'], typed: 'Friday at 3' }])).toBe('Friday at 3');
		const both = askQuestions({ questions: [day, tone] })!;
		expect(askReply(both, [{ picked: ['Tuesday'], typed: '' }, { picked: ['Warm', 'Dry'], typed: 'short' }])).toBe(
			'Which day? Tuesday\nTone? Warm, Dry, short'
		);
		expect(askComplete([{ picked: [], typed: ' ' }])).toBe(false);
	});
});
