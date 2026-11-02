<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import {
		Button,
		EmptyState,
		MobileNav,
		PageTransition,
		SideBar,
		Spinner,
		icons
	} from '@facile/muse';
	import { LOGO_ICON } from '$lib/icons';
	import { loadSession, session } from '$lib/session.svelte';

	let { children } = $props();

	let collapsed = $state(false);
	let scroller = $state<HTMLElement | null>(null);

	function isActive(href: string) {
		return page.url.pathname === href || page.url.pathname.startsWith(`${href}/`);
	}

	const chatRoute = $derived(
		page.url.pathname === '/sessions' || page.url.pathname.startsWith('/s/')
	);

	const navPages = $derived([
		{
			href: '/sessions',
			label: 'Sessions',
			icon: icons.history,
			active: chatRoute
		},
		{
			href: '/computer',
			label: 'Computer',
			icon: icons.monitor,
			active: isActive('/computer')
		}
	]);

	const navUser = $derived(
		session.user ? { name: session.user.name.trim() || session.user.email } : undefined
	);

	onMount(async () => {
		await loadSession();
		if (session.status === 'signed-out') {
			const next = page.url.pathname + page.url.search;
			await goto(`/login?next=${encodeURIComponent(next)}`, { replaceState: true });
		}
	});

	$effect(() => {
		void page.url.pathname;
		scroller?.scrollTo({ top: 0 });
	});
</script>

<div class="flex h-dvh w-full overflow-hidden bg-fc-page">
	<div class="hidden h-full shrink-0 p-3 md:block">
		<SideBar
			class="h-full"
			icon={LOGO_ICON}
			title="Bulle"
			bind:collapsed
			pages={navPages}
			user={navUser}
			userHref="/account"
			userActive={isActive('/account')}
		/>
	</div>
	<main
		bind:this={scroller}
		class="min-w-0 flex-1 overscroll-contain pb-24 md:pb-0 {chatRoute && session.status === 'signed-in'
			? 'overflow-hidden'
			: 'overflow-auto'}"
	>
		<div
			class={chatRoute && session.status === 'signed-in'
				? 'h-full'
				: 'mx-auto flex w-full max-w-fc-xl flex-col gap-10 px-4 py-8 sm:px-6 md:px-10 md:py-10'}
		>
			{#if session.status === 'signed-in' && chatRoute}
				{@render children()}
			{:else if session.status === 'signed-in'}
				<PageTransition key={page.url.pathname}>
					{@render children()}
				</PageTransition>
			{:else if session.status === 'unreachable'}
				<EmptyState
					icon={icons.error}
					title="Bulle is not answering"
					description="The server did not return your account. You are still signed in."
				>
					<Button icon={icons.refresh} onclick={() => location.reload()}>Try again</Button>
				</EmptyState>
			{:else}
				<div class="flex min-h-[50dvh] items-center justify-center">
					<Spinner />
				</div>
			{/if}
		</div>
	</main>
	<MobileNav
		items={navPages}
		user={navUser}
		profileHref="/account"
		profileActive={isActive('/account')}
	/>
</div>
