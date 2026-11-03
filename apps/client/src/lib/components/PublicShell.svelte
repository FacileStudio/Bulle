<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Button } from '@facile/muse';
	import Logo from '$lib/components/Logo.svelte';

	let { children }: { children: Snippet } = $props();
</script>

<div class="flex min-h-dvh flex-col bg-fc-page text-fc-fg">
	<header>
		<div class="mx-auto flex max-w-5xl items-center justify-between gap-4 px-6 py-4">
			<a href="/" class="flex items-center gap-2.5">
				<Logo size={28} />
				<span class="font-fc-title text-fc-xl font-semibold tracking-tight text-fc-fg">Bulle</span>
			</a>
			<nav class="flex items-center gap-2" aria-label="Main">
				<Button href="/downloads" variant="ghost" size="sm" class="rounded-fc-md">Downloads</Button>
				<Button href="/login" size="sm" class="rounded-fc-md">Sign in</Button>
			</nav>
		</div>
	</header>

	<main class="flex-1">
		{@render children()}
	</main>

	<footer>
		<div
			class="mx-auto flex max-w-5xl flex-col items-center justify-between gap-2 px-6 py-6 text-fc-xs text-fc-fg-muted sm:flex-row"
		>
			<p>
				© {new Date().getFullYear()} Bulle by{' '}
				<a
					href="https://facile.studio"
					target="_blank"
					rel="noopener"
					class="font-medium underline underline-offset-2 transition-colors hover:text-fc-fg"
				>
					Facile
				</a>.
			</p>
			<p>Part of the Facile Suite.</p>
		</div>
	</footer>
</div>
