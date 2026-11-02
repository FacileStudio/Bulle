<script lang="ts">
	import { page } from '$app/state';
	import { Divider, NavButton } from '@facile/muse';
	import Logo from '$lib/components/Logo.svelte';
	import SessionList from '$lib/components/SessionList.svelte';
	import { listenForSwipes, mobileSidebar } from '$lib/mobile-sidebar.svelte';

	type Item = { href: string; label: string; icon?: string; active?: boolean };

	let {
		items,
		userName,
		userActive = false
	}: { items: Item[]; userName?: string; userActive?: boolean } = $props();

	const close = () => (mobileSidebar.open = false);

	$effect(() => listenForSwipes());

	$effect(() => {
		void page.url.pathname;
		close();
	});

	$effect(() => {
		if (!mobileSidebar.open) return;
		const onkey = (event: KeyboardEvent) => event.key === 'Escape' && close();
		window.addEventListener('keydown', onkey);
		return () => window.removeEventListener('keydown', onkey);
	});
</script>

<div class="fixed inset-0 z-40 md:hidden {mobileSidebar.open ? '' : 'pointer-events-none'}">
	<button
		type="button"
		tabindex="-1"
		aria-label="Close the sidebar"
		class="absolute inset-0 bg-black/40 transition-opacity duration-300 {mobileSidebar.open
			? 'opacity-100'
			: 'opacity-0'}"
		onclick={close}
	></button>
	<aside
		inert={!mobileSidebar.open}
		aria-label="Navigation"
		class="absolute inset-y-0 left-0 flex w-[85%] max-w-80 flex-col border-r border-fc-border bg-fc-page shadow-xl transition-transform duration-300 ease-out {mobileSidebar.open
			? 'translate-x-0'
			: '-translate-x-full'}"
	>
		<div class="flex h-14 shrink-0 items-center gap-2 px-4">
			<Logo size={22} />
			<span class="font-fc-title text-fc-md font-semibold tracking-tight">Bulle</span>
		</div>
		<nav class="flex flex-col gap-1 px-3" aria-label="Pages">
			{#each items as item (item.href)}
				<NavButton href={item.href} icon={item.icon} label={item.label} active={item.active} />
			{/each}
		</nav>
		<Divider class="mx-3 mt-3" />
		<div class="min-h-0 flex-1">
			<SessionList onnavigate={close} />
		</div>
		{#if userName}
			<div class="border-t border-fc-border p-3">
				<NavButton href="/account" label={userName} active={userActive} />
			</div>
		{/if}
	</aside>
</div>
