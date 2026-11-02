<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import {
		Button,
		EmptyState,
		Icon,
		IconButton,
		PageTransition,
		Spinner,
		icons
	} from '@facile/muse';
	import AppSidebar from '$lib/components/AppSidebar.svelte';
	import MobileSidebar from '$lib/components/MobileSidebar.svelte';
	import { sessions } from '$lib/chat/sessions.svelte';
	import { anyRunning, jobsActivity, watchJobs } from '$lib/jobs-activity.svelte';
	import { CHAT_ICONS, registerAgentAvatarIcon } from '$lib/icons';
	import { mobileSidebar } from '$lib/mobile-sidebar.svelte';
	import { needsOnboarding } from '$lib/onboarding/steps';
	import { loadSession, session } from '$lib/session.svelte';

	let { children } = $props();

	let scroller = $state<HTMLElement | null>(null);

	function isActive(href: string) {
		return page.url.pathname === href || page.url.pathname.startsWith(`${href}/`);
	}

	const ready = $derived(session.status === 'signed-in' && !needsOnboarding(session.user));
	const chatRoute = $derived(
		page.url.pathname === '/sessions' || page.url.pathname.startsWith('/s/')
	);

	const assistantName = $derived(session.user?.assistant?.name?.trim() || 'Bulle');
	const assistantIcon = $derived(registerAgentAvatarIcon(assistantName));

	const navPages = $derived([
		{
			href: '/sessions',
			label: assistantName,
			icon: assistantIcon,
			active: chatRoute
		},
		{
			href: '/jobs',
			label: 'Jobs',
			icon: icons.clock,
			active: isActive('/jobs'),
			dot: anyRunning(jobsActivity.jobs)
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
			return;
		}
		if (session.status === 'signed-in' && needsOnboarding(session.user)) {
			await goto('/welcome', { replaceState: true });
		}
	});

	$effect(() => {
		if (ready) void sessions.refresh();
	});

	$effect(() => {
		if (ready) return watchJobs();
	});

	$effect(() => {
		void page.url.pathname;
		scroller?.scrollTo({ top: 0 });
	});
</script>

<div class="flex h-dvh w-full overflow-hidden bg-fc-page">
	{#if ready}
		<div class="hidden h-full w-72 shrink-0 p-3 md:block">
			<div class="h-full overflow-hidden rounded-fc-lg bg-fc-component">
				<AppSidebar
					items={navPages}
					userName={navUser?.name}
					userActive={isActive('/account')}
				/>
			</div>
		</div>
	{/if}
	<main
		bind:this={scroller}
		class="min-w-0 flex-1 overscroll-contain {chatRoute && ready
			? 'overflow-hidden'
			: 'overflow-auto'}"
	>
		{#if !(chatRoute && ready)}
			<header class="flex h-14 items-center gap-2 px-2 md:hidden">
				<IconButton
					variant="ghost"
					aria-label="Open the sidebar"
					onclick={() => (mobileSidebar.open = true)}
				>
					<Icon icon={CHAT_ICONS.menu} size={18} />
				</IconButton>
				<span class="font-fc-title text-fc-md font-semibold tracking-tight">Bulle</span>
			</header>
		{/if}
		<div
			class={chatRoute && ready
				? 'h-full'
				: 'mx-auto flex w-full max-w-fc-xl flex-col gap-10 px-4 py-8 sm:px-6 md:px-10 md:py-10'}
		>
			{#if ready && chatRoute}
				{@render children()}
			{:else if ready}
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
	{#if ready}
		<MobileSidebar items={navPages} userName={navUser?.name} userActive={isActive('/account')} />
	{/if}
</div>
