import { blobUri } from '@facile/blob';
import { describe, expect, it } from 'vitest';
import {
	archetypeTint,
	hasTrait,
	initialDraft,
	profileFromDraft,
	randomName,
	toggleTrait
} from './profile';
import { composeSample, describeProfile } from './sample';
import {
	STEPS,
	advance,
	applySkip,
	clampStep,
	isLastStep,
	needsOnboarding,
	primaryLabel,
	progressText,
	retreat
} from './steps';
import { DRAFT_KEY, clearDraft, loadDraft, saveDraft, type DraftStore } from './storage';

function memoryStore(): DraftStore {
	const data = new Map<string, string>();
	return {
		getItem: (k) => data.get(k) ?? null,
		setItem: (k, v) => void data.set(k, v),
		removeItem: (k) => void data.delete(k)
	};
}

const brokenStore: DraftStore = {
	getItem() {
		throw new Error('blocked');
	},
	setItem() {
		throw new Error('quota');
	},
	removeItem() {
		throw new Error('blocked');
	}
};

describe('composeSample', () => {
	it('gives each archetype its own voice', () => {
		const ids = ['pragmatic_senior', 'terse_minimalist', 'supportive_mentor', 'witty_hacker'];
		const voices = ids.map((id) => composeSample(id, []).join('\n'));
		expect(new Set(voices).size).toBe(4);
	});

	it('shortens the reply when concise is on', () => {
		const long = composeSample('supportive_mentor', []);
		const short = composeSample('supportive_mentor', ['concise']);
		expect(short.join(' ').length).toBeLessThan(long.join(' ').length);
	});

	it('adds exactly one line per extra trait', () => {
		const base = composeSample('pragmatic_senior', ['concise']);
		const more = composeSample('pragmatic_senior', ['concise', 'explain_reasoning', 'strict']);
		expect(more.length).toBe(base.length + 2);
	});

	it('summarises the chosen profile in plain words', () => {
		expect(describeProfile('witty_hacker', ['concise', 'witty'])).toBe(
			'Witty Hacker, concise, light sarcasm'
		);
	});

	it('accepts labels as well as ids and falls back on unknown archetypes', () => {
		expect(composeSample('Witty Hacker', ['Light Sarcasm'])).toEqual(
			composeSample('witty_hacker', ['witty'])
		);
		expect(composeSample('nope', [])).toEqual(composeSample('pragmatic_senior', []));
	});
});

describe('profile helpers', () => {
	it('toggles traits by id and by legacy label', () => {
		expect(toggleTrait(['concise'], 'witty')).toEqual(['concise', 'witty']);
		expect(toggleTrait(['Concise', 'witty'], 'concise')).toEqual(['witty']);
		expect(hasTrait(['Light Sarcasm'], 'witty')).toBe(true);
	});

	it('never picks the current name and survives an extreme random', () => {
		expect(randomName('Ada', () => 0)).not.toBe('Ada');
		expect(randomName('', () => 0.999999)).toBeTruthy();
	});

	it('builds the payload with a default name and its blob', () => {
		const payload = profileFromDraft({ ...initialDraft(), name: '   ' });
		expect(payload.name).toBe('Bulle');
		expect(payload.avatar_url).toBe(blobUri('Bulle'));
		expect(payload.traits).toEqual(['concise']);
	});

	it('starts from an existing profile and tints every archetype', () => {
		const draft = initialDraft({ name: ' Ada ', archetype: 'Witty Hacker', traits: ['Concise'] });
		expect(draft).toEqual({ step: 0, name: 'Ada', archetype: 'witty_hacker', traits: ['concise'] });
		expect(archetypeTint('unknown')).toEqual(archetypeTint('pragmatic_senior'));
	});
});

describe('step machine', () => {
	it('stays inside the bounds', () => {
		expect(retreat(0)).toBe(0);
		expect(advance(STEPS.length - 1)).toBe(STEPS.length - 1);
		expect(clampStep(-4)).toBe(0);
		expect(clampStep(99)).toBe(STEPS.length - 1);
		expect(clampStep(Number.NaN)).toBe(0);
		expect(isLastStep(STEPS.length - 1)).toBe(true);
		expect(isLastStep(0)).toBe(false);
	});

	it('skips the name and the traits with defaults, and nothing else', () => {
		const base = { ...initialDraft(), name: 'Ada', traits: ['witty'] };
		expect(applySkip({ ...base, step: 1 })).toMatchObject({ step: 2, name: '', traits: ['witty'] });
		expect(applySkip({ ...base, step: 3 })).toMatchObject({
			step: 4,
			name: 'Ada',
			traits: ['concise']
		});
		const fixed = { ...base, step: 2 };
		expect(applySkip(fixed)).toBe(fixed);
		const apps = { ...base, step: 4 };
		expect(applySkip(apps)).toMatchObject({ step: 5, name: apps.name, traits: apps.traits });
	});

	it('labels buttons by what happens next', () => {
		expect(primaryLabel(0, '')).toBe('Give it a name');
		expect(primaryLabel(3, 'Ada')).toBe('Connect your apps');
		expect(primaryLabel(4, 'Ada')).toBe('Meet Ada');
		expect(primaryLabel(4, '  ')).toBe('Meet Bulle');
		expect(primaryLabel(5, 'Ada')).toBe('Say hello');
		expect(progressText(1)).toBe('Step 2 of 6: Name');
	});

	it('gates users without an assistant name', () => {
		expect(needsOnboarding(null)).toBe(true);
		expect(needsOnboarding({ assistant: null })).toBe(true);
		expect(needsOnboarding({ assistant: { name: '  ' } })).toBe(true);
		expect(needsOnboarding({ assistant: { name: 'Ada' } })).toBe(false);
	});
});

describe('draft storage', () => {
	it('round-trips a draft', () => {
		const store = memoryStore();
		const draft = { step: 3, name: 'Ada', archetype: 'witty_hacker', traits: ['witty'] };
		saveDraft(draft, store);
		expect(loadDraft(store)).toEqual(draft);
		clearDraft(store);
		expect(loadDraft(store)).toBeNull();
	});

	it('rejects corrupt and out-of-range data', () => {
		const store = memoryStore();
		store.setItem(DRAFT_KEY, '{nope');
		expect(loadDraft(store)).toBeNull();
		store.setItem(DRAFT_KEY, JSON.stringify({ name: 4, traits: [] }));
		expect(loadDraft(store)).toBeNull();
		store.setItem(
			DRAFT_KEY,
			JSON.stringify({ step: 99, name: 'x'.repeat(100), archetype: 'zzz', traits: ['concise', 7] })
		);
		const loaded = loadDraft(store);
		expect(loaded?.step).toBe(STEPS.length - 1);
		expect(loaded?.name).toHaveLength(40);
		expect(loaded?.archetype).toBe('pragmatic_senior');
		expect(loaded?.traits).toEqual(['concise']);
	});

	it('never throws when storage is blocked or missing', () => {
		const draft = initialDraft();
		expect(() => saveDraft(draft, brokenStore)).not.toThrow();
		expect(() => clearDraft(brokenStore)).not.toThrow();
		expect(loadDraft(brokenStore)).toBeNull();
		expect(loadDraft(null)).toBeNull();
		expect(() => saveDraft(draft, null)).not.toThrow();
	});
});
