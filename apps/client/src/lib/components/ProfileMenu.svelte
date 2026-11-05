<script lang="ts">
	import { goto } from '$app/navigation';
	import { Avatar, Icon, icons } from '@facile/muse';
	import { isDesktop, isNative } from '$lib/platform';
	import { signOut } from '$lib/session.svelte';

	type Entry = { href: string; label: string; icon: string };

	let {
		name,
		email = '',
		avatar,
		active = false,
		onnavigate
	}: {
		name: string;
		email?: string;
		avatar?: string;
		active?: boolean;
		onnavigate?: () => void;
	} = $props();

	const entries: Entry[] = [
		{ href: '/account', label: 'Settings', icon: icons.settings },
		{ href: '/account/notifications', label: 'Notifications', icon: icons.notification },
		...(isNative()
			? []
			: [
					{
						href: '/install',
						label: isDesktop() ? 'Get the apps' : 'Get the desktop app',
						icon: icons.download
					}
				])
	];

	let open = $state(false);
	let root = $state<HTMLElement | null>(null);
	let trigger = $state<HTMLButtonElement | null>(null);

	function close() {
		open = false;
	}

	function picked() {
		close();
		onnavigate?.();
	}

	async function leave() {
		close();
		await signOut();
		await goto('/login', { replaceState: true });
	}

	$effect(() => {
		if (!open) return;
		const outside = (event: MouseEvent) => {
			if (root && !root.contains(event.target as Node)) close();
		};
		const escape = (event: KeyboardEvent) => {
			if (event.key !== 'Escape') return;
			event.stopPropagation();
			close();
			trigger?.focus();
		};
		document.addEventListener('click', outside);
		document.addEventListener('keydown', escape);
		return () => {
			document.removeEventListener('click', outside);
			document.removeEventListener('keydown', escape);
		};
	});
</script>

<div bind:this={root} class="relative">
	{#if open}
		<div
			role="menu"
			aria-label="Account"
			class="absolute right-0 bottom-full left-0 z-30 mb-2 flex flex-col gap-0.5 rounded-fc-lg bg-fc-page p-1.5 shadow-lg"
		>
			<div class="flex flex-col px-2.5 pt-1.5 pb-2">
				<span class="truncate text-fc-sm font-medium text-fc-fg">{name}</span>
				{#if email && email !== name}
					<span class="truncate text-fc-xs text-fc-fg-muted">{email}</span>
				{/if}
			</div>
			{#each entries as entry (entry.href)}
				<a
					href={entry.href}
					role="menuitem"
					onclick={picked}
					class="flex min-h-10 items-center gap-2.5 rounded-fc-md px-2.5 text-fc-sm text-fc-fg transition-colors hover:bg-fc-surface focus-visible:bg-fc-surface focus-visible:outline-none"
				>
					<Icon icon={entry.icon} size={16} class="text-fc-fg-muted" />
					{entry.label}
				</a>
			{/each}
			<div class="my-1"></div>
			<button
				type="button"
				role="menuitem"
				onclick={leave}
				class="flex min-h-10 items-center gap-2.5 rounded-fc-md px-2.5 text-left text-fc-sm text-fc-danger transition-colors hover:bg-fc-surface focus-visible:bg-fc-surface focus-visible:outline-none"
			>
				<Icon icon={icons.logout} size={16} />
				Sign out
			</button>
		</div>
	{/if}
	<button
		bind:this={trigger}
		type="button"
		aria-haspopup="menu"
		aria-expanded={open}
		onclick={() => (open = !open)}
		class="flex min-h-11 w-full items-center gap-2.5 rounded-fc-md px-2 py-1.5 text-left text-fc-sm text-fc-fg transition-colors hover:bg-fc-surface focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring {open ||
		active
			? 'bg-fc-surface'
			: ''}"
	>
		<Avatar {name} src={avatar} size="sm" />
		<span class="min-w-0 flex-1 truncate font-medium">{name}</span>
	</button>
</div>
