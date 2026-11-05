<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, Button, Field, Input, Section, Spinner, toast } from '@facile/muse';
	import { getPreferences, updatePreferences, type ModelOption } from '$lib/preferences';

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
			loadError = err instanceof Error ? err.message : 'Could not load the model settings';
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
			toast.success('Default model saved. New sessions start on it.');
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not save the default model');
		} finally {
			saving = false;
		}
	}
</script>

<Section
	card
	id="model"
	title="Default model"
	description="The model a new session starts on, here and in the CLI. The list is what the current provider offers; the CLI's /model browses the same catalogue."
>
	{#if loading}
		<div class="flex justify-center py-6"><Spinner /></div>
	{:else if loadError}
		<Alert tone="danger">{loadError}</Alert>
	{:else}
		<div class="flex flex-col gap-4">
			<div class="text-fc-sm text-fc-fg">
				Current:
				<span class="font-mono">{model || `server default (${instanceModel})`}</span>
				{#if current && current.name !== current.id}
					<span class="text-fc-fg-muted"> — {current.name}</span>
				{/if}
			</div>

			<Field label="Search models" helper={`${catalog.length} models from this provider`}>
				<Input
					bind:value={query}
					placeholder="Search by name or id, e.g. mimo, claude, gpt"
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
						No model matches “{query.trim()}”.
					</p>
				{/each}
			</div>
			{#if matches.length > shown.length}
				<p class="text-fc-xs text-fc-fg-muted">
					Showing {shown.length} of {matches.length}. Narrow the search to see the rest.
				</p>
			{/if}

			<div class="flex flex-wrap gap-2">
				<Button variant="outline" disabled={saving || model === ''} onclick={() => (model = '')}>
					Use the server default
				</Button>
				<Button variant="primary" disabled={saving || model === saved} onclick={save}>
					{#if saving}<Spinner size="sm" />{/if}
					Save model
				</Button>
			</div>
		</div>
	{/if}
</Section>
