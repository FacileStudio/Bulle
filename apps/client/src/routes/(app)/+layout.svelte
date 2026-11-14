<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { onMount, untrack } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import {
		Button,
		EmptyState,
		Icon,
		IconButton,
		PageTransition,
		Spinner,
		icons,
		type UserMenuOptions
	} from '@facile/muse';
	import AppSidebar from '$lib/components/AppSidebar.svelte';
	import { ARTIFACT_ICON } from '$lib/suite-icons';
	import MobileSidebar from '$lib/components/MobileSidebar.svelte';
	import { sessions } from '$lib/chat/sessions.svelte';
	import { anyRunning, jobsActivity, refreshJobs, watchJobs } from '$lib/jobs-activity.svelte';
	import { ideasActivity, newIdeasCount, refreshIdeas } from '$lib/ideas-activity.svelte';
	import { spaces } from '$lib/spaces.svelte';
	import { backend } from '$lib/backend';
	import { syncDevice } from '$lib/notifications';
	import { startLive } from '$lib/live/sync';
	import { CHAT_ICONS, registerAgentAvatarIcon } from '$lib/icons';
	import { desktopSidebar, isSidebarShortcut, openSidebar, setCollapsed } from '$lib/mobile-sidebar.svelte';
	import { hasNativeTrafficLights, isDesktop, isNative } from '$lib/platform';
	import { rememberLocale } from '$lib/i18n/index.svelte';
	import { chooseTheme, storedTheme } from '$lib/theme';
	import { planLabel, SUBSCRIBE_URL } from '$lib/plan';
	import { needsOnboarding } from '$lib/onboarding/steps';
	import { assistantName, loadSession, session, signOut } from '$lib/session.svelte';

	let { children } = $props();

	let scroller = $state<HTMLElement | null>(null);

	function isActive(href: string) {
		return page.url.pathname === href || page.url.pathname.startsWith(`${href}/`);
	}

	const ready = $derived(session.status === 'signed-in' && !needsOnboarding(session.user));
	const chatRoute = $derived(
		page.url.pathname === '/sessions' || page.url.pathname.startsWith('/s/')
	);

	const name = $derived(assistantName());
	const assistantIcon = $derived(registerAgentAvatarIcon(name));

	const navPages = $derived([
		{
			href: '/sessions',
			label: name,
			icon: assistantIcon,
			active: chatRoute
		},
		{
			href: '/ideas',
			label: t('nav.ideas'),
			icon: 'solar:lightbulb-linear',
			active: isActive('/ideas'),
			running: ideasActivity.generating,
			count: newIdeasCount()
		},
		{
			href: '/artifacts',
			label: t('nav.artifacts'),
			icon: ARTIFACT_ICON,
			active: isActive('/artifacts')
		}
	]);

	const jobsRunning = $derived(anyRunning(jobsActivity.jobs));

	const navUser = $derived(
		session.user
			? {
					name: session.user.name.trim() || session.user.email,
					email: session.user.email,
					avatar: session.user.avatar_url
				}
			: undefined
	);

	async function leave() {
		await signOut();
		await goto('/login', { replaceState: true });
	}

	let theme = $state(storedTheme());

	const plan = $derived(planLabel(session.user?.plan));

	const menu = $derived<UserMenuOptions>({
		subscription: plan ? { plan, href: SUBSCRIBE_URL } : null,
		items: [
			{
				label: jobsRunning ? `${t('profile.jobs')} (${t('profile.jobRunning')})` : t('profile.jobs'),
				icon: icons.clock,
				href: '/account/jobs'
			},
			...(isNative()
				? []
				: [
						{
							label: isDesktop() ? t('profile.getApps') : t('profile.getDesktop'),
							icon: icons.download,
							href: '/install'
						}
					])
		],
		helpHref: 'https://facile.studio/docs/bulle',
		onSignOut: leave,
		theme,
		onThemeChange: (next) => {
			theme = next;
			chooseTheme(next);
		},
		onLocaleChange: rememberLocale
	});

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
		if (ready) void spaces.load();
	});

	$effect(() => {
		if (!ready || !spaces.loaded) return;
		void spaces.currentId;
		untrack(() => {
			void sessions.refresh();
			void refreshJobs();
			void refreshIdeas();
		});
	});

	$effect(() => {
		if (ready && spaces.loaded) return untrack(watchJobs);
	});

	$effect(() => {
		if (ready) return startLive();
	});

	$effect(() => {
		if (ready) void backend.getPush().then(syncDevice).catch(() => undefined);
	});

	$effect(() => {
		void page.url.pathname;
		scroller?.scrollTo({ top: 0 });
	});
</script>

<svelte:window
	onkeydown={(event) => {
		if (!ready || !isSidebarShortcut(event)) return;
		event.preventDefault();
		setCollapsed(!desktopSidebar.collapsed);
	}}
/>

<div class="flex h-dvh w-full overflow-hidden bg-fc-page">
	{#if ready && !desktopSidebar.collapsed}
		<div class="hidden h-full w-72 shrink-0 p-3 md:block">
			<div class="h-full overflow-hidden rounded-fc-lg bg-fc-component">
				<AppSidebar
					items={navPages}
					user={navUser}
					userMenu={menu}
					userActive={isActive('/account')}
					oncollapse={() => setCollapsed(true)}
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
			<header
				class="flex h-14 items-center gap-2 px-2 {desktopSidebar.collapsed
					? hasNativeTrafficLights()
						? 'md:pl-[76px]'
						: ''
					: 'md:hidden'}"
				data-tauri-drag-region
			>
				<IconButton variant="ghost" aria-label={t('sidebar.open')} onclick={openSidebar}>
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
					title={t('app.unreachableTitle')}
					description={t('app.unreachableBody')}
				>
					<Button icon={icons.refresh} onclick={() => location.reload()}>{t('common.tryAgain')}</Button>
				</EmptyState>
			{:else}
				<div class="flex min-h-[50dvh] items-center justify-center">
					<Spinner />
				</div>
			{/if}
		</div>
	</main>
	{#if ready}
		<MobileSidebar
			items={navPages}
			user={navUser}
			userMenu={menu}
			userActive={isActive('/account')}
		/>
	{/if}
</div>
