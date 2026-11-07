<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, Button, Field, Input, Section, Spinner, toast } from '@facile/muse';
	import { getPreferences, updatePreferences, type ModelOption } from '$lib/preferences';
	import { t } from '$lib/i18n/index.svelte';

	const PAGE = 60;

	let catalog = $state<ModelOption[]>([]);
	let model = $state('');
	let saved = $state('');
	let instanceModel = $state('');
	let query = $state('');
	let loading = $state(true);
	let saving = $state(false);
	let loadError = $state<string | null>(null);

	onMount(async () => {
		try {
			const prefs = await getPreferences();
			catalog = prefs.catalog;
			model = prefs.model;
			saved = prefs.model;
			instanceModel = prefs.default_model;
		} catch (err) {
			loadError = err instanceof Error ? err.message : t('models.loadFailed');
		} finally {
			loading = false;
		}
	});

	const matches = $derived.by(() => {
		const needle = query.trim().toLowerCase();
		if (!needle) return catalog;
		return catalog.filter(
			(option) =>
				option.id.toLowerCase().includes(needle) || option.name.toLowerCase().includes(needle)
		);
	});
	const shown = $derived(matches.slice(0, PAGE));
	const current = $derived(catalog.find((option) => option.id === model));

	async function save() {
		saving = true;
		try {
			const prefs = await updatePreferences({ model });
			model = prefs.model;
			saved = prefs.model;
			toast.success(t('models.saved'));
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : t('models.saveFailed'));
		} finally {
			saving = false;
		}
	}
</script>

<Section
	card
	id="model"
	title={t('models.title')}
	description={t('models.description')}
>
	{#if loading}
		<div class="flex justify-center py-6"><Spinner /></div>
	{:else if loadError}
		<Alert tone="danger">{loadError}</Alert>
	{:else}
		<div class="flex flex-col gap-4">
			<div class="text-fc-sm text-fc-fg">
				{t('models.current')}
				<span class="font-mono">{model || t('models.serverDefault', { model: instanceModel })}</span>
				{#if current && current.name !== current.id}
					<span class="text-fc-fg-muted"> — {current.name}</span>
				{/if}
			</div>

			<Field label={t('models.searchLabel')} helper={t('models.count', { count: catalog.length })}>
				<Input
					bind:value={query}
					placeholder={t('models.searchPlaceholder')}
					autocomplete="off"
					spellcheck={false}
				/>
			</Field>

			<div class="max-h-80 overflow-y-auto rounded-fc-md bg-fc-surface">
				{#each shown as option (option.id)}
					<button
						type="button"
						class="flex w-full flex-col items-start gap-0.5 px-3 py-2 text-left hover:bg-fc-component"
						class:bg-fc-component={option.id === model}
						onclick={() => (model = option.id)}
					>
						<span class="font-mono text-fc-sm text-fc-fg">{option.id}</span>
						{#if option.name !== option.id}
							<span class="text-fc-xs text-fc-fg-muted">{option.name}</span>
						{/if}
					</button>
				{:else}
					<p class="px-3 py-4 text-fc-sm text-fc-fg-muted">
						{t('models.noMatch', { query: query.trim() })}
					</p>
				{/each}
			</div>
			{#if matches.length > shown.length}
				<p class="text-fc-xs text-fc-fg-muted">
					{t('models.showing', { shown: shown.length, total: matches.length })}
				</p>
			{/if}

			<div class="flex flex-wrap gap-2">
				<Button variant="secondary" disabled={saving || model === ''} onclick={() => (model = '')}>
					{t('models.useDefault')}
				</Button>
				<Button variant="primary" disabled={saving || model === saved} onclick={save}>
					{#if saving}<Spinner size="sm" />{/if}
					{t('models.save')}
				</Button>
			</div>
		</div>
	{/if}
</Section>
