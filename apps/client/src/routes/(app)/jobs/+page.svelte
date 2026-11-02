<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Alert,
		Button,
		Card,
		EmptyState,
		Field,
		Icon,
		IconButton,
		Input,
		Modal,
		PageHeader,
		Section,
		Select,
		Spinner,
		Stack,
		StatusDot,
		Switch,
		Textarea,
		icons,
		toast
	} from '@facile/muse';
	import {
		type Job,
		type JobRun,
		listJobs,
		createJob,
		updateJob,
		deleteJob,
		runJob,
		listJobRuns
	} from '$lib/api';
	import { describeSchedule } from '$lib/schedule';
	import { CHAT_ICONS } from '$lib/icons';
	import { jobsActivity, refreshJobs } from '$lib/jobs-activity.svelte';

	const jobs = $derived(jobsActivity.jobs);
	let loading = $state(true);
	let error = $state<string | null>(null);

	let createModalOpen = $state(false);
	let createSubmitting = $state(false);
	let createError = $state<string | null>(null);
	let newName = $state('');
	let newSchedule = $state('');
	let newPrompt = $state('');
	let newProfile = $state('');
	let newDelivery = $state('');
	let newEnabled = $state(true);

	let editModalOpen = $state(false);
	let editingJob = $state<Job | null>(null);
	let editSubmitting = $state(false);
	let editError = $state<string | null>(null);
	let editName = $state('');
	let editSchedule = $state('');
	let editPrompt = $state('');
	let editProfile = $state('');
	let editDelivery = $state('');
	let editEnabled = $state(true);
	const editLegacyDelivery = $derived(editDelivery !== '' && editDelivery !== 'push');

	let historyModalOpen = $state(false);
	let historyJob = $state<Job | null>(null);
	let historyRuns = $state<JobRun[]>([]);
	let historyLoading = $state(false);

	let runningJobIds = $state<Record<string, boolean>>({});

	async function load() {
		loading = !jobsActivity.loaded;
		error = null;
		try {
			jobsActivity.jobs = await listJobs();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Failed to load background jobs';
		} finally {
			loading = false;
		}
	}

	onMount(() => {
		void load();
	});

	function openCreate() {
		newName = '';
		newSchedule = '0 9 * * *';
		newPrompt = '';
		newProfile = '';
		newDelivery = '';
		newEnabled = true;
		createError = null;
		createModalOpen = true;
	}

	async function handleCreate(e: SubmitEvent) {
		e.preventDefault();
		createSubmitting = true;
		createError = null;
		try {
			const created = await createJob({
				name: newName.trim(),
				schedule: newSchedule.trim(),
				prompt: newPrompt.trim(),
				profile: newProfile.trim() || undefined,
				delivery: newDelivery.trim() || undefined,
				enabled: newEnabled
			});
			jobsActivity.jobs = [created, ...jobs];
			createModalOpen = false;
			toast.success('Job created');
		} catch (err) {
			createError = err instanceof Error ? err.message : 'Failed to create job';
		} finally {
			createSubmitting = false;
		}
	}

	function openEdit(job: Job) {
		editingJob = job;
		editName = job.name;
		editSchedule = job.schedule;
		editPrompt = job.prompt;
		editProfile = job.profile || '';
		editDelivery = job.delivery || '';
		editEnabled = job.enabled;
		editError = null;
		editModalOpen = true;
	}

	async function handleUpdate(e: SubmitEvent) {
		e.preventDefault();
		if (!editingJob) return;
		editSubmitting = true;
		editError = null;
		try {
			const updated = await updateJob(editingJob.id, {
				name: editName.trim(),
				schedule: editSchedule.trim(),
				prompt: editPrompt.trim(),
				profile: editProfile.trim() || undefined,
				delivery: editDelivery.trim() || undefined,
				enabled: editEnabled
			});
			jobsActivity.jobs = jobs.map((j) => (j.id === updated.id ? { ...j, ...updated } : j));
			editModalOpen = false;
			toast.success('Job updated');
		} catch (err) {
			editError = err instanceof Error ? err.message : 'Failed to update job';
		} finally {
			editSubmitting = false;
		}
	}

	async function toggleEnabled(job: Job) {
		const nextState = !job.enabled;
		try {
			const updated = await updateJob(job.id, {
				name: job.name,
				schedule: job.schedule,
				prompt: job.prompt,
				profile: job.profile,
				delivery: job.delivery,
				enabled: nextState
			});
			jobsActivity.jobs = jobs.map((j) => (j.id === updated.id ? { ...j, ...updated } : j));
			toast.success(nextState ? 'Job enabled' : 'Job disabled');
		} catch (err) {
			toast.danger('Failed to toggle job state');
		}
	}

	async function handleDelete(job: Job) {
		if (!confirm(`Delete background job "${job.name}"?`)) return;
		try {
			await deleteJob(job.id);
			jobsActivity.jobs = jobs.filter((j) => j.id !== job.id);
			toast.success('Job deleted');
		} catch (err) {
			toast.danger('Failed to delete job');
		}
	}

	async function handleRunNow(job: Job) {
		runningJobIds = { ...runningJobIds, [job.id]: true };
		try {
			await runJob(job.id);
			await refreshJobs();
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Failed to start job run');
		} finally {
			runningJobIds = { ...runningJobIds, [job.id]: false };
		}
	}

	async function openHistory(job: Job) {
		historyJob = job;
		historyRuns = [];
		historyLoading = true;
		historyModalOpen = true;
		try {
			historyRuns = await listJobRuns(job.id);
		} catch (err) {
			toast.danger('Failed to load run history');
		} finally {
			historyLoading = false;
		}
	}

	function statusTone(status: string): 'neutral' | 'success' | 'warning' | 'danger' {
		switch (status) {
			case 'finished':
				return 'success';
			case 'running':
				return 'warning';
			case 'error':
				return 'danger';
			default:
				return 'neutral';
		}
	}

	function formatDate(val: string): string {
		try {
			return new Date(val).toLocaleString(undefined, {
				month: 'short',
				day: 'numeric',
				hour: '2-digit',
				minute: '2-digit',
				second: '2-digit'
			});
		} catch {
			return val;
		}
	}
</script>

<svelte:head>
	<title>Jobs - Bulle</title>
</svelte:head>

<Stack gap="section">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<PageHeader
			title="Jobs"
			description="Tasks your assistant runs on a schedule."
		/>
		<Button icon={icons.plus} onclick={openCreate}>New Job</Button>
	</div>

	{#if error}
		<Alert tone="danger">{error}</Alert>
	{/if}

	{#if loading}
		<div class="flex min-h-[30vh] items-center justify-center">
			<Spinner />
		</div>
	{:else if jobs.length === 0}
		<EmptyState
			icon={icons.clock}
			title="No scheduled jobs"
			description="Set up scheduled background jobs to perform regular analyses, summaries, or reports."
		>
			<Button icon={icons.plus} onclick={openCreate}>Create your first job</Button>
		</EmptyState>
	{:else}
		<div class="grid grid-cols-1 gap-3">
			{#each jobs as job (job.id)}
				<Card class="flex flex-col gap-3 p-5 {job.enabled ? '' : 'opacity-60'}">
					<div class="flex items-start justify-between gap-4">
						<div class="flex min-w-0 flex-col gap-0.5">
							<span class="truncate text-fc-base font-medium text-fc-fg">{job.name}</span>
							<span class="flex items-center gap-2 text-fc-sm text-fc-fg-muted">
								{describeSchedule(job.schedule)}
								{#if job.running}
									<span class="flex items-center gap-1.5 text-fc-warning">
										<span class="size-2 animate-pulse rounded-full bg-fc-warning" aria-hidden="true"></span>
										Running
									</span>
								{/if}
							</span>
						</div>
						<Switch
							checked={job.enabled}
							aria-label={`Toggle ${job.name}`}
							onchange={() => toggleEnabled(job)}
						/>
					</div>

					<p class="line-clamp-2 text-fc-sm text-fc-fg-muted">{job.prompt}</p>

					<div class="-ml-2 flex items-center gap-1">
						<Button
							variant="ghost"
							size="sm"
							icon={icons.refresh}
							disabled={runningJobIds[job.id] || job.running}
							onclick={() => handleRunNow(job)}
						>
							Run now
						</Button>
						{#if job.last_session_id}
							<Button variant="ghost" size="sm" icon={CHAT_ICONS.chat} href={`/s/${job.last_session_id}`}>
								{job.running ? 'Watch' : 'View output'}
							</Button>
						{/if}
						<IconButton variant="ghost" aria-label={`Runs of ${job.name}`} onclick={() => openHistory(job)}>
							<Icon icon={icons.history} size={16} />
						</IconButton>
						<IconButton variant="ghost" aria-label={`Edit ${job.name}`} onclick={() => openEdit(job)}>
							<Icon icon={icons.edit} size={16} />
						</IconButton>
						<IconButton
							variant="ghost"
							class="ml-auto"
							aria-label={`Delete ${job.name}`}
							onclick={() => handleDelete(job)}
						>
							<Icon icon={icons.remove} size={16} />
						</IconButton>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</Stack>

<Modal bind:open={createModalOpen} title="Create Background Job" size="md">
	<form onsubmit={handleCreate} class="flex flex-col gap-4">
		{#if createError}
			<Alert tone="danger">{createError}</Alert>
		{/if}

		<Field label="Name">
			<Input
				bind:value={newName}
				placeholder="e.g. daily-summary"
				required
			/>
		</Field>

		<Field label="Schedule (cron expression)" helper="Standard 5-part cron or macros like @hourly, @daily">
			<Input
				bind:value={newSchedule}
				placeholder="0 9 * * *"
				required
				class="font-mono text-fc-sm"
			/>
		</Field>

		<Field label="Prompt" helper="The instruction given to Bulle on every scheduled execution">
			<Textarea
				bind:value={newPrompt}
				rows={4}
				placeholder="e.g. Check for urgent incoming messages and summarize top priorities"
				required
			/>
		</Field>

		<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
			<Field label="Profile (optional)">
				<Input
					bind:value={newProfile}
					placeholder="default"
				/>
			</Field>

			<Field label="Delivery">
				<Select bind:value={newDelivery} disabled={false}>
					<option value="">None</option>
					<option value="push">Push notification</option>
				</Select>
			</Field>
		</div>

		<div class="flex items-center justify-between rounded-fc-md bg-fc-surface p-3">
			<span class="text-fc-sm font-medium text-fc-fg">Enable job immediately</span>
			<Switch bind:checked={newEnabled} aria-label="Enable job" />
		</div>

		<div class="mt-2 flex justify-end gap-2">
			<Button
				variant="outline"
				type="button"
				onclick={() => {
					createModalOpen = false;
				}}
			>
				Cancel
			</Button>
			<Button variant="primary" type="submit" disabled={createSubmitting}>
				{createSubmitting ? 'Creating...' : 'Create Job'}
			</Button>
		</div>
	</form>
</Modal>

<Modal bind:open={editModalOpen} title={`Edit Job: ${editingJob?.name || ''}`} size="md">
	<form onsubmit={handleUpdate} class="flex flex-col gap-4">
		{#if editError}
			<Alert tone="danger">{editError}</Alert>
		{/if}

		<Field label="Name">
			<Input
				bind:value={editName}
				required
			/>
		</Field>

		<Field label="Schedule (cron expression)" helper="Standard 5-part cron or macros like @hourly, @daily">
			<Input
				bind:value={editSchedule}
				required
				class="font-mono text-fc-sm"
			/>
		</Field>

		<Field label="Prompt">
			<Textarea
				bind:value={editPrompt}
				rows={4}
				required
			/>
		</Field>

		<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
			<Field label="Profile (optional)">
				<Input
					bind:value={editProfile}
					placeholder="default"
				/>
			</Field>

			<Field label="Delivery">
				<Select bind:value={editDelivery} disabled={editLegacyDelivery}>
					<option value="">None</option>
					<option value="push">Push notification</option>
					{#if editLegacyDelivery}
						<option value={editDelivery}>{editDelivery} (legacy)</option>
					{/if}
				</Select>
			</Field>
		</div>

		<div class="flex items-center justify-between rounded-fc-md bg-fc-surface p-3">
			<span class="text-fc-sm font-medium text-fc-fg">Active</span>
			<Switch bind:checked={editEnabled} aria-label="Active" />
		</div>

		<div class="mt-2 flex justify-end gap-2">
			<Button
				variant="outline"
				type="button"
				onclick={() => {
					editModalOpen = false;
				}}
			>
				Cancel
			</Button>
			<Button variant="primary" type="submit" disabled={editSubmitting}>
				{editSubmitting ? 'Saving...' : 'Save Changes'}
			</Button>
		</div>
	</form>
</Modal>

<Modal bind:open={historyModalOpen} title={historyJob?.name || 'Runs'} size="lg">
	<div class="flex flex-col gap-4">
		{#if historyLoading}
			<div class="flex min-h-[20vh] items-center justify-center">
				<Spinner />
			</div>
		{:else if historyRuns.length === 0}
			<EmptyState
				icon={icons.history}
				title="No runs recorded"
				description="This job has not been executed yet. Use 'Run now' to trigger an immediate test run."
			/>
		{:else}
			<div class="flex flex-col gap-1 overflow-hidden rounded-fc-md">
				{#each historyRuns as run (run.id)}
					<div class="flex flex-col gap-2 p-3 sm:flex-row sm:items-center sm:justify-between bg-fc-surface/30">
						<div class="flex items-center gap-3">
							<StatusDot tone={statusTone(run.status)} label={run.status} />
							<div class="flex flex-col">
								<span class="font-medium text-fc-sm text-fc-fg capitalize">{run.status}</span>
								<span class="text-fc-xs text-fc-fg-muted">{formatDate(run.started_at)}</span>
							</div>
						</div>

						<div class="text-fc-xs text-fc-fg-muted">
							{#if !run.finished_at}
								<span class="text-fc-warning">In progress</span>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}

		<div class="flex justify-end pt-2">
			<Button
				variant="outline"
				type="button"
				onclick={() => {
					historyModalOpen = false;
				}}
			>
				Close
			</Button>
		</div>
	</div>
</Modal>
