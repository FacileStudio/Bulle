<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { Alert, Button, Card, Stack, icons } from '@facile/muse';
	import Logo from '$lib/components/Logo.svelte';
	import { ssoLoginUrl } from '$lib/backend';
	import { safeNext } from '$lib/nav';
	import { loadSession, session } from '$lib/session.svelte';

	let error = $state('');

	onMount(async () => {
		const params = new URLSearchParams(location.search);
		error = params.get('error') ?? '';
		await loadSession();
		if (session.status === 'signed-in') await goto(safeNext(params.get('next')));
	});
</script>

<svelte:head>
	<title>Sign in - Bulle</title>
</svelte:head>

<div class="flex min-h-dvh items-center justify-center bg-fc-page px-4 py-8">
	<Card class="w-full max-w-sm">
		<Stack gap="section">
			<Stack gap="tight">
				<a href="/" class="flex items-center gap-3 text-fc-fg" aria-label="Bulle home">
					<Logo size={28} />
					<span class="font-fc-title text-fc-xl font-semibold">Bulle</span>
				</a>
				<p class="text-fc-sm text-fc-fg-muted">Sign in with your Facile account to continue.</p>
			</Stack>

			{#if error}
				<Alert tone="danger">{error}</Alert>
			{/if}

			<Button href={ssoLoginUrl} data-sveltekit-reload icon={icons.key} size="lg" class="w-full">
				Sign in with Facile
			</Button>
		</Stack>
	</Card>
</div>
