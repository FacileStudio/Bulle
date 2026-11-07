<script lang="ts">
	import { onMount } from 'svelte';
	import { t } from '$lib/i18n/index.svelte';
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
		Select,
		Spinner,
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
	import NotificationPrompt from '$lib/components/NotificationPrompt.svelte';
	import { formatDate } from '$lib/i18n/format';
	import { jobsActivity, refreshJobs } from '$lib/jobs-activity.svelte';
	import { spaces } from '$lib/spaces.svelte';

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
			jobsActivity.jobs = await listJobs(spaces.currentId);
		} catch (err) {
			error = err instanceof Error ? err.message : t('jobs.loadFailed');
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
				space_id: spaces.currentId ?? undefined,
				name: newName.trim(),
				schedule: newSchedule.trim(),
				prompt: newPrompt.trim(),
				profile: newProfile.trim() || undefined,
				delivery: newDelivery.trim() || undefined,
				enabled: newEnabled
			});
			jobsActivity.jobs = [created, ...jobs];
			createModalOpen = false;
			toast.success(t('jobs.created'));
		} catch (err) {
			createError = err instanceof Error ? err.message : t('jobs.createFailed');
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
			toast.success(t('jobs.updated'));
		} catch (err) {
			editError = err instanceof Error ? err.message : t('jobs.updateFailed');
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
			toast.success(nextState ? t('jobs.enabled') : t('jobs.disabled'));
		} catch (err) {
			toast.danger(t('jobs.toggleFailed'));
		}
	}

	async function handleDelete(job: Job) {
		if (!confirm(t('jobs.deleteConfirm', { name: job.name }))) return;
		try {
			await deleteJob(job.id);
			jobsActivity.jobs = jobs.filter((j) => j.id !== job.id);
			toast.success(t('jobs.deleted'));
		} catch (err) {
			toast.danger(t('jobs.deleteFailed'));
		}
	}

	async function handleRunNow(job: Job) {
		runningJobIds = { ...runningJobIds, [job.id]: true };
		try {
			await runJob(job.id);
			await refreshJobs();
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('jobs.runFailed'));
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
			toast.danger(t('jobs.historyFailed'));
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

	function runStarted(val: string): string {
		try {
			return formatDate(val, {
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

	function statusLabel(status: string): string {
		switch (status) {
			case 'finished':
				return t('jobs.status.finished');
			case 'running':
				return t('jobs.status.running');
			case 'error':
				return t('jobs.status.error');
			default:
				return status;
		}
	}
</script>

<section class="flex flex-col gap-6" aria-labelledby="jobs-title">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<div class="flex flex-col gap-1">
			<h2 id="jobs-title" class="text-fc-lg font-semibold text-fc-fg">{t('jobs.title')}</h2>
			<p class="text-fc-sm text-fc-fg-muted">
				{spaces.current
					? t('jobs.subtitleSpace', { space: spaces.current.name })
					: t('jobs.subtitle')}
			</p>
		</div>
		<Button icon={icons.plus} size="lg" onclick={openCreate}>{t('jobs.new')}</Button>
	</div>

	<NotificationPrompt />

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
			title={t('jobs.empty.title')}
			description={t('jobs.empty.description')}
		>
			<Button icon={icons.plus} onclick={openCreate}>{t('jobs.empty.create')}</Button>
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
										{t('jobs.running')}
									</span>
								{/if}
							</span>
						</div>
						<Switch
							checked={job.enabled}
							aria-label={t('jobs.toggleAria', { name: job.name })}
							onchange={() => toggleEnabled(job)}
						/>
					</div>

					<p class="line-clamp-2 text-fc-sm text-fc-fg-muted">{job.prompt}</p>

					<div class="-ml-2 flex items-center gap-1">
						<Button
							variant="ghost"
							class="h-11"
							icon={icons.refresh}
							disabled={runningJobIds[job.id] || job.running}
							onclick={() => handleRunNow(job)}
						>
							{t('jobs.runNow')}
						</Button>
						{#if job.last_session_id}
							<Button variant="ghost" class="h-11" icon={CHAT_ICONS.chat} href={`/s/${job.last_session_id}`}>
								{job.running ? t('jobs.watch') : t('jobs.viewOutput')}
							</Button>
						{/if}
						<IconButton variant="ghost" aria-label={t('jobs.runsOf', { name: job.name })} onclick={() => openHistory(job)}>
							<Icon icon={icons.history} size={16} />
						</IconButton>
						<IconButton variant="ghost" aria-label={t('jobs.editAria', { name: job.name })} onclick={() => openEdit(job)}>
							<Icon icon={icons.edit} size={16} />
						</IconButton>
						<IconButton
							variant="ghost"
							class="ml-auto"
							aria-label={t('jobs.deleteAria', { name: job.name })}
							onclick={() => handleDelete(job)}
						>
							<Icon icon={icons.remove} size={16} />
						</IconButton>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</section>

<Modal bind:open={createModalOpen} title={t('jobs.createTitle')} size="md">
	<form onsubmit={handleCreate} class="flex flex-col gap-4">
		{#if createError}
			<Alert tone="danger">{createError}</Alert>
		{/if}

		<Field label={t('jobs.field.name')}>
			<Input
				class="border-0 bg-fc-surface"
				bind:value={newName}
				placeholder={t('jobs.field.namePlaceholder')}
				required
			/>
		</Field>

		<Field label={t('jobs.field.schedule')} helper={t('jobs.field.scheduleHelp')}>
			<Input
				bind:value={newSchedule}
				placeholder="0 9 * * *"
				required
				class="border-0 bg-fc-surface font-fc-mono text-fc-sm"
			/>
		</Field>

		<Field label={t('jobs.field.prompt')} helper={t('jobs.field.promptHelp')}>
			<Textarea
				bind:value={newPrompt}
				rows={4}
				class="border-0 bg-fc-surface"
				placeholder={t('jobs.field.promptPlaceholder')}
				required
			/>
		</Field>

		<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
			<Field label={t('jobs.field.profile')}>
				<Input
					class="border-0 bg-fc-surface"
					bind:value={newProfile}
					placeholder="default"
				/>
			</Field>

			<Field label={t('jobs.field.delivery')}>
				<Select class="border-0 bg-fc-surface" bind:value={newDelivery} disabled={false}>
					<option value="">{t('jobs.delivery.none')}</option>
					<option value="push">{t('jobs.delivery.push')}</option>
				</Select>
			</Field>
		</div>

		<div class="flex min-h-11 items-center justify-between rounded-fc-md bg-fc-surface p-3">
			<span class="text-fc-sm font-medium text-fc-fg">{t('jobs.enableNow')}</span>
			<Switch bind:checked={newEnabled} aria-label={t('jobs.enableAria')} />
		</div>

		<div class="mt-2 flex justify-end gap-2">
			<Button
				variant="ghost"
				type="button"
				onclick={() => {
					createModalOpen = false;
				}}
			>
				{t('common.cancel')}
			</Button>
			<Button variant="primary" type="submit" disabled={createSubmitting}>
				{createSubmitting ? t('jobs.creating') : t('jobs.createSubmit')}
			</Button>
		</div>
	</form>
</Modal>

<Modal bind:open={editModalOpen} title={t('jobs.editTitle', { name: editingJob?.name || '' })} size="md">
	<form onsubmit={handleUpdate} class="flex flex-col gap-4">
		{#if editError}
			<Alert tone="danger">{editError}</Alert>
		{/if}

		<Field label={t('jobs.field.name')}>
			<Input
				class="border-0 bg-fc-surface"
				bind:value={editName}
				required
			/>
		</Field>

		<Field label={t('jobs.field.schedule')} helper={t('jobs.field.scheduleHelp')}>
			<Input
				bind:value={editSchedule}
				required
				class="border-0 bg-fc-surface font-fc-mono text-fc-sm"
			/>
		</Field>

		<Field label={t('jobs.field.prompt')}>
			<Textarea
				bind:value={editPrompt}
				rows={4}
				class="border-0 bg-fc-surface"
				required
			/>
		</Field>

		<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
			<Field label={t('jobs.field.profile')}>
				<Input
					class="border-0 bg-fc-surface"
					bind:value={editProfile}
					placeholder="default"
				/>
			</Field>

			<Field label={t('jobs.field.delivery')}>
				<Select class="border-0 bg-fc-surface" bind:value={editDelivery} disabled={editLegacyDelivery}>
					<option value="">{t('jobs.delivery.none')}</option>
					<option value="push">{t('jobs.delivery.push')}</option>
					{#if editLegacyDelivery}
						<option value={editDelivery}>{t('jobs.delivery.legacy', { value: editDelivery })}</option>
					{/if}
				</Select>
			</Field>
		</div>

		<div class="flex min-h-11 items-center justify-between rounded-fc-md bg-fc-surface p-3">
			<span class="text-fc-sm font-medium text-fc-fg">{t('jobs.active')}</span>
			<Switch bind:checked={editEnabled} aria-label={t('jobs.active')} />
		</div>

		<div class="mt-2 flex justify-end gap-2">
			<Button
				variant="ghost"
				type="button"
				onclick={() => {
					editModalOpen = false;
				}}
			>
				{t('common.cancel')}
			</Button>
			<Button variant="primary" type="submit" disabled={editSubmitting}>
				{editSubmitting ? t('jobs.saving') : t('jobs.saveChanges')}
			</Button>
		</div>
	</form>
</Modal>

<Modal bind:open={historyModalOpen} title={historyJob?.name || t('jobs.historyTitle')} size="lg">
	<div class="flex flex-col gap-4">
		{#if historyLoading}
			<div class="flex min-h-[20vh] items-center justify-center">
				<Spinner />
			</div>
		{:else if historyRuns.length === 0}
			<EmptyState
				icon={icons.history}
				title={t('jobs.history.empty')}
				description={t('jobs.history.emptyDescription')}
			/>
		{:else}
			<div class="flex flex-col gap-1 overflow-hidden rounded-fc-md">
				{#each historyRuns as run (run.id)}
					<div class="flex flex-col gap-2 p-3 sm:flex-row sm:items-center sm:justify-between bg-fc-surface/30">
						<div class="flex items-center gap-3">
							<StatusDot tone={statusTone(run.status)} label={statusLabel(run.status)} />
							<div class="flex flex-col">
								<span class="font-medium text-fc-sm text-fc-fg">{statusLabel(run.status)}</span>
								<span class="text-fc-xs text-fc-fg-muted">{runStarted(run.started_at)}</span>
							</div>
						</div>

						<div class="text-fc-xs text-fc-fg-muted">
							{#if !run.finished_at}
								<span class="text-fc-warning">{t('jobs.history.inProgress')}</span>
							{/if}
						</div>
					</div>
				{/each}
			</div>
		{/if}

		<div class="flex justify-end pt-2">
			<Button
				variant="ghost"
				type="button"
				onclick={() => {
					historyModalOpen = false;
				}}
			>
				{t('common.close')}
			</Button>
		</div>
	</div>
</Modal>
