<script lang="ts">
	import { Alert, Button, Section, Spinner } from '@facile/muse';
	import { Blob } from '@facile/blob';
	import { updateAssistant } from '$lib/backend';
	import ArchetypeCards from '$lib/components/onboarding/ArchetypeCards.svelte';
	import NameField from '$lib/components/onboarding/NameField.svelte';
	import TraitChips from '$lib/components/onboarding/TraitChips.svelte';
	import { initialDraft, profileFromDraft } from '$lib/onboarding/profile';
	import { session, setAssistant } from '$lib/session.svelte';

	const start = initialDraft(session.user?.assistant);
	let name = $state(start.name);
	let archetype = $state(start.archetype);
	let traits = $state<string[]>(start.traits);
	let saving = $state(false);
	let saveSuccess = $state(false);
	let saveError = $state<string | null>(null);

	$effect(() => {
		if (session.user?.assistant && !name && session.user.assistant.name) {
			const loaded = initialDraft(session.user.assistant);
			name = loaded.name;
			archetype = loaded.archetype;
			traits = loaded.traits;
		}
	});

	async function save() {
		saving = true;
		saveError = null;
		saveSuccess = false;
		try {
			const updated = await updateAssistant(profileFromDraft({ step: 0, name, archetype, traits }));
			setAssistant(updated);
			saveSuccess = true;
		} catch (err) {
			saveError = err instanceof Error ? err.message : 'Could not save assistant profile';
		} finally {
			saving = false;
		}
	}
</script>

<Section
	card
	id="assistant"
	title="Assistant"
	description="How your assistant looks, sounds and behaves, here and in the CLI."
>
	<div class="flex flex-col gap-5">
		{#if saveSuccess}
			<Alert tone="success">Assistant profile updated.</Alert>
		{/if}
		{#if saveError}
			<Alert tone="danger">{saveError}</Alert>
		{/if}

		<div class="flex items-center gap-4">
			<Blob name={name.trim() || 'bulle'} size={64} class="size-16 shrink-0" />
			<div class="min-w-0 flex-1">
				<NameField id="account-assistant-name" bind:value={name} />
			</div>
		</div>

		<div class="flex flex-col gap-2">
			<span class="text-fc-sm font-medium text-fc-fg">Personality</span>
			<ArchetypeCards bind:archetype {traits} />
		</div>

		<div class="flex flex-col gap-2">
			<span class="text-fc-sm font-medium text-fc-fg">Traits</span>
			<TraitChips bind:traits />
		</div>

		<div>
			<Button variant="primary" disabled={saving} onclick={save}>
				{#if saving}
					<Spinner size="sm" />
				{/if}
				Save Changes
			</Button>
		</div>
	</div>
</Section>
