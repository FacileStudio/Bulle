export type SseFrame = { id: string; event: string; data: string };

export function createSseParser(onFrame: (frame: SseFrame) => void): (chunk: string) => void {
	let buffer = '';
	let id = '';
	let event = '';
	let data: string[] = [];

	function handleLine(line: string) {
		if (line === '') {
			if (data.length > 0) onFrame({ id, event, data: data.join('\n') });
			event = '';
			data = [];
			return;
		}
		if (line.startsWith(':')) return;
		const colon = line.indexOf(':');
		const field = colon < 0 ? line : line.slice(0, colon);
		let value = colon < 0 ? '' : line.slice(colon + 1);
		if (value.startsWith(' ')) value = value.slice(1);
		if (field === 'id') id = value;
		else if (field === 'event') event = value;
		else if (field === 'data') data.push(value);
	}

	return (chunk) => {
		buffer += chunk;
		for (;;) {
			const end = buffer.search(/[\r\n]/);
			if (end < 0) return;
			if (buffer[end] === '\r' && end === buffer.length - 1) return;
			const width = buffer[end] === '\r' && buffer[end + 1] === '\n' ? 2 : 1;
			const line = buffer.slice(0, end);
			buffer = buffer.slice(end + width);
			handleLine(line);
		}
	};
}
