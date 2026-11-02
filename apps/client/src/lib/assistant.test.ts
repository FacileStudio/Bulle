import { describe, expect, it } from 'vitest';
import { ARCHETYPES, RANDOM_NAMES, TRAITS, avatarUrlFor } from './assistant';
import { session, setAssistant } from './session.svelte';

describe('assistant', () => {
	it('generates blobatar avatar urls correctly', () => {
		expect(avatarUrlFor('')).toBe('https://blobatar.dev/avatar/bulle?gen=2');
		expect(avatarUrlFor('   ')).toBe('https://blobatar.dev/avatar/bulle?gen=2');
		expect(avatarUrlFor('Ada')).toBe('https://blobatar.dev/avatar/Ada?gen=2');
		expect(avatarUrlFor('Ada Lovelace')).toBe(
			'https://blobatar.dev/avatar/Ada%20Lovelace?gen=2'
		);
	});

	it('contains the expected base archetypes', () => {
		const labels = ARCHETYPES.map((a) => a.label);
		expect(labels).toContain('Pragmatic Senior');
		expect(labels).toContain('Terse Minimalist');
		expect(labels).toContain('Supportive Mentor');
		expect(labels).toContain('Witty Hacker');
	});

	it('contains the expected traits', () => {
		const ids = TRAITS.map((t) => t.id);
		expect(ids).toContain('concise');
		expect(ids).toContain('explain_reasoning');
		expect(ids).toContain('witty');
		expect(ids).toContain('strict');
		expect(ids).toContain('creative');
	});

	it('provides random names for the dice button', () => {
		expect(RANDOM_NAMES.length).toBeGreaterThan(5);
		expect(RANDOM_NAMES).toContain('Ada');
		expect(RANDOM_NAMES).toContain('Atlas');
	});

	it('updates session user assistant via setAssistant', () => {
		session.user = {
			id: 1,
			email: 'test@example.com',
			name: 'Tester',
			created_at: '2026-01-01T00:00:00Z',
			assistant: null
		};

		setAssistant({
			name: 'Jarvis',
			avatar_url: 'https://blobatar.dev/avatar/Jarvis?gen=2',
			archetype: 'witty_hacker',
			traits: ['witty', 'creative']
		});

		expect(session.user.assistant).toEqual({
			name: 'Jarvis',
			avatar_url: 'https://blobatar.dev/avatar/Jarvis?gen=2',
			archetype: 'witty_hacker',
			traits: ['witty', 'creative']
		});
	});
});
