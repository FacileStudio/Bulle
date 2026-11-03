<script lang="ts">
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Icon, IconButton, NavButton, SpaceSwitcher, icons } from '@facile/muse';
	import CreateSpaceDialog from '$lib/components/CreateSpaceDialog.svelte';
	import Logo from '$lib/components/Logo.svelte';
	import ProfileMenu from '$lib/components/ProfileMenu.svelte';
	import SessionList from '$lib/components/SessionList.svelte';
	import { isDesktop } from '$lib/platform';
	import { spaces } from '$lib/spaces.svelte';

	type Item = { href: string; label: string; icon?: string; active?: boolean; dot?: boolean };

	let {
		items,
		userName,
		userEmail,
		userActive = false,
		onnavigate
	}: {
		items: Item[];
		userName?: string;
		userEmail?: string;
		userActive?: boolean;
		onnavigate?: () => void;
	} = $props();

	const desktop = isDesktop();
	let createOpen = $state(false);

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
	</div>
	{#if spaces.items.length > 0}
		<div class="flex items-center gap-1 px-3 pb-2">
			<SpaceSwitcher
				class="min-w-0 flex-1"
				spaces={spaces.items}
				activeId={spaces.currentId}
				personalLabel={null}
				manageHref="/account/spaces"
				onSelect={choose}
			/>
			<IconButton variant="ghost" aria-label="New space" onclick={() => (createOpen = true)}>
				<Icon icon={icons.plus} size={18} />
			</IconButton>
		</div>
	{/if}
	<nav class="flex flex-col gap-1 px-3" aria-label="Pages">
		{#each items as item (item.href)}
			<NavButton href={item.href} icon={item.icon} label={item.label} active={item.active}>
				{#snippet right()}
					{#if item.dot}
						<span class="size-2 shrink-0 rounded-full bg-fc-warning" role="status" aria-label="A job is running"></span>
					{/if}
				{/snippet}
			</NavButton>
		{/each}
	</nav>
	<div class="min-h-0 flex-1">
		<SessionList {onnavigate} />
	</div>
	{#if userName}
		<div class="p-3">
			<ProfileMenu name={userName} email={userEmail} active={userActive} {onnavigate} />
		</div>
	{/if}
</div>

<CreateSpaceDialog bind:open={createOpen} oncreated={() => void leaveChat()} />
