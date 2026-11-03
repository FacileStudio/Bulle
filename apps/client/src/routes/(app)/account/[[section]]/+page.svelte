<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		Alert,
		Button,
		PageHeader,
		ProfileCard,
		Section,
		Spinner,
		Stack,
		Tabs,
		icons
	} from '@facile/muse';
	import { Blob } from '@facile/blob';
	import ComputerSettings from '$lib/components/settings/ComputerSettings.svelte';
	import NotificationSettings from '$lib/components/settings/NotificationSettings.svelte';
	import ProviderKeys from '$lib/components/settings/ProviderKeys.svelte';
	import UsageSettings from '$lib/components/settings/UsageSettings.svelte';
	import SpaceSettings from '$lib/components/settings/SpaceSettings.svelte';
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

	const sections = $derived([
		{ id: 'profile', label: 'Profile', icon: icons.userCircle },
		{ id: 'assistant', label: 'Assistant', icon: icons.bolt },
		{ id: 'spaces', label: 'Spaces', icon: icons.usersGroup },
		{ id: 'notifications', label: 'Notifications', icon: icons.notification },
		{ id: 'usage', label: 'Usage', icon: icons.card },
		{ id: 'providers', label: 'Providers', icon: icons.key },
		{ id: 'computer', label: 'Computer', icon: icons.monitor },
		{ id: 'connections', label: 'Apps', icon: icons.plug }
	]);
	const tabs = $derived(sections.map((s) => ({ ...s, href: `/account/${s.id}` })));
	const section = $derived(
		sections.some((s) => s.id === page.params.section) ? page.params.section : 'profile'
	);

	async function leave() {
		await signOut();
		await goto('/login', { replaceState: true });
	}
</script>

<svelte:head>
	<title>Settings - Bulle</title>
</svelte:head>

<Stack gap="section">
	<PageHeader title="Settings" />
	<Tabs items={tabs} value={section} label="Settings sections" />
	{#if session.user && section === 'profile'}
		<ProfileCard name={session.user.name.trim() || session.user.email} email={session.user.email}>
			{#snippet actions()}
				<Button variant="outline" icon={icons.logout} onclick={leave}>Sign out</Button>
			{/snippet}
		</ProfileCard>
	{:else if section === 'assistant'}
		<Section card title="Assistant Profile" description="Customize your assistant persona, avatar, and traits.">
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

	{:else if section === 'spaces'}
		<SpaceSettings />
	{:else if section === 'notifications'}
		<NotificationSettings />
	{:else if section === 'usage'}
		<UsageSettings />
	{:else if section === 'providers'}
		<ProviderKeys />
	{:else if section === 'computer'}
		<ComputerSettings />
	{:else if section === 'connections'}
		<SuiteConnections />
	{/if}
</Stack>
