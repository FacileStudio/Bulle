<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import {
		NotificationsSection,
		PreferencesSection,
		ProfileSection,
		SettingsLayout,
		SubscriptionSection,
		icons,
		settingsRoute
	} from '@facile/muse';
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
	import { isApp } from '$lib/platform';
	import { rememberLocale } from '$lib/i18n/index.svelte';
	import { session } from '$lib/session.svelte';
	import { chooseTheme, storedTheme } from '$lib/theme';
	import { planLabel, SUBSCRIBE_URL } from '$lib/plan';
	import { t } from '$lib/i18n/index.svelte';

	const plan = $derived(planLabel(session.user?.plan));

	const sections = $derived(
		[
			{ id: 'profile', label: t('account.tab.profile'), icon: icons.userCircle },
			{ id: 'preferences', label: t('account.tab.preferences'), icon: icons.settings },
			{ id: 'notifications', label: t('account.tab.notifications'), icon: icons.notification },
			{ id: 'model', label: t('account.tab.model'), icon: icons.compass },
			{ id: 'apps', label: t('account.tab.apps'), icon: icons.plug },
			{ id: 'jobs', label: t('account.tab.jobs'), icon: icons.clock },
			{ id: 'devices', label: t('account.tab.devices'), icon: icons.monitor },
			...(plan ? [{ id: 'subscription', label: t('account.tab.subscription'), icon: icons.card }] : []),
			...(isApp() ? [{ id: 'updates', label: t('account.tab.updates'), icon: icons.refresh }] : [])
		].map((s) => ({ ...s, href: `/account/${s.id}` }))
	);

	const ALIASES: Record<string, string> = {
		appearance: 'preferences',
		assistant: 'profile',
		providers: 'model',
		usage: 'model',
		spaces: 'apps',
		matrix: 'apps',
		connections: 'apps',
		computer: 'devices'
	};

	let theme = $state(storedTheme());

	$effect(() => {
		const requested = page.params.section;
		if (!requested) return;
		const alias = ALIASES[requested];
		if (alias) void goto(`/account/${alias}`, { replaceState: true });
		else if (!settingsRoute(requested, sections).known) void goto('/account', { replaceState: true });
	});
</script>

<svelte:head>
	<title>{t('account.pageTitle')}</title>
</svelte:head>

<SettingsLayout
	{sections}
	active={page.params.section}
	backHref="/account"
	title={t('account.heading')}
	navLabel={t('account.tabsLabel')}
>
	{#snippet children(section)}
		{#if section === 'profile'}
			{#if session.user}
				<ProfileSection
					name={session.user.name.trim() || session.user.email}
					email={session.user.email}
					avatar={session.user.avatar_url}
				/>
			{/if}
			<AssistantSettings />
		{:else if section === 'preferences'}
			<PreferencesSection
				{theme}
				onThemeChange={(next) => {
					theme = next;
					chooseTheme(next);
				}}
				onLocaleChange={rememberLocale}
			/>
		{:else if section === 'notifications'}
			<NotificationSettings />
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
		{:else if section === 'subscription' && plan}
			<SubscriptionSection
				{plan}
				manageHref={SUBSCRIBE_URL}
				upgradeHref={session.user?.plan === 'free' ? SUBSCRIBE_URL : undefined}
			/>
		{:else if section === 'updates'}
			<UpdateSettings />
		{/if}
	{/snippet}
</SettingsLayout>
