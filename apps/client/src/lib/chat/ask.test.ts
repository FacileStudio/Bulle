import { describe, expect, it } from 'vitest';
import { askComplete, askQuestions, askReply, askShortcut } from './ask';

describe('askShortcut', () => {
	const questions = askQuestions({
		questions: [
			{ question: 'Which day?', options: [{ label: 'Monday' }, { label: 'Tuesday' }] },
			{ question: 'Time?', options: [{ label: 'Morning' }] }
		]
	})!;
	const key = (k: string, target: EventTarget | null = null) => ({
		key: k,
		ctrlKey: false,
		metaKey: false,
		altKey: false,
		target
	});

	it('picks an option of the first unanswered question', () => {
		const empty = [
			{ picked: [], typed: '' },
			{ picked: [], typed: '' }
		];
		expect(askShortcut(key('2'), questions, empty)).toEqual({ question: 0, label: 'Tuesday' });
		const first = [
			{ picked: ['Monday'], typed: '' },
			{ picked: [], typed: '' }
		];
		expect(askShortcut(key('1'), questions, first)).toEqual({ question: 1, label: 'Morning' });
		expect(askShortcut(key('2'), questions, first)).toBeNull();
	});

	it('ignores other keys, modifiers and typing', () => {
		const empty = [
			{ picked: [], typed: '' },
			{ picked: [], typed: '' }
		];
		expect(askShortcut(key('5'), questions, empty)).toBeNull();
		expect(askShortcut({ ...key('1'), ctrlKey: true }, questions, empty)).toBeNull();
		const input = { tagName: 'INPUT' } as unknown as EventTarget;
		expect(askShortcut(key('1', input), questions, empty)).toBeNull();
	});
});

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
