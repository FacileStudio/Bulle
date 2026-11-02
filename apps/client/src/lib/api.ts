import { apiFetch } from '$lib/backend';

export * from './api/index';

export type Job = {
	id: string;
	user_id?: number;
	name: string;
	schedule: string;
	prompt: string;
	profile?: string;
	delivery?: string;
	enabled: boolean;
	created_at: string;
	updated_at: string;
};

export type JobRun = {
	id: string;
	job_id: string;
	run_id: string;
	status: string;
	started_at: string;
	finished_at?: string | null;
};

export type CreateJobInput = {
	name: string;
	schedule: string;
	prompt: string;
	profile?: string;
	delivery?: string;
	enabled?: boolean;
};

export type UpdateJobInput = {
	name: string;
	schedule: string;
	prompt: string;
	profile?: string;
	delivery?: string;
	enabled?: boolean;
};

export type RunJobResult = {
	run_id: string;
	job_run: JobRun;
};

export function listJobs(): Promise<Job[]> {
	return apiFetch<Job[]>('/jobs');
}

export function createJob(input: CreateJobInput): Promise<Job> {
	return apiFetch<Job>('/jobs', {
		method: 'POST',
		body: JSON.stringify(input)
	});
}

export function getJob(id: string): Promise<Job> {
	return apiFetch<Job>(`/jobs/${encodeURIComponent(id)}`);
}

export function updateJob(id: string, input: UpdateJobInput): Promise<Job> {
	return apiFetch<Job>(`/jobs/${encodeURIComponent(id)}`, {
		method: 'PUT',
		body: JSON.stringify(input)
	});
}

export function deleteJob(id: string): Promise<void> {
	return apiFetch<void>(`/jobs/${encodeURIComponent(id)}`, {
		method: 'DELETE'
	});
}

export function runJob(id: string): Promise<RunJobResult> {
	return apiFetch<RunJobResult>(`/jobs/${encodeURIComponent(id)}/run`, {
		method: 'POST'
	});
}

export function listJobRuns(id: string): Promise<JobRun[]> {
	return apiFetch<JobRun[]>(`/jobs/${encodeURIComponent(id)}/runs`);
}

export const jobsApi = {
	listJobs,
	createJob,
	getJob,
	updateJob,
	deleteJob,
	runJob,
	listJobRuns
};
