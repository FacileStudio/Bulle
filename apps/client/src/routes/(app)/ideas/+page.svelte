<script lang="ts">
	import { onMount, untrack } from 'svelte';
	import { goto } from '$app/navigation';
	import {
		Alert,
		Button,
		EmptyState,
		Icon,
		IconButton,
		Input,
		Modal,
		PageHeader,
		Select,
		SettingsRow,
		Spinner,
		Stack,
		Switch,
		icons,
		toast
	} from '@facile/muse';
	import { ApiError } from '$lib/backend';
	import { sessions } from '$lib/chat/sessions.svelte';
	import GoalsPanel from '$lib/components/ideas/GoalsPanel.svelte';
	import IdeaCard from '$lib/components/ideas/IdeaCard.svelte';
	import {
		generateIdeas,
		groupIdeas,
		replaceById,
		startIdea,
		timezoneToSync,
		updateIdea,
		updateIdeaSettings,
		type Goal,
		type Idea,
		type IdeaPatch,
		type SettingsPatch
	} from '$lib/ideas';
	import { ideasActivity, pollWhileGenerating, refreshIdeas } from '$lib/ideas-activity.svelte';
	import { spaces } from '$lib/spaces.svelte';

	const HOURS = Array.from({ length: 24 }, (_, hour) => hour);

	let busy = $state<Record<string, boolean>>({});
	let requesting = $state(false);
	let settingsOpen = $state(false);
	let steering = $state('');
	const synced = new Set<string>();

	const spaceId = $derived(spaces.currentId);
	const settings = $derived(ideasActivity.settings);
	const groups = $derived(groupIdeas(ideasActivity.ideas, Date.now()));
	const generating = $derived(ideasActivity.generating || requesting);
	const empty = $derived(groups.fresh.length === 0 && groups.saved.length === 0);
	const neverGenerated = $derived(empty && !ideasActivity.lastBatchAt && !generating);

	onMount(() => {
		void refreshIdeas();
	});

	const storedSteering = $derived(settings?.steering ?? '');

	$effect(() => {
		steering = storedSteering;
	});

	$effect(() => {
		if (ideasActivity.generating) return untrack(pollWhileGenerating);
	});

	$effect(() => {
		const current = settings;
		if (!current || synced.has(current.space_id)) return;
		synced.add(current.space_id);
		const timezone = timezoneToSync(current.timezone, Intl.DateTimeFormat().resolvedOptions().timeZone);
		if (timezone) untrack(() => void saveSettings({ timezone }));
	});

	function formatTime(value: string): string {
		return new Date(value).toLocaleString(undefined, {
			weekday: 'short',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function hourLabel(hour: number): string {
		return `${String(hour).padStart(2, '0')}:00`;
	}

	async function saveSettings(patch: SettingsPatch) {
		if (!spaceId) return;
		try {
			ideasActivity.settings = await updateIdeaSettings(spaceId, patch);
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not save the settings');
		}
	}

	function saveSteering() {
		const text = steering.trim();
		if (!settings || text === settings.steering) return;
		void saveSettings({ steering: text });
	}

	async function generate() {
		if (!spaceId || generating) return;
		requesting = true;
		try {
			await generateIdeas(spaceId);
			ideasActivity.generating = true;
		} catch (err) {
			if (err instanceof ApiError && err.status === 409) toast.info('Bulle is already thinking of ideas.');
			else if (err instanceof ApiError && err.status === 429)
				toast.warning('That is enough ideas for today. Try again tomorrow.');
			else toast.danger(err instanceof Error ? err.message : 'Could not start thinking');
		} finally {
			requesting = false;
		}
	}

	async function patch(idea: Idea, change: IdeaPatch) {
		busy = { ...busy, [idea.id]: true };
		try {
			ideasActivity.ideas = replaceById(ideasActivity.ideas, await updateIdea(idea.id, change));
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not update the idea');
		} finally {
			busy = { ...busy, [idea.id]: false };
		}
	}

	async function start(idea: Idea) {
		busy = { ...busy, [idea.id]: true };
		try {
			const { session_id } = await startIdea(idea.id);
			ideasActivity.ideas = ideasActivity.ideas.filter((item) => item.id !== idea.id);
			void sessions.refresh();
			await goto(`/s/${encodeURIComponent(session_id)}`);
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not start the chat');
			busy = { ...busy, [idea.id]: false };
		}
	}

	function setGoals(goals: Goal[]) {
		ideasActivity.goals = goals;
	}
</script>

<svelte:head>
	<title>Ideas - Bulle</title>
</svelte:head>

<Stack gap="section">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<PageHeader
			title="Ideas"
			description={spaces.current
				? `What Bulle could do for you in ${spaces.current.name}.`
				: 'What Bulle could do for you.'}
		/>
		<div class="flex items-center gap-2">
			<Button class="h-11" icon={icons.bolt} disabled={!spaceId || generating} onclick={generate}>
				{#if generating}
					<Spinner size="sm" />
					Thinking...
				{:else}
					Think of more
				{/if}
			</Button>
			<IconButton aria-label="Ideas settings" disabled={!settings} onclick={() => (settingsOpen = true)}>
				<Icon icon={icons.settings} size={18} />
			</IconButton>
		</div>
	</div>

	{#if ideasActivity.error}
		<Alert tone="danger">{ideasActivity.error}</Alert>
	{/if}

	{#if !ideasActivity.loaded}
		<div class="flex min-h-[30vh] items-center justify-center">
			<Spinner />
		</div>
	{:else}
		<Input
			bind:value={steering}
			aria-label="What should Bulle think about?"
			placeholder="What should Bulle think about?"
			maxlength={500}
			disabled={!settings}
			onblur={saveSteering}
			onkeydown={(event) => {
				if (event.key === 'Enter') saveSteering();
			}}
		/>

		{#if spaceId}
			<GoalsPanel {spaceId} goals={ideasActivity.goals} onchange={setGoals} />
		{/if}

		{#if neverGenerated}
			<EmptyState
				icon={icons.bolt}
				title="No ideas yet"
				description="Bulle will leave ideas here each morning."
			>
				<Button class="h-11" icon={icons.bolt} disabled={!spaceId} onclick={generate}>
					Think of ideas now
				</Button>
			</EmptyState>
		{:else if empty && generating}
			<EmptyState icon={icons.bolt} title="Bulle is thinking" description="Ideas will show up here in a moment.">
				<Spinner />
			</EmptyState>
		{:else}
			{#if groups.fresh.length > 0}
				<section class="flex flex-col gap-3" aria-label="New ideas">
					{#each groups.fresh as idea (idea.id)}
						<IdeaCard {idea} busy={busy[idea.id]} onstart={start} onpatch={patch} />
					{/each}
				</section>
			{:else if !generating}
				<EmptyState icon={icons.check} title="Nothing new" description="Bulle has no fresh ideas right now." />
			{/if}

			{#if groups.saved.length > 0}
				<section class="flex flex-col gap-3" aria-labelledby="saved-heading">
					<h2 id="saved-heading" class="text-fc-md font-medium text-fc-fg">Saved</h2>
					{#each groups.saved as idea (idea.id)}
						<IdeaCard {idea} busy={busy[idea.id]} onstart={start} onpatch={patch} />
					{/each}
				</section>
			{/if}

			<p class="text-center text-fc-sm text-fc-fg-muted">
				That's all for now.
				{#if ideasActivity.lastBatchAt}
					Last ideas {formatTime(ideasActivity.lastBatchAt)}.
				{/if}
			</p>
		{/if}
	{/if}
</Stack>

<Modal bind:open={settingsOpen} title="Ideas settings" size="sm">
	{#if settings}
		<div class="flex flex-col gap-4">
			<SettingsRow label="Morning ideas" description="Bulle thinks of ideas once a day.">
				<Switch
					checked={settings.enabled}
					aria-label="Morning ideas"
					onchange={(event) => saveSettings({ enabled: event.currentTarget.checked })}
				/>
			</SettingsRow>
			<SettingsRow label="Morning hour" description={`In your time zone, ${settings.timezone}.`} stacked>
				<Select
					value={String(settings.hour)}
					aria-label="Morning hour"
					disabled={!settings.enabled}
					onchange={(event) => saveSettings({ hour: Number(event.currentTarget.value) })}
				>
					{#each HOURS as hour (hour)}
						<option value={String(hour)}>{hourLabel(hour)}</option>
					{/each}
				</Select>
			</SettingsRow>
			<SettingsRow label="Push notification" description="One notification when the morning ideas are ready.">
				<Switch
					checked={settings.push}
					aria-label="Push notification"
					disabled={!settings.enabled}
					onchange={(event) => saveSettings({ push: event.currentTarget.checked })}
				/>
			</SettingsRow>
			<div class="flex justify-end">
				<Button variant="outline" class="h-11" onclick={() => (settingsOpen = false)}>Close</Button>
			</div>
		</div>
	{/if}
</Modal>
