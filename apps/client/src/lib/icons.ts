import { registerIcons } from '@facile/muse';
import { parts as blobParts } from '@facile/blob';

export const LOGO_ICON = 'solar:smile-circle-bold-duotone';

const stroke = (body: string) =>
	`<g fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">${body}</g>`;

export const CHAT_ICONS = {
	send: 'bulle:send',
	stop: 'bulle:stop',
	menu: 'bulle:menu',
	chat: 'bulle:chat',
	terminal: 'bulle:terminal',
	phone: 'bulle:phone',
	tool: 'bulle:tool',
	thinking: 'bulle:thinking',
	chevronRight: 'bulle:chevron-right',
	dice: 'bulle:dice',
	file: 'bulle:file'
} as const;

registerIcons({
	[CHAT_ICONS.send]: {
		width: 24,
		height: 24,
		body: stroke('<path d="M12 19V5"/><path d="m5 12 7-7 7 7"/>')
	},
	[CHAT_ICONS.stop]: {
		width: 24,
		height: 24,
		body: '<rect x="6" y="6" width="12" height="12" rx="2.5" fill="currentColor"/>'
	},
	[CHAT_ICONS.menu]: {
		width: 24,
		height: 24,
		body: stroke('<path d="M4 7h16"/><path d="M4 12h16"/><path d="M4 17h16"/>')
	},
	[CHAT_ICONS.chat]: {
		width: 24,
		height: 24,
		body: stroke('<path d="M21 12a8 8 0 0 1-11.6 7.1L4 20l1-4.6A8 8 0 1 1 21 12Z"/>')
	},
	[CHAT_ICONS.terminal]: {
		width: 24,
		height: 24,
		body: stroke('<path d="m5 8 5 4-5 4"/><path d="M13 17h6"/>')
	},
	[CHAT_ICONS.phone]: {
		width: 24,
		height: 24,
		body: stroke('<rect x="6.5" y="2.5" width="11" height="19" rx="2.5"/><path d="M11 18h2"/>')
	},
	[CHAT_ICONS.tool]: {
		width: 24,
		height: 24,
		body: stroke(
			'<path d="M14.7 6.3a4 4 0 0 0-5 5L4 17l3 3 5.7-5.7a4 4 0 0 0 5-5l-2.4 2.4-2.6-.6-.6-2.6Z"/>'
		)
	},
	[CHAT_ICONS.thinking]: {
		width: 24,
		height: 24,
		body: stroke(
			'<circle cx="6" cy="12" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="18" cy="12" r="1"/>'
		)
	},
	[CHAT_ICONS.chevronRight]: {
		width: 24,
		height: 24,
		body: stroke('<path d="m9 6 6 6-6 6"/>')
	},
	[CHAT_ICONS.file]: {
		width: 24,
		height: 24,
		body: stroke('<path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8Z"/><path d="M14 3v5h5"/>')
	},
	[CHAT_ICONS.dice]: {
		width: 24,
		height: 24,
		body: stroke(
			'<rect width="16" height="16" x="4" y="4" rx="2"/><circle cx="8.5" cy="8.5" r="1" fill="currentColor"/><circle cx="15.5" cy="8.5" r="1" fill="currentColor"/><circle cx="12" cy="12" r="1" fill="currentColor"/><circle cx="8.5" cy="15.5" r="1" fill="currentColor"/><circle cx="15.5" cy="15.5" r="1" fill="currentColor"/>'
		)
	}
});

registerIcons({
	[LOGO_ICON]: {
		width: 24,
		height: 24,
		body: '<g fill="currentColor"><path d="M12 22c5.523 0 10-4.477 10-10S17.523 2 12 2S2 6.477 2 12s4.477 10 10 10" opacity=".5"/><path d="M8.397 15.553a.75.75 0 0 1 1.05-.155c.728.54 1.607.852 2.553.852s1.825-.313 2.553-.852a.75.75 0 1 1 .894 1.204A5.77 5.77 0 0 1 12 17.75a5.77 5.77 0 0 1-3.447-1.148a.75.75 0 0 1-.156-1.049M15 12c.552 0 1-.672 1-1.5S15.552 9 15 9s-1 .672-1 1.5s.448 1.5 1 1.5m-6 0c.552 0 1-.672 1-1.5S9.552 9 9 9s-1 .672-1 1.5s.448 1.5 1 1.5"/></g>'
	}
});

export const PLATFORM_ICONS = {
	apple: 'bulle:os-apple',
	windows: 'bulle:os-windows',
	linux: 'bulle:os-linux',
	android: 'bulle:os-android',
} as const;

registerIcons({
	[PLATFORM_ICONS.apple]: { width: 24, height: 24, body: '<path fill="currentColor" d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47c-1.34.03-1.77-.79-3.29-.79c-1.53 0-2 .77-3.27.82c-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51c1.28-.02 2.5.87 3.29.87c.78 0 2.26-1.07 3.81-.91c.65.03 2.47.26 3.64 1.98c-.09.06-2.17 1.28-2.15 3.81c.03 3.02 2.65 4.03 2.68 4.04c-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5c.13 1.17-.34 2.35-1.04 3.19c-.69.85-1.83 1.51-2.95 1.42c-.15-1.15.41-2.35 1.05-3.11"/>' },
	[PLATFORM_ICONS.windows]: { width: 24, height: 24, body: '<path fill="currentColor" d="M3 12V6.75l6-1.32v6.48zm17-9v8.75l-10 .15V5.21zM3 13l6 .09v6.81l-6-1.15zm17 .25V22l-10-1.91V13.1z"/>' },
	[PLATFORM_ICONS.linux]: { width: 24, height: 24, body: '<path fill="currentColor" d="M14.62 8.35c-.42.28-1.75 1.04-1.95 1.19c-.39.31-.75.29-1.14-.01c-.2-.16-1.53-.92-1.95-1.19c-.48-.31-.45-.7.08-.92c1.64-.69 3.28-.64 4.91.03c.49.21.51.6.05.9m7.22 7.28c-.93-2.09-2.2-3.99-3.84-5.66a4.3 4.3 0 0 1-1.06-1.88c-.1-.33-.17-.67-.24-1.01c-.2-.88-.29-1.78-.7-2.61c-.73-1.58-2-2.4-3.84-2.47c-1.81.05-3.16.81-3.95 2.4c-.21.43-.36.88-.46 1.34c-.17.76-.32 1.55-.5 2.32c-.15.65-.45 1.21-.96 1.71c-1.61 1.57-2.9 3.37-3.88 5.35c-.14.29-.28.58-.37.88c-.19.66.29 1.12.99.96c.44-.09.88-.18 1.3-.31c.41-.15.57-.05.67.35c.65 2.15 2.07 3.66 4.24 4.5c4.12 1.56 8.93-.66 9.97-4.58c.07-.27.17-.37.47-.27c.46.14.93.24 1.4.35c.49.09.85-.16.92-.64c.03-.26-.06-.49-.16-.73"/>' },
	[PLATFORM_ICONS.android]: { width: 24, height: 24, body: '<path fill="currentColor" d="M16.61 15.15c-.46 0-.84-.37-.84-.83s.38-.82.84-.82s.84.36.84.82s-.38.83-.84.83m-9.2 0c-.46 0-.84-.37-.84-.83s.38-.82.84-.82s.83.36.83.82s-.37.83-.83.83m9.5-5.01l1.67-2.88c.09-.17.03-.38-.13-.47c-.17-.1-.38-.04-.45.13l-1.71 2.91A10.15 10.15 0 0 0 12 8.91c-1.53 0-3 .33-4.27.91L6.04 6.91a.334.334 0 0 0-.47-.13c-.17.09-.22.3-.13.47l1.66 2.88C4.25 11.69 2.29 14.58 2 18h20c-.28-3.41-2.23-6.3-5.09-7.86"/>' }
});

const AVATAR_CROP = 10;

export function registerAgentAvatarIcon(name: string): string {
	const safe = name.trim().toLowerCase() || 'bulle';
	const key = `bulle:avatar:${safe}`;
	const parts = blobParts(safe);
	registerIcons({
		[key]: {
			width: 100 - 2 * AVATAR_CROP,
			height: 100 - 2 * AVATAR_CROP,
			body: `<g transform="translate(-${AVATAR_CROP} -${AVATAR_CROP})">${parts?.inner || ''}</g>`
		}
	});
	return key;
}
