import { listJobs, type Job } from './api';

const BUSY_MS = 5000;
const IDLE_MS = 30000;

export const jobsActivity = $state({ jobs: [] as Job[], loaded: false });

/** Says whether any job has a run in progress. */
export function anyRunning(jobs: readonly Job[]): boolean {
	return jobs.some((job) => job.running);
}

/** Reloads the job list, keeping the last good list when the request fails. */
export async function refreshJobs(): Promise<Job[]> {
	try {
		jobsActivity.jobs = await listJobs();
		jobsActivity.loaded = true;
	} catch {
		return jobsActivity.jobs;
	}
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
