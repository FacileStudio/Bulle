<script lang="ts">
	import { NavButton } from '@facile/muse';
	import Logo from '$lib/components/Logo.svelte';
	import ProfileMenu from '$lib/components/ProfileMenu.svelte';
	import SessionList from '$lib/components/SessionList.svelte';
	import { isDesktop } from '$lib/platform';

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
</script>

<div class="flex h-full min-h-0 flex-col">
	<div class="flex h-14 shrink-0 items-center gap-2 px-4" data-tauri-drag-region>
		{#if desktop}
			<span class="w-[60px] shrink-0 self-stretch" data-tauri-drag-region></span>
		{/if}
		<Logo size={22} />
		<span class="font-fc-title text-fc-md font-semibold tracking-tight">Bulle</span>
	</div>
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
