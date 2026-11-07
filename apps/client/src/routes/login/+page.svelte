<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Alert, AuthScreen } from '@facile/muse';
	import Logo from '$lib/components/Logo.svelte';
	import { ssoLoginUrl } from '$lib/backend';
	import DesktopLogin from '$lib/components/native/DesktopLogin.svelte';
	import NativeLogin from '$lib/components/native/NativeLogin.svelte';
	import { rememberLocale, t } from '$lib/i18n/index.svelte';
	import { isDesktop, isNative } from '$lib/platform';
	import { safeNext } from '$lib/nav';
	import { loadSession, session } from '$lib/session.svelte';

	let error = $state('');
	const native = isNative();
	const desktop = isDesktop();

	$effect(() => {
		error = page.url.searchParams.get('error') ?? '';
	});

	onMount(async () => {
		await loadSession();
		if (session.status === 'signed-in') await goto(safeNext(page.url.searchParams.get('next')));
	});
</script>

<svelte:head>
	<title>{t('login.title')}</title>
</svelte:head>

{#if desktop}
	<DesktopLogin bind:error />
{:else if native}
	<NativeLogin bind:error />
{:else}
	<AuthScreen appName="Bulle" href={ssoLoginUrl} onLocaleChange={rememberLocale}>
		{#snippet logo()}
			<Logo size={40} />
		{/snippet}
		{#snippet tagline()}
			{t('login.tagline')}
		{/snippet}
		{#snippet notice()}
			{#if error}
				<Alert tone="danger">{error}</Alert>
			{/if}
		{/snippet}
	</AuthScreen>
{/if}
