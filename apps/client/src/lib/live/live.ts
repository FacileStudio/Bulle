import { authedInit, backend } from '$lib/backend';
import { LiveStream } from './stream';

export const live = new LiveStream({
	connect: async (signal) =>
		fetch(
			`${backend.baseUrl}/push/stream`,
			await authedInit({ headers: { Accept: 'text/event-stream' }, signal })
		),
	wait: (ms) => new Promise((resolve) => setTimeout(resolve, ms))
});
