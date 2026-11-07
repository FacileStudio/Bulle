<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
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
		errorText,
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
	import {
		ideasActivity,
		isCurrentSpace,
		pollWhileGenerating,
		refreshIdeas
	} from '$lib/ideas-activity.svelte';
	import { spaces } from '$lib/spaces.svelte';
	import { formatDate } from '$lib/i18n/format';

	const HOURS = Array.from({ length: 24 }, (_, hour) => hour);
	const FIELD = 'border-0 bg-fc-component';
	const SWITCH = '[&>span]:border-transparent';

	let busy = $state<Record<string, boolean>>({});
	let requesting = $state(false);
	let settingsOpen = $state(false);
	let steering = $state('');
	let sentSteering: string | null = null;
	const synced = new Set<string>();

	const spaceId = $derived(spaces.currentId);
	const settings = $derived(ideasActivity.settings);
	const groups = $derived(groupIdeas(ideasActivity.ideas, Date.now()));
	const generating = $derived(ideasActivity.generating || requesting);
	const empty = $derived(groups.fresh.length === 0 && groups.saved.length === 0);
	const neverGenerated = $derived(empty && !ideasActivity.lastBatchAt && !generating);
	const lastBatch = $derived(
		ideasActivity.lastBatchAt ? t('ideas.lastBatch', { time: formatTime(ideasActivity.lastBatchAt) }) : ''
	);

	onMount(() => {
		void refreshIdeas();
	});

	const storedSteering = $derived(settings?.steering ?? '');

	$effect(() => {
		steering = storedSteering;
		sentSteering = null;
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
		return formatDate(value, {
			weekday: 'short',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function hourLabel(hour: number): string {
		return `${String(hour).padStart(2, '0')}:00`;
	}

	async function saveSettings(patch: SettingsPatch): Promise<boolean> {
		const target = spaceId;
		if (!target) return false;
		try {
			const saved = await updateIdeaSettings(target, patch);
			if (isCurrentSpace(target)) ideasActivity.settings = saved;
			return true;
		} catch (err) {
			toast.danger(errorText(err, t('ideas.error.saveSettings')));
			return false;
		}
	}

	async function toggleSetting(event: Event, key: 'enabled' | 'push') {
		const input = event.currentTarget as HTMLInputElement;
		const value = input.checked;
		if (!(await saveSettings({ [key]: value }))) input.checked = !value;
	}

	async function saveHour(event: Event) {
		const select = event.currentTarget as HTMLSelectElement;
		if (!(await saveSettings({ hour: Number(select.value) })) && settings)
			select.value = String(settings.hour);
	}

	function saveSteering() {
		const text = steering.trim();
		if (!settings || text === settings.steering || text === sentSteering) return;
		sentSteering = text;
		void saveSettings({ steering: text }).then((ok) => {
			if (!ok) sentSteering = null;
		});
	}

	async function generate() {
		const target = spaceId;
		if (!target || generating) return;
		requesting = true;
		try {
			await generateIdeas(target);
			if (isCurrentSpace(target)) ideasActivity.generating = true;
		} catch (err) {
			if (err instanceof ApiError && err.status === 409) toast.info(t('ideas.toast.alreadyThinking'));
			else if (err instanceof ApiError && err.status === 429)
				toast.warning(t('ideas.toast.dailyLimit'));
			else toast.danger(errorText(err, t('ideas.error.startThinking')));
		} finally {
			requesting = false;
		}
	}

	function drop(id: string) {
		ideasActivity.ideas = ideasActivity.ideas.filter((item) => item.id !== id);
	}

	function setBusy(id: string, value: boolean) {
		busy = { ...busy, [id]: value };
	}

	async function patch(idea: Idea, change: IdeaPatch) {
		if (busy[idea.id]) return;
		setBusy(idea.id, true);
		try {
			const updated = await updateIdea(idea.id, change);
			ideasActivity.ideas = replaceById(ideasActivity.ideas, updated);
		} catch (err) {
			if (err instanceof ApiError && (err.status === 409 || err.status === 404)) {
				drop(idea.id);
				toast.info(t('ideas.toast.notOpen'));
			} else toast.danger(errorText(err, t('ideas.error.updateIdea')));
		} finally {
			setBusy(idea.id, false);
		}
	}

	async function start(idea: Idea) {
		if (busy[idea.id]) return;
		setBusy(idea.id, true);
		try {
			const { session_id } = await startIdea(idea.id);
			drop(idea.id);
			void sessions.refresh();
			await goto(`/s/${encodeURIComponent(session_id)}`);
		} catch (err) {
			if (err instanceof ApiError && err.status === 429)
				toast.warning(t('ideas.toast.tooManyChats'));
			else toast.danger(errorText(err, t('ideas.error.startChat')));
		} finally {
			setBusy(idea.id, false);
		}
	}

	function setGoals(goals: Goal[]) {
		ideasActivity.goals = goals;
	}
</script>

<svelte:head>
	<title>{t('ideas.pageTitle')}</title>
</svelte:head>

<Stack gap="section">
	<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
		<PageHeader
			title={t('ideas.title')}
			description={spaces.current
				? t('ideas.subtitleSpace', { space: spaces.current.name })
				: t('ideas.subtitle')}
		/>
		<div class="flex items-center gap-2">
			<Button
				class="h-11"
				icon={generating ? undefined : icons.bolt}
				disabled={!spaceId || !settings || generating}
				aria-busy={generating}
				onclick={generate}
			>
				{#if generating}
					<Spinner size="sm" />
					{t('ideas.thinking')}
				{:else}
					{t('ideas.thinkMore')}
				{/if}
			</Button>
			<IconButton
				variant="ghost"
				aria-label={t('ideas.settings.title')}
				disabled={!settings}
				onclick={() => (settingsOpen = true)}
			>
				<Icon icon={icons.settings} size={18} />
			</IconButton>
		</div>
	</div>

	{#if !ideasActivity.loaded}
		<div class="flex min-h-[30vh] items-center justify-center">
			<Spinner />
		</div>
	{:else if !settings}
		<EmptyState
			icon={icons.warning}
			title={t('ideas.loadError')}
			description={ideasActivity.error ?? t('ideas.genericError')}
		>
			<Button class="h-11" icon={icons.refresh} onclick={() => void refreshIdeas()}>{t('common.tryAgain')}</Button>
		</EmptyState>
	{:else}
		{#if ideasActivity.error}
			<Alert tone="danger">{ideasActivity.error}</Alert>
		{/if}

		<Input
			bind:value={steering}
			class={FIELD}
			aria-label={t('ideas.steering')}
			placeholder={t('ideas.steering')}
			maxlength={500}
			enterkeyhint="done"
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
				title={t('ideas.empty.title')}
				description={t('ideas.empty.description')}
			>
				<Button class="h-11" icon={icons.bolt} disabled={!spaceId} onclick={generate}>
					{t('ideas.thinkNow')}
				</Button>
			</EmptyState>
		{:else if empty && generating}
			<EmptyState icon={icons.bolt} title={t('ideas.busy.title')} description={t('ideas.busy.description')}>
				<Spinner />
			</EmptyState>
		{:else}
			{#if groups.fresh.length > 0}
				<section class="flex flex-col gap-3" aria-labelledby="new-heading">
					<h2 id="new-heading" class="sr-only">{t('ideas.newIdeas')}</h2>
					{#each groups.fresh as idea (idea.id)}
						<IdeaCard {idea} busy={busy[idea.id]} onstart={start} onpatch={patch} />
					{/each}
				</section>
			{:else if !generating}
				<EmptyState
					icon={icons.check}
					title={t('ideas.nothingNew.title')}
					description={`${t('ideas.nothingNew.description')} ${lastBatch}`.trim()}
				/>
			{/if}

			{#if groups.saved.length > 0}
				<section class="flex flex-col gap-3" aria-labelledby="saved-heading">
					<h2 id="saved-heading" class="text-fc-md font-medium text-fc-fg">{t('ideas.saved')}</h2>
					{#each groups.saved as idea (idea.id)}
						<IdeaCard {idea} busy={busy[idea.id]} onstart={start} onpatch={patch} />
					{/each}
				</section>
			{/if}

			{#if groups.fresh.length > 0 || generating}
				<p class="text-center text-fc-sm text-fc-fg-muted" aria-live="polite">
					{generating ? t('ideas.thinkingMore') : `${t('ideas.allForNow')} ${lastBatch}`.trim()}
				</p>
			{/if}
		{/if}
	{/if}
</Stack>

<Modal bind:open={settingsOpen} title={t('ideas.settings.title')} size="sm">
	{#if settings}
		<div class="flex flex-col gap-4">
			<SettingsRow label={t('ideas.settings.morning')} description={t('ideas.settings.morningHint')}>
				<Switch
					checked={settings.enabled}
					class={SWITCH}
					aria-label={t('ideas.settings.morning')}
					onchange={(event) => toggleSetting(event, 'enabled')}
				/>
			</SettingsRow>
			<SettingsRow
				label={t('ideas.settings.hour')}
				description={t('ideas.settings.hourHint', { timezone: settings.timezone })}
				stacked
			>
				<Select
					value={String(settings.hour)}
					class="border-0 bg-fc-surface"
					aria-label={t('ideas.settings.hour')}
					disabled={!settings.enabled}
					onchange={saveHour}
				>
					{#each HOURS as hour (hour)}
						<option value={String(hour)}>{hourLabel(hour)}</option>
					{/each}
				</Select>
			</SettingsRow>
			<SettingsRow label={t('ideas.settings.push')} description={t('ideas.settings.pushHint')}>
				<Switch
					checked={settings.push}
					class={SWITCH}
					aria-label={t('ideas.settings.push')}
					disabled={!settings.enabled}
					onchange={(event) => toggleSetting(event, 'push')}
				/>
			</SettingsRow>
		</div>
	{/if}
</Modal>
