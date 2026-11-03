import { describe, expect, it } from 'vitest';
import {
	MAX_ATTACHMENT_BYTES,
	MAX_ATTACHMENTS,
	admit,
	formatSize,
	mediaTypeFor,
	messageBody,
	validateFile
} from './attachments';

const file = (name: string, type: string, size = 10) => ({ name, type, size });

describe('validateFile', () => {
	it('accepts images, pdfs and text-like files', () => {
		expect(validateFile(file('a.png', 'image/png'))).toBeNull();
		expect(validateFile(file('a.webp', 'image/webp'))).toBeNull();
		expect(validateFile(file('a.pdf', 'application/pdf'))).toBeNull();
		expect(validateFile(file('notes.md', ''))).toBeNull();
		expect(validateFile(file('data.json', 'application/json'))).toBeNull();
		expect(validateFile(file('main.rs', ''))).toBeNull();
		expect(validateFile(file('Dockerfile', ''))).toBeNull();
	});

	it('rejects unsupported types and files over 20 MB', () => {
		expect(validateFile(file('a.zip', 'application/zip'))).toBe('unsupported');
		expect(validateFile(file('a.mp4', 'video/mp4'))).toBe('unsupported');
		expect(validateFile(file('a.svg', 'image/svg+xml'))).toBe('unsupported');
		expect(validateFile(file('a.png', 'image/png', MAX_ATTACHMENT_BYTES))).toBeNull();
		expect(validateFile(file('a.png', 'image/png', MAX_ATTACHMENT_BYTES + 1))).toBe('too-large');
	});
});

describe('mediaTypeFor', () => {
	it('sends source files as text/plain when the browser guesses wrong', () => {
		expect(mediaTypeFor(file('app.ts', 'video/mp2t'))).toBe('text/plain');
		expect(mediaTypeFor(file('main.go', ''))).toBe('text/plain');
		expect(mediaTypeFor(file('a.csv', 'text/csv'))).toBe('text/csv');
		expect(mediaTypeFor(file('a.jpg', 'image/jpeg'))).toBe('image/jpeg');
	});
});

describe('admit', () => {
	it('caps a message at ten files and skips invalid ones when counting', () => {
		const picked = [
			file('bad.zip', 'application/zip'),
			...Array.from({ length: 4 }, (_, i) => file(`${i}.png`, 'image/png'))
		];
		const result = admit(picked, MAX_ATTACHMENTS - 3);
		expect(result.map((r) => r.error)).toEqual(['unsupported', null, null, null, 'too-many']);
	});

	it('refuses everything when the message is full', () => {
		expect(admit([file('a.png', 'image/png')], MAX_ATTACHMENTS)[0].error).toBe('too-many');
	});
});

describe('messageBody', () => {
	it('omits attachments when there are none', () => {
		expect(messageBody('hi')).toEqual({ text: 'hi' });
		expect(messageBody('hi', [])).toEqual({ text: 'hi' });
	});

	it('sends attachment ids alongside possibly empty text', () => {
		expect(messageBody('', ['a1', 'a2'])).toEqual({ text: '', attachments: ['a1', 'a2'] });
	});
});

describe('formatSize', () => {
	it('picks a readable unit', () => {
		expect(formatSize(512)).toBe('512 B');
		expect(formatSize(2048)).toBe('2 KB');
		expect(formatSize(5 * 1024 * 1024)).toBe('5.0 MB');
	});
});
