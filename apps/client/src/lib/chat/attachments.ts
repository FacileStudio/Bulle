export type Attachment = {
	id: string;
	name: string;
	media_type: string;
	size: number;
};

export type AttachmentError = 'too-large' | 'unsupported' | 'too-many' | 'failed';

export const MAX_ATTACHMENT_BYTES = 20 * 1024 * 1024;
export const MAX_ATTACHMENTS = 10;

const IMAGE_TYPES = ['image/png', 'image/jpeg', 'image/gif', 'image/webp'];
const DOCUMENT_TYPES = ['application/pdf'];
const TEXT_TYPES = [
	'application/json',
	'application/xml',
	'application/yaml',
	'application/x-yaml',
	'application/toml',
	'application/javascript',
	'application/typescript',
	'application/x-sh',
	'application/sql'
];

const TEXT_EXTENSIONS = [
	'txt', 'md', 'markdown', 'csv', 'tsv', 'json', 'jsonl', 'xml', 'yaml', 'yml', 'toml', 'ini',
	'cfg', 'conf', 'log', 'html', 'htm', 'css', 'scss', 'sass', 'less', 'js', 'mjs', 'cjs', 'ts',
	'mts', 'cts', 'tsx', 'jsx', 'svelte', 'vue', 'astro', 'py', 'rb', 'go', 'rs', 'java', 'kt',
	'kts', 'swift', 'c', 'h', 'cpp', 'hpp', 'cc', 'cs', 'php', 'sh', 'bash', 'zsh', 'fish', 'sql',
	'lua', 'r', 'dart', 'scala', 'ex', 'exs', 'erl', 'hs', 'ml', 'clj', 'el', 'vim', 'zig', 'nix',
	'tf', 'proto', 'graphql', 'gql', 'gradle', 'diff', 'patch', 'tex', 'rst', 'org'
];

const TEXT_NAMES = ['dockerfile', 'makefile', 'justfile', 'license', 'readme'];

export const ATTACHMENT_ACCEPT = [
	...IMAGE_TYPES,
	...DOCUMENT_TYPES,
	'text/*',
	...TEXT_TYPES,
	...TEXT_EXTENSIONS.map((ext) => `.${ext}`)
].join(',');

type FileLike = { name: string; type: string; size: number };

function extension(name: string): string {
	const dot = name.lastIndexOf('.');
	return dot > 0 ? name.slice(dot + 1).toLowerCase() : '';
}

function isTextFile(file: FileLike): boolean {
	const type = file.type.toLowerCase();
	if (type.startsWith('text/') || TEXT_TYPES.includes(type)) return true;
	const lower = file.name.toLowerCase();
	return TEXT_EXTENSIONS.includes(extension(lower)) || TEXT_NAMES.includes(lower);
}

export function isImageType(mediaType: string): boolean {
	return IMAGE_TYPES.includes(mediaType.toLowerCase());
}

export function mediaTypeFor(file: FileLike): string | null {
	const type = file.type.toLowerCase();
	if (IMAGE_TYPES.includes(type) || DOCUMENT_TYPES.includes(type)) return type;
	if (!isTextFile(file)) return null;
	return type.startsWith('text/') || TEXT_TYPES.includes(type) ? type : 'text/plain';
}

export function validateFile(file: FileLike): AttachmentError | null {
	if (mediaTypeFor(file) === null) return 'unsupported';
	if (file.size > MAX_ATTACHMENT_BYTES) return 'too-large';
	return null;
}

export function admit<T extends FileLike>(
	files: T[],
	taken: number
): { file: T; error: AttachmentError | null }[] {
	let room = Math.max(0, MAX_ATTACHMENTS - taken);
	return files.map((file) => {
		const error = validateFile(file);
		if (error) return { file, error };
		if (room === 0) return { file, error: 'too-many' };
		room -= 1;
		return { file, error: null };
	});
}

export function errorText(error: AttachmentError): string {
	switch (error) {
		case 'too-large':
			return 'Too large, 20 MB max';
		case 'unsupported':
			return 'Unsupported file type';
		case 'too-many':
			return `${MAX_ATTACHMENTS} files max per message`;
		case 'failed':
			return 'Upload failed';
	}
}

export function formatSize(bytes: number): string {
	if (!Number.isFinite(bytes) || bytes < 0) return '';
	if (bytes < 1024) return `${bytes} B`;
	if (bytes < 1024 * 1024) return `${Math.round(bytes / 1024)} KB`;
	return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function messageBody(text: string, attachmentIds: string[] = []): {
	text: string;
	attachments?: string[];
} {
	return attachmentIds.length ? { text, attachments: attachmentIds } : { text };
}
