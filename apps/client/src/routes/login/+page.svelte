<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import Logo from '$lib/components/Logo.svelte';
	import { ssoLoginUrl } from '$lib/backend';
	import DesktopLogin from '$lib/components/native/DesktopLogin.svelte';
	import NativeLogin from '$lib/components/native/NativeLogin.svelte';
	import { isDesktop, isNative } from '$lib/platform';
	import { safeNext } from '$lib/nav';
	import { loadSession, session } from '$lib/session.svelte';

	const primaryButtonClass =
		'inline-flex h-11 w-full items-center justify-center rounded-fc-md bg-fc-accent px-4 text-fc-sm font-medium text-fc-accent-fg transition-opacity hover:opacity-90 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring';

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
	<title>Log in - Bulle</title>
</svelte:head>

{#if desktop}
	<DesktopLogin bind:error />
{:else if native}
	<NativeLogin bind:error />
{:else}
<div class="flex min-h-dvh">
	<div class="hidden flex-col bg-black px-12 py-10 lg:flex lg:w-1/2">
		<a href="/" class="mb-auto flex items-center gap-3 text-white" aria-label="Bulle home">
			<Logo size={24} />
			<span class="font-fc-title text-fc-xl font-semibold tracking-tight">Bulle</span>
		</a>

		<div class="mb-auto">
			<h2 class="font-fc-title text-fc-3xl leading-tight font-semibold tracking-tight text-white">
				Your assistant.<br />Your suite.
			</h2>
			<p class="mt-4 max-w-xs text-fc-sm leading-relaxed text-white/50">
				A self-hosted AI assistant that works across your Facile apps.
			</p>
		</div>

		<p class="text-fc-xs text-white/30">
			© {new Date().getFullYear()} Bulle by Facile.
		</p>
	</div>

	<div class="flex w-full flex-col items-center justify-center bg-fc-page px-8 py-12 lg:w-1/2">
		<div class="w-full max-w-sm">
			<div class="mb-8">
				<h1 class="font-fc-title text-fc-2xl font-semibold tracking-tight text-fc-fg">
					Welcome back
				</h1>
				<p class="mt-1.5 text-fc-sm text-fc-fg-muted">
					Sign in with your Facile account to access Bulle.
				</p>
			</div>

			{#if error}
				<p role="alert" class="mb-4 rounded-fc-md bg-fc-danger/10 px-3 py-2 text-fc-sm text-fc-danger">
					{error}
				</p>
			{/if}

			<a href={ssoLoginUrl} data-sveltekit-reload class={primaryButtonClass}>
				Sign in with Facile
			</a>

			<a
				href="/"
				class="mt-6 inline-flex min-h-11 items-center text-fc-sm text-fc-fg-muted transition-colors hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
			>
				&larr; Back to home
			</a>
		</div>
	</div>
</div>
{/if}
