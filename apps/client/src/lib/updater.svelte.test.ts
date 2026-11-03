import { beforeEach, describe, expect, it, vi } from 'vitest';
import pkg from '../../package.json';
import { UpdaterStore } from './updater.svelte';

vi.mock('@tauri-apps/api/core', () => ({
	isTauri: vi.fn(() => false)
}));

vi.mock('@tauri-apps/plugin-updater', () => ({
	check: vi.fn()
}));

vi.mock('@tauri-apps/plugin-process', () => ({
	relaunch: vi.fn()
}));

import { isTauri } from '@tauri-apps/api/core';
import { check } from '@tauri-apps/plugin-updater';
import { relaunch } from '@tauri-apps/plugin-process';

describe('UpdaterStore outside Tauri', () => {
	beforeEach(() => {
		vi.resetAllMocks();
		vi.mocked(isTauri).mockReturnValue(false);
	});

	it('initializes with idle state and current version', () => {
		const store = new UpdaterStore();
		expect(store.state).toBe('idle');
		expect(store.currentVersion).toBe(pkg.version);
		expect(store.availableVersion).toBeNull();
		expect(store.downloadProgress).toBe(0);
		expect(store.errorMessage).toBeNull();
	});

	it('checkNow marks upToDate on manual check', async () => {
		const store = new UpdaterStore();
		await store.checkNow(false);
		expect(store.state).toBe('upToDate');
	});

	it('checkNow stays idle on background check', async () => {
		const store = new UpdaterStore();
		await store.checkNow(true);
		expect(store.state).toBe('idle');
	});

	it('installNow transitions to ready with full progress', async () => {
		const store = new UpdaterStore();
		await store.installNow();
		expect(store.state).toBe('ready');
		expect(store.downloadProgress).toBe(100);
	});

	it('restartNow completes without throwing', async () => {
		const store = new UpdaterStore();
		await expect(store.restartNow()).resolves.toBeUndefined();
	});
});

describe('UpdaterStore inside Tauri', () => {
	beforeEach(() => {
		vi.resetAllMocks();
		vi.mocked(isTauri).mockReturnValue(true);
	});

	it('detects available update and records version', async () => {
		const store = new UpdaterStore();
		vi.mocked(check).mockResolvedValueOnce({
			version: '3.5.0',
			currentVersion: '3.4.0',
			downloadAndInstall: vi.fn()
		} as any);

		await store.checkNow(false);

		expect(store.state).toBe('available');
		expect(store.availableVersion).toBe('3.5.0');
		expect(store.currentVersion).toBe('3.4.0');
	});

	it('detects no update and transitions to upToDate', async () => {
		const store = new UpdaterStore();
		vi.mocked(check).mockResolvedValueOnce(null);

		await store.checkNow(false);

		expect(store.state).toBe('upToDate');
		expect(store.availableVersion).toBeNull();
	});

	it('captures errors on manual check', async () => {
		const store = new UpdaterStore();
		vi.mocked(check).mockRejectedValueOnce(new Error('Network offline'));

		await store.checkNow(false);

		expect(store.state).toBe('error');
		expect(store.errorMessage).toBe('Network offline');
	});

	it('stays idle on background check failure', async () => {
		const store = new UpdaterStore();
		vi.mocked(check).mockRejectedValueOnce(new Error('Network offline'));

		await store.checkNow(true);

		expect(store.state).toBe('idle');
		expect(store.errorMessage).toBeNull();
	});

	it('errors when installNow is called without available update', async () => {
		const store = new UpdaterStore();
		await store.installNow();

		expect(store.state).toBe('error');
		expect(store.errorMessage).toBe('No update available to install');
	});

	it('downloads update and tracks progress through completion', async () => {
		const store = new UpdaterStore();
		const downloadAndInstall = vi.fn().mockImplementation(async (cb) => {
			cb({ event: 'Started', data: { contentLength: 200 } });
			cb({ event: 'Progress', data: { chunkLength: 100 } });
			cb({ event: 'Finished' });
		});

		vi.mocked(check).mockResolvedValueOnce({
			version: '3.5.0',
			currentVersion: '3.4.0',
			downloadAndInstall
		} as any);

		await store.checkNow(false);
		await store.installNow();

		expect(downloadAndInstall).toHaveBeenCalledTimes(1);
		expect(store.state).toBe('ready');
		expect(store.downloadProgress).toBe(100);
	});

	it('handles download failure gracefully', async () => {
		const store = new UpdaterStore();
		const downloadAndInstall = vi.fn().mockRejectedValueOnce(new Error('Disk full'));

		vi.mocked(check).mockResolvedValueOnce({
			version: '3.5.0',
			currentVersion: '3.4.0',
			downloadAndInstall
		} as any);

		await store.checkNow(false);
		await store.installNow();

		expect(store.state).toBe('error');
		expect(store.errorMessage).toBe('Disk full');
	});

	it('calls relaunch on restartNow', async () => {
		const store = new UpdaterStore();
		vi.mocked(relaunch).mockResolvedValueOnce(undefined as never);

		await store.restartNow();

		expect(relaunch).toHaveBeenCalledTimes(1);
	});

	it('handles relaunch error', async () => {
		const store = new UpdaterStore();
		vi.mocked(relaunch).mockRejectedValueOnce(new Error('Spawn failed'));

		await store.restartNow();

		expect(store.state).toBe('error');
		expect(store.errorMessage).toBe('Spawn failed');
	});
});
