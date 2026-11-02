<script lang="ts">
	import { Alert, Button, Spinner, icons } from '@facile/muse';
	import Logo from '$lib/components/Logo.svelte';
	import OnboardingAvatar from '$lib/components/onboarding/OnboardingAvatar.svelte';
	import { startNativeLogin } from '$lib/native';

	let { error = $bindable('') }: { error?: string } = $props();

	let opening = $state(false);

	async function signIn() {
		if (opening) return;
		error = '';
		opening = true;
		try {
			await startNativeLogin();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Could not open the sign-in page.';
		} finally {
			opening = false;
		}
	}
</script>

<div class="fixed inset-0 flex flex-col overflow-hidden bg-fc-page text-fc-fg">

	<header class="relative z-10 flex h-14 items-center gap-2 px-5">
		<Logo size={22} />
		<span class="font-fc-title text-fc-md font-semibold tracking-tight">Bulle</span>
	</header>

	<main
		class="relative z-10 flex min-h-0 flex-1 flex-col items-center justify-center gap-8 overflow-y-auto px-8 text-center"
	>
		<OnboardingAvatar name="Pixel" size={136} />
		<div class="flex max-w-xs flex-col gap-3">
			<h1 class="font-fc-title text-fc-3xl leading-tight font-semibold tracking-tight">
				Sign in to Bulle
			</h1>
			<p class="text-fc-md leading-relaxed text-fc-fg-muted">
				Use your Facile account. Your sessions, jobs and approvals follow you from the web.
			</p>
		</div>
	</main>

	<footer class="relative z-10 flex flex-col items-center gap-4 px-6 pt-4 pb-6">
		{#if error}
			<div class="w-full max-w-sm" role="alert">
				<Alert tone="danger">{error}</Alert>
			</div>
		{/if}
		<Button
			size="lg"
			class="w-full max-w-sm rounded-fc-md"
			icon={opening ? undefined : icons.key}
			disabled={opening}
			onclick={signIn}
		>
			{#if opening}
				<Spinner size="sm" />
			{/if}
			Continue with Facile
		</Button>
		<p class="max-w-xs text-fc-xs text-fc-fg-muted">
			Sign-in opens in your browser, then brings you back here.
		</p>
	</footer>
</div>
