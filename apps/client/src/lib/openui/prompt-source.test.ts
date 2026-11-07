import { describe, expect, it } from 'vitest';
import { createInspector } from './inspect';
import { library } from './library';
import { EXAMPLES } from './prompt-source';

describe('openui prompt examples', () => {
	const inspect = createInspector(library);
	it.each(EXAMPLES.map((example, i) => [i, example]))('example %i renders', (_, example) => {
		const block = /```openui\n([\s\S]*?)\n```/.exec(example as string);
		expect(block).not.toBeNull();
		expect(inspect(block![1], false)).toEqual({ status: 'ready' });
	});
});
