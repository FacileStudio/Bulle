<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Spinner } from '@facile/muse';
	import Onboarding from '$lib/components/onboarding/Onboarding.svelte';
	import { needsOnboarding } from '$lib/onboarding/steps';
	import { loadSession, session } from '$lib/session.svelte';

	let ready = $state(false);

	onMount(async () => {
		await loadSession();
		if (session.status === 'signed-out') {
			await goto('/login?next=%2Fwelcome', { replaceState: true });
			return;
		}
		if (session.status === 'signed-in' && !needsOnboarding(session.user)) {
			await goto('/sessions', { replaceState: true });
			return;
		}
		ready = true;
	});
</script>

<svelte:head>
	<title>{t('onboarding.pageTitle')}</title>
</svelte:head>

{#if ready && session.status === 'signed-in'}
	<Onboarding />
{:else if ready}
	<div class="flex h-dvh items-center justify-center bg-fc-page px-4 text-center text-fc-sm text-fc-fg-muted">
		{t('onboarding.unreachable')}
	</div>
{:else}
	<div class="flex h-dvh items-center justify-center bg-fc-page">
		<Spinner />
	</div>
{/if}
