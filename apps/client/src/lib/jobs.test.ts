import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import {
	createJob,
	deleteJob,
	getJob,
	listJobRuns,
	listJobs,
	runJob,
	updateJob
} from './api';

describe('jobs api client', () => {
	const originalFetch = globalThis.fetch;

	beforeEach(() => {
		globalThis.fetch = vi.fn();
	});

	afterEach(() => {
		globalThis.fetch = originalFetch;
	});

	it('listJobs calls GET /api/jobs', async () => {
		const mockJobs = [{ id: 'job-1', name: 'daily', schedule: '@daily', prompt: 'hi', enabled: true }];
		(globalThis.fetch as any).mockResolvedValueOnce({
			ok: true,
			status: 200,
			json: async () => mockJobs
		});

		const res = await listJobs();
		expect(res).toEqual(mockJobs);
		expect(globalThis.fetch).toHaveBeenCalledWith('/api/jobs', expect.objectContaining({ credentials: 'same-origin' }));
	});

	it('createJob sends POST to /api/jobs', async () => {
		const input = { name: 'weekly', schedule: '0 0 * * 1', prompt: 'report', enabled: true };
		const created = { id: 'job-2', ...input };
		(globalThis.fetch as any).mockResolvedValueOnce({
			ok: true,
			status: 201,
			json: async () => created
		});

		const res = await createJob(input);
		expect(res).toEqual(created);
		expect(globalThis.fetch).toHaveBeenCalledWith(
			'/api/jobs',
			expect.objectContaining({
				method: 'POST',
				body: JSON.stringify(input)
			})
		);
	});

	it('getJob calls GET /api/jobs/{id}', async () => {
		const mockJob = { id: 'job-1', name: 'daily' };
		(globalThis.fetch as any).mockResolvedValueOnce({
			ok: true,
			status: 200,
			json: async () => mockJob
		});

		const res = await getJob('job-1');
		expect(res).toEqual(mockJob);
		expect(globalThis.fetch).toHaveBeenCalledWith('/api/jobs/job-1', expect.anything());
	});

	it('updateJob calls PUT /api/jobs/{id}', async () => {
		const input = { name: 'daily-updated', schedule: '@hourly', prompt: 'check', enabled: false };
		const updated = { id: 'job-1', ...input };
		(globalThis.fetch as any).mockResolvedValueOnce({
			ok: true,
			status: 200,
			json: async () => updated
		});

		const res = await updateJob('job-1', input);
		expect(res).toEqual(updated);
		expect(globalThis.fetch).toHaveBeenCalledWith(
			'/api/jobs/job-1',
			expect.objectContaining({
				method: 'PUT',
				body: JSON.stringify(input)
			})
		);
	});

	it('deleteJob calls DELETE /api/jobs/{id}', async () => {
		(globalThis.fetch as any).mockResolvedValueOnce({
			ok: true,
			status: 204
		});

		await deleteJob('job-1');
		expect(globalThis.fetch).toHaveBeenCalledWith(
			'/api/jobs/job-1',
			expect.objectContaining({ method: 'DELETE' })
		);
	});

	it('runJob calls POST /api/jobs/{id}/run', async () => {
		const mockResult = { run_id: 'r-1', job_run: { id: 'jr-1', job_id: 'job-1', run_id: 'r-1', status: 'running' } };
		(globalThis.fetch as any).mockResolvedValueOnce({
			ok: true,
			status: 200,
			json: async () => mockResult
		});

		const res = await runJob('job-1');
		expect(res).toEqual(mockResult);
		expect(globalThis.fetch).toHaveBeenCalledWith(
			'/api/jobs/job-1/run',
			expect.objectContaining({ method: 'POST' })
		);
	});

	it('listJobRuns calls GET /api/jobs/{id}/runs', async () => {
		const mockRuns = [{ id: 'jr-1', job_id: 'job-1', status: 'finished' }];
		(globalThis.fetch as any).mockResolvedValueOnce({
			ok: true,
			status: 200,
			json: async () => mockRuns
		});

		const res = await listJobRuns('job-1');
		expect(res).toEqual(mockRuns);
		expect(globalThis.fetch).toHaveBeenCalledWith('/api/jobs/job-1/runs', expect.anything());
	});
});
