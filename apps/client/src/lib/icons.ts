import { registerIcons } from '@facile/muse';

export const LOGO_ICON = 'solar:smile-square-bold-duotone';

const stroke = (body: string) =>
	`<g fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round">${body}</g>`;

export const CHAT_ICONS = {
	send: 'bulle:send',
	stop: 'bulle:stop',
	menu: 'bulle:menu',
	chat: 'bulle:chat',
	terminal: 'bulle:terminal',
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
