<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, Button, Field, Section, Spinner, Textarea, toast } from '@facile/muse';
	import { getPreferences, updatePreferences } from '$lib/preferences';

	let config = $state('');
	let loading = $state(true);
	let saving = $state(false);
	let loadError = $state<string | null>(null);

	onMount(async () => {
		try {
			config = (await getPreferences()).cli_config;
		} catch (err) {
			loadError = err instanceof Error ? err.message : 'Could not load the CLI settings';
		} finally {
			loading = false;
		}
	});

	async function save() {
		saving = true;
		try {
			const saved = await updatePreferences({ cli_config: config });
			config = saved.cli_config;
			toast.success('CLI settings saved. The next bulle start reads them.');
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not save the CLI settings');
		} finally {
			saving = false;
		}
	}
</script>

<Section
	card
	id="cli"
	title="CLI settings"
	description="A config.yml every signed-in machine reads. It sits above the built-in defaults and below that machine's own ~/.config/bulle/config.yml, so a local file still wins. Only the keys the CLI already reads are accepted: no provider settings, and no server address."
>
	{#if loading}
		<div class="flex justify-center py-6"><Spinner /></div>
	{:else if loadError}
		<Alert tone="danger">{loadError}</Alert>
	{:else}
		<form
			class="flex flex-col gap-4"
			onsubmit={(e) => {
				e.preventDefault();
				void save();
			}}
		>
			<Field
				label="config.yml"
				helper="Same keys and groups as the file, validated on save: an unknown key or a server: block is refused."
			>
				<Textarea
					bind:value={config}
					rows={16}
					spellcheck={false}
					class="font-mono text-fc-sm"
					placeholder={'limits:\n  max_concurrency: 16\n\nsession:\n  root: .'}
				/>
			</Field>
			<div>
				<Button type="submit" variant="primary" disabled={saving}>
					{#if saving}<Spinner size="sm" />{/if}
					Save settings
				</Button>
			</div>
		</form>
	{/if}
</Section>
