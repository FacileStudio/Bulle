<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Icon, IconButton, NavButton, SpaceSwitcher, StatusDot, UserMenu, icons, type UserMenuOptions } from '@facile/muse';
	import Logo from '$lib/components/Logo.svelte';
	import SessionList from '$lib/components/SessionList.svelte';
	import { isDesktop } from '$lib/platform';
	import { spaces } from '$lib/spaces.svelte';

	type Item = { href: string; label: string; icon?: string; active?: boolean; running?: boolean; count?: number };

	let {
		items,
		user,
		userMenu,
		userActive = false,
		onnavigate,
		oncollapse
	}: {
		items: Item[];
		user?: { name: string; email?: string; avatar?: string; secondary?: string };
		userMenu: UserMenuOptions;
		userActive?: boolean;
		onnavigate?: () => void;
		oncollapse?: () => void;
	} = $props();

	const desktop = isDesktop();

	async function leaveChat() {
		if (page.url.pathname.startsWith('/s/')) await goto('/sessions');
		onnavigate?.();
	}

	function choose(id: string | null) {
		if (!id || id === spaces.currentId) return;
		spaces.select(id);
		void leaveChat();
	}
</script>

<div class="flex h-full min-h-0 flex-col">
	<div class="flex h-14 shrink-0 items-center gap-2 px-4" data-tauri-drag-region>
		{#if desktop}
			<span class="w-[60px] shrink-0 self-stretch" data-tauri-drag-region></span>
		{/if}
		<Logo size={22} />
		<span class="font-fc-title text-fc-md font-semibold tracking-tight">Bulle</span>
		{#if oncollapse}
			<IconButton
				variant="ghost"
				class="ml-auto"
				aria-label={t('sidebar.collapse')}
				title={t('sidebar.collapseHint')}
				onclick={oncollapse}
			>
				<Icon icon={icons.collapse} size={18} />
			</IconButton>
		{/if}
	</div>
	{#if spaces.items.length > 0}
		<div class="flex items-center px-3 pb-2">
			<SpaceSwitcher
				class="min-w-0 flex-1"
				spaces={spaces.items}
				activeId={spaces.currentId}
				personalLabel={null}
				manageHref="/account/spaces"
				onSelect={choose}
			/>
		</div>
	{/if}
	<nav class="flex flex-col gap-1 px-3" aria-label={t('sidebar.pages')}>
		{#each items as item (item.href)}
			<NavButton href={item.href} icon={item.icon} label={item.label} active={item.active}>
				{#snippet right()}
					{#if item.running}
						<StatusDot tone="warning" pulse aria-label={t('sidebar.running')} />
					{/if}
					{#if item.count}
						<span
							class="inline-flex h-5 min-w-5 shrink-0 items-center justify-center rounded-fc-pill bg-fc-accent px-1.5 text-fc-xs font-medium text-fc-accent-fg tabular-nums"
							aria-label={t('sidebar.newCount', { count: String(item.count) })}
						>
							{item.count > 99 ? '99+' : item.count}
						</span>
					{/if}
				{/snippet}
			</NavButton>
		{/each}
	</nav>
	<div class="min-h-0 flex-1">
		<SessionList {onnavigate} />
	</div>
	{#if user}
		<div class="p-3">
			<UserMenu {user} {...userMenu} settingsHref="/account" active={userActive} />
		</div>
	{/if}
</div>
