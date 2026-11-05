import { describe, expect, it } from 'vitest';
import { createFrameWriter, type Characteristic } from './gatt';

function characteristic(failures: Error[]): { target: Characteristic; sizes: number[] } {
	const sizes: number[] = [];
	const write = async (value: Uint8Array) => {
		const failure = failures.shift();
		if (failure) throw failure;
		sizes.push(value.length);
	};
	const target = {
		readValue: async () => new DataView(new ArrayBuffer(0)),
		writeValueWithResponse: write,
		writeValueWithoutResponse: write,
		startNotifications: async () => target,
		addEventListener: () => undefined
	} as Characteristic;
	return { target, sizes };
}

function named(name: string, message: string): Error {
	const error = new Error(message);
	error.name = name;
	return error;
}

describe('bohs.v1 GATT frame writer', () => {
	it('retries a transient failure at full size', async () => {
		const { target, sizes } = characteristic([named('NetworkError', 'GATT operation already in progress.')]);
		const write = createFrameWriter();
		await write(target, new Uint8Array(400), false);
		await write(target, new Uint8Array(400), false);
		expect(sizes).toEqual([244, 158, 244, 158]);
	});

	it('drops to 20-byte chunks only when the link refuses the length', async () => {
		const { target, sizes } = characteristic([named('InvalidModificationError', 'Value is too long.')]);
		const write = createFrameWriter();
		await write(target, new Uint8Array(38), true);
		expect(sizes).toEqual([20, 20]);
	});
});
