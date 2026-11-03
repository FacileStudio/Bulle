import { listJobs, type Job } from './api';
import { sessions } from './chat/sessions.svelte';
import { spaces } from './spaces.svelte';

const BUSY_MS = 5000;
const IDLE_MS = 30000;

export const jobsActivity = $state({ jobs: [] as Job[], loaded: false });

/** Says whether any job has a run in progress. */
export function anyRunning(jobs: readonly Job[]): boolean {
	return jobs.some((job) => job.running);
}

/** Lists the conversations the jobs' latest runs write to, so a new run shows up as a change. */
export function runSessions(jobs: readonly Job[]): string {
	return jobs
		.map((job) => job.last_session_id ?? '')
		.filter(Boolean)
		.sort()
		.join(',');
}

/**
 * Reloads the job list, keeping the last good list when the request fails. A job that
 * started a new run also refreshes the chat list, so its conversation appears at once.
 */
export async function refreshJobs(): Promise<Job[]> {
	const before = runSessions(jobsActivity.jobs);
	try {
		jobsActivity.jobs = await listJobs(spaces.currentId);
	} catch {
		return jobsActivity.jobs;
	}
	if (jobsActivity.loaded && runSessions(jobsActivity.jobs) !== before) void sessions.refresh();
	jobsActivity.loaded = true;
	return jobsActivity.jobs;
}

/**
 * Polls the job list while the tab is visible: fast while a run is going, slow otherwise.
 * Returns the cleanup that stops polling.
 */
export function watchJobs(): () => void {
	let timer: ReturnType<typeof setTimeout> | undefined;
	let stopped = false;
	const tick = async () => {
		if (document.visibilityState === 'visible') await refreshJobs();
		if (!stopped) timer = setTimeout(tick, anyRunning(jobsActivity.jobs) ? BUSY_MS : IDLE_MS);
	};
	const onvisible = () => {
		if (document.visibilityState !== 'visible') return;
		clearTimeout(timer);
		void tick();
	};
	void tick();
	document.addEventListener('visibilitychange', onvisible);
	return () => {
		stopped = true;
		clearTimeout(timer);
		document.removeEventListener('visibilitychange', onvisible);
	};
}
