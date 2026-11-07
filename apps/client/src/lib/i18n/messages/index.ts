import type { Catalogs, Message } from '@facile/muse';
import type { Row } from '../row';
import { chat } from './chat';
import { ideas } from './ideas';
import { misc } from './misc';
import { settings } from './settings';
import { shell } from './shell';

const rows = {
	...shell,
	...chat,
	...settings,
	...ideas,
	...misc
};

export type AppKey = keyof typeof rows;
type AppCatalog = Record<AppKey, Message>;

function column(index: number): AppCatalog {
	const entries = Object.entries(rows as Record<string, Row>).map(([key, row]) => [key, row[index]]);
	return Object.fromEntries(entries) as AppCatalog;
}

export const catalogs: Catalogs<AppCatalog> = {
	en: column(0),
	fr: column(1),
	es: column(2),
	de: column(3)
};
