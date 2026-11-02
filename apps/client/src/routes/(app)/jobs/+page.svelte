<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Alert,
		Badge,
		Button,
		Card,
		EmptyState,
		Field,
		Icon,
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

	let jobs = $state<Job[]>([]);
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
		loading = true;
		error = null;
		try {
			jobs = await listJobs();
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
			jobs = [created, ...jobs];
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
			jobs = jobs.map((j) => (j.id === updated.id ? updated : j));
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
			jobs = jobs.map((j) => (j.id === updated.id ? updated : j));
			toast.success(nextState ? 'Job enabled' : 'Job disabled');
		} catch (err) {
			toast.danger('Failed to toggle job state');
		}
	}

	async function handleDelete(job: Job) {
		if (!confirm(`Delete background job "${job.name}"?`)) return;
		try {
			await deleteJob(job.id);
			jobs = jobs.filter((j) => j.id !== job.id);
			toast.success('Job deleted');
		} catch (err) {
			toast.danger('Failed to delete job');
		}
	}

	async function handleRunNow(job: Job) {
		runningJobIds = { ...runningJobIds, [job.id]: true };
		try {
			await runJob(job.id);
			toast.success(`Run started for "${job.name}"`);
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
			title="Background Jobs"
			description="Automated cron tasks executed by Bulle on a schedule."
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
		<div class="grid grid-cols-1 gap-4">
			{#each jobs as job (job.id)}
				<Card class="flex flex-col gap-4 p-5">
					<div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
						<div class="flex items-start gap-3">
							<div class="mt-0.5 flex size-9 shrink-0 items-center justify-center rounded-fc-md border border-fc-border bg-fc-surface text-fc-fg">
								<Icon icon={icons.clock} size={18} />
							</div>
							<div class="flex flex-col gap-1">
								<div class="flex flex-wrap items-center gap-2">
									<span class="font-medium text-fc-base text-fc-fg">{job.name}</span>
									<Badge tone={job.enabled ? 'success' : 'neutral'}>
										{job.enabled ? 'Active' : 'Disabled'}
									</Badge>
									{#if job.delivery}
										<Badge tone="neutral">
											delivery: {job.delivery}
										</Badge>
									{/if}
								</div>
								<div class="flex items-center gap-2 font-mono text-fc-xs text-fc-fg-muted">
									<Icon icon={icons.calendar} size={13} />
									<span>{job.schedule}</span>
								</div>
							</div>
						</div>

						<div class="flex flex-wrap items-center gap-2">
							<div class="flex items-center gap-2 pr-2">
								<span class="text-fc-xs text-fc-fg-muted">
									{job.enabled ? 'Enabled' : 'Disabled'}
								</span>
								<Switch
									checked={job.enabled}
									aria-label={`Toggle ${job.name}`}
									onchange={() => toggleEnabled(job)}
								/>
							</div>

							<Button
								variant="outline"
								size="sm"
								icon={icons.refresh}
								disabled={runningJobIds[job.id]}
								onclick={() => handleRunNow(job)}
							>
								{runningJobIds[job.id] ? 'Running...' : 'Run now'}
							</Button>

							<Button
								variant="ghost"
								size="sm"
								icon={icons.history}
								onclick={() => openHistory(job)}
							>
								Runs
							</Button>

							<Button
								variant="ghost"
								size="sm"
								icon={icons.edit}
								onclick={() => openEdit(job)}
							>
								Edit
							</Button>

							<Button
								variant="ghost"
								size="sm"
								icon={icons.remove}
								onclick={() => handleDelete(job)}
							>
								Delete
							</Button>
						</div>
					</div>

					<div class="rounded-fc-md border border-fc-border/60 bg-fc-surface/40 p-3 text-fc-sm text-fc-fg-muted">
						<div class="mb-1 text-fc-xs font-medium uppercase tracking-wider text-fc-fg-subtle">Prompt</div>
						<div class="whitespace-pre-wrap line-clamp-3">{job.prompt}</div>
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

		<div class="flex items-center justify-between rounded-fc-md border border-fc-border p-3">
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

		<div class="flex items-center justify-between rounded-fc-md border border-fc-border p-3">
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

<Modal bind:open={historyModalOpen} title={`Run History: ${historyJob?.name || ''}`} size="lg">
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
			<div class="flex flex-col divide-y divide-fc-border rounded-fc-md border border-fc-border overflow-hidden">
				{#each historyRuns as run (run.id)}
					<div class="flex flex-col gap-2 p-3 sm:flex-row sm:items-center sm:justify-between bg-fc-surface/30">
						<div class="flex items-center gap-3">
							<StatusDot tone={statusTone(run.status)} label={run.status} />
							<div class="flex flex-col">
								<div class="flex items-center gap-2">
									<span class="font-medium text-fc-sm text-fc-fg capitalize">{run.status}</span>
									<span class="font-mono text-fc-xs text-fc-fg-subtle">ID: {run.id.slice(0, 8)}</span>
								</div>
								<span class="text-fc-xs text-fc-fg-muted">
									Started: {formatDate(run.started_at)}
								</span>
							</div>
						</div>

						<div class="text-fc-xs text-fc-fg-muted">
							{#if run.finished_at}
								<span>Finished: {formatDate(run.finished_at)}</span>
							{:else}
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
