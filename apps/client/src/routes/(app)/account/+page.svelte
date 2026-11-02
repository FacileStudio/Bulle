<script lang="ts">
	import { goto } from '$app/navigation';
	import {
		Alert,
		Button,
		PageHeader,
		ProfileCard,
		Section,
		Spinner,
		Stack,
		icons
	} from '@facile/muse';
	import { avatarUrlFor } from '$lib/assistant';
	import SuiteConnections from '$lib/components/SuiteConnections.svelte';
	import { updateAssistant } from '$lib/backend';
	import ArchetypeCards from '$lib/components/onboarding/ArchetypeCards.svelte';
	import NameField from '$lib/components/onboarding/NameField.svelte';
	import TraitChips from '$lib/components/onboarding/TraitChips.svelte';
	import { initialDraft, profileFromDraft } from '$lib/onboarding/profile';
	import { session, setAssistant, signOut } from '$lib/session.svelte';

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

	async function saveAssistant() {
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

	async function leave() {
		await signOut();
		await goto('/login', { replaceState: true });
	}
</script>

<svelte:head>
	<title>Account - Bulle</title>
</svelte:head>

<Stack gap="section">
	<PageHeader title="Account" description="Your identity comes from Facile." />
	{#if session.user}
		<ProfileCard name={session.user.name.trim() || session.user.email} email={session.user.email}>
			{#snippet actions()}
				<Button variant="outline" icon={icons.logout} onclick={leave}>Sign out</Button>
			{/snippet}
		</ProfileCard>

		<Section card title="Assistant Profile" description="Customize your assistant persona, avatar, and traits.">
			<div class="flex flex-col gap-5">
				{#if saveSuccess}
					<Alert tone="success">Assistant profile updated.</Alert>
				{/if}
				{#if saveError}
					<Alert tone="danger">{saveError}</Alert>
				{/if}

				<div class="flex items-center gap-4">
					<img
						src={avatarUrlFor(name)}
						alt=""
						aria-hidden="true"
						class="size-16 rounded-full object-cover"
					/>
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
					<Button
						variant="primary"
						disabled={saving}
						onclick={saveAssistant}
					>
						{#if saving}
							<Spinner size="sm" />
						{/if}
						Save Changes
					</Button>
				</div>
			</div>
		</Section>

		<SuiteConnections />
	{/if}
</Stack>
