<script lang="ts">
	import { Icon, NavButton, icons } from '@facile/muse';
	import Logo from '$lib/components/Logo.svelte';
	import SessionList from '$lib/components/SessionList.svelte';

	type Item = { href: string; label: string; icon?: string; active?: boolean; dot?: boolean };

	let {
		items,
		userName,
		userActive = false,
		onnavigate
	}: {
		items: Item[];
		userName?: string;
		userActive?: boolean;
		onnavigate?: () => void;
	} = $props();
</script>

<div class="flex h-full min-h-0 flex-col">
	<div class="flex h-14 shrink-0 items-center gap-2 px-4">
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
			<a
				href="/account"
				aria-current={userActive ? 'page' : undefined}
				aria-label="{userName}, settings"
				class="flex min-h-11 w-full items-center gap-2.5 rounded-fc-md px-2.5 py-2 text-fc-sm text-fc-fg transition-colors hover:bg-fc-surface {userActive
					? 'bg-fc-surface'
					: ''}"
			>
				<span
					class="flex size-7 shrink-0 items-center justify-center rounded-fc-pill bg-fc-accent text-fc-xs font-semibold text-fc-accent-fg"
				>
					{userName.charAt(0).toUpperCase()}
				</span>
				<span class="min-w-0 flex-1 truncate">{userName}</span>
				<Icon icon={icons.settings} size={18} />
			</a>
		</div>
	{/if}
</div>
