import { registerIcons } from '@facile/muse';
import { _parts } from 'blobatar/internal';

export const LOGO_ICON = 'solar:smile-square-bold-duotone';

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
	dice: 'bulle:dice'
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
		body: '<g fill="currentColor"><path d="M3.46447 20.5355C4.92893 22 7.28595 22 12 22C16.714 22 19.0711 22 20.5355 20.5355C22 19.0711 22 16.714 22 12C22 7.28595 22 4.92893 20.5355 3.46447C19.0711 2 16.714 2 12 2C7.28595 2 4.92893 2 3.46447 3.46447C2 4.92893 2 7.28595 2 12C2 16.714 2 19.0711 3.46447 20.5355Z" opacity=".5"/><path d="M9.44661 15.3975C9.11385 15.1508 8.64413 15.2206 8.39747 15.5534C8.15082 15.8862 8.22062 16.3559 8.55339 16.6025C9.5258 17.3233 10.715 17.75 12 17.75C13.285 17.75 14.4742 17.3233 15.4466 16.6025C15.7794 16.3559 15.8492 15.8862 15.6025 15.5534C15.3559 15.2206 14.8862 15.1508 14.5534 15.3975C13.825 15.9373 12.9459 16.25 12 16.25C11.0541 16.25 10.175 15.9373 9.44661 15.3975Z"/><path d="M16 10.5C16 11.3284 15.5523 12 15 12C14.4477 12 14 11.3284 14 10.5C14 9.67157 14.4477 9 15 9C15.5523 9 16 9.67157 16 10.5Z"/><path d="M10 10.5C10 11.3284 9.55228 12 9 12C8.44772 12 8 11.3284 8 10.5C8 9.67157 8.44772 9 9 9C9.55228 9 10 9.67157 10 10.5Z"/></g>'
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
	const parts = _parts(safe);
	registerIcons({
		[key]: {
			width: 100 - 2 * AVATAR_CROP,
			height: 100 - 2 * AVATAR_CROP,
			body: `<g transform="translate(-${AVATAR_CROP} -${AVATAR_CROP})">${parts?.inner || ''}</g>`
		}
	});
	return key;
}
