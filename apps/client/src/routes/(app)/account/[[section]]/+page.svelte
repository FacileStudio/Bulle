<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Button, PageHeader, ProfileCard, Stack, Tabs, icons } from '@facile/muse';
	import AssistantSettings from '$lib/components/settings/AssistantSettings.svelte';
	import ComputerSettings from '$lib/components/settings/ComputerSettings.svelte';
	import HardwareSettings from '$lib/components/settings/HardwareSettings.svelte';
	import JobsSettings from '$lib/components/settings/JobsSettings.svelte';
	import ModelSettings from '$lib/components/settings/ModelSettings.svelte';
	import NotificationSettings from '$lib/components/settings/NotificationSettings.svelte';
	import MatrixSettings from '$lib/components/settings/MatrixSettings.svelte';
	import ProviderKeys from '$lib/components/settings/ProviderKeys.svelte';
	import UsageSettings from '$lib/components/settings/UsageSettings.svelte';
	import SpaceSettings from '$lib/components/settings/SpaceSettings.svelte';
	import SuiteConnections from '$lib/components/SuiteConnections.svelte';
	import UpdateSettings from '$lib/components/settings/UpdateSettings.svelte';
	import { isDesktop } from '$lib/platform';
	import { session, signOut } from '$lib/session.svelte';

	/*
	 * Five places to go, not twelve. Related settings share a tab, so nobody has to remember
	 * whether the model picker sat under "Model" and the keys under "Providers", or that the
	 * assistant was its own page while the account was another.
	 */
	const sections = $derived([
		{ id: 'profile', label: 'Profile', icon: icons.userCircle },
		{ id: 'model', label: 'Model', icon: icons.compass },
		{ id: 'apps', label: 'Apps', icon: icons.plug },
		{ id: 'jobs', label: 'Jobs', icon: icons.clock },
		{ id: 'devices', label: 'Devices', icon: icons.monitor },
		{ id: 'notifications', label: 'Notifications', icon: icons.notification },
		...(isDesktop() ? [{ id: 'updates', label: 'Updates', icon: icons.refresh }] : [])
	]);
	const tabs = $derived(sections.map((s) => ({ ...s, href: `/account/${s.id}` })));

	/* Old links keep working: a bookmark to /account/providers lands on Model. */
	const ALIASES: Record<string, string> = {
		assistant: 'profile',
		providers: 'model',
		usage: 'model',
		spaces: 'apps',
		matrix: 'apps',
		connections: 'apps',
		computer: 'devices'
	};
	const section = $derived(
		sections.some((s) => s.id === page.params.section)
			? (page.params.section as string)
			: (ALIASES[page.params.section ?? ''] ?? 'profile')
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
	{#if !session.user}
		<!-- The profile card needs the signed-in user; nothing else on this page does. -->
	{:else if section === 'profile'}
		<ProfileCard
			name={session.user.name.trim() || session.user.email}
			email={session.user.email}
			avatar={session.user.avatar_url}
		>
			{#snippet actions()}
				<Button variant="ghost" icon={icons.logout} onclick={leave}>Sign out</Button>
			{/snippet}
		</ProfileCard>
		<AssistantSettings />
	{:else if section === 'model'}
		<ModelSettings />
		<ProviderKeys />
		<UsageSettings />
	{:else if section === 'apps'}
		<SpaceSettings />
		<MatrixSettings />
		<SuiteConnections />
	{:else if section === 'jobs'}
		<JobsSettings />
	{:else if section === 'devices'}
		<ComputerSettings />
		<HardwareSettings />
	{:else if section === 'notifications'}
		<NotificationSettings />
	{:else if section === 'updates'}
		<UpdateSettings />
	{/if}
</Stack>
