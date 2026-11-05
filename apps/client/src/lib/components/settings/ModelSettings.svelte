<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, Button, Field, Section, Select, Spinner, toast } from '@facile/muse';
	import { getPreferences, updatePreferences } from '$lib/preferences';

	let models = $state<string[]>([]);
	let model = $state('');
	let instanceModel = $state('');
	let loading = $state(true);
	let saving = $state(false);
	let loadError = $state<string | null>(null);

	onMount(async () => {
		try {
			const prefs = await getPreferences();
			models = prefs.models;
			model = prefs.model;
			instanceModel = prefs.default_model;
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Could not load the model settings';
		} finally {
			loading = false;
		}
	});

	// A model saved before the operator narrowed the list still shows, so the
	// picker never silently claims another model than the one in effect.
	const options = $derived(model && !models.includes(model) ? [model, ...models] : models);

	async function save() {
		saving = true;
		try {
			const saved = await updatePreferences({ model });
			model = saved.model;
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
	description="The model a new session starts on, here and in the CLI. Changing it does not touch sessions already running."
>
	{#if loading}
		<div class="flex justify-center py-6"><Spinner /></div>
	{:else if loadError}
		<Alert tone="danger">{loadError}</Alert>
	{:else}
		<div class="flex flex-col gap-4">
			<Field
				label="Model"
				helper={instanceModel
					? `The server's own default is ${instanceModel}.`
					: 'The server picks the model when this is left empty.'}
			>
				<Select bind:value={model}>
					<option value="">Server default</option>
					{#each options as option (option)}
						<option value={option}>{option}</option>
					{/each}
				</Select>
			</Field>
			<div>
				<Button variant="primary" disabled={saving} onclick={save}>
					{#if saving}<Spinner size="sm" />{/if}
					Save model
				</Button>
			</div>
		</div>
	{/if}
</Section>
