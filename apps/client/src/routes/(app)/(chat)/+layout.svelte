<script lang="ts">
	import { page } from '$app/state';
	import { IconButton, Icon } from '@facile/muse';
	import { chat } from '$lib/chat/chat.svelte';
	import { session } from '$lib/session.svelte';
	import { CHAT_ICONS } from '$lib/icons';
	import { mobileSidebar } from '$lib/mobile-sidebar.svelte';

	let { children } = $props();

	const title = $derived(chat.title || 'Chat');
	const assistantName = $derived(session.user?.assistant?.name || 'Bulle');
	const avatarUrl = $derived(
		session.user?.assistant?.avatar_url || 'https://blobatar.dev/avatar/bulle?gen=2'
	);
</script>

<section class="relative flex h-full min-w-0 flex-col">
	{#if page.params.id}
		<header class="flex h-14 shrink-0 items-center gap-2 px-2 sm:px-4">
			<IconButton
				variant="ghost"
				class="md:hidden"
				aria-label="Open the sidebar"
				onclick={() => (mobileSidebar.open = true)}
			>
				<Icon icon={CHAT_ICONS.menu} size={18} />
			</IconButton>
			<h1 class="min-w-0 flex-1 truncate text-fc-md font-medium">{title}</h1>
			<img
				src={avatarUrl}
				alt=""
				aria-hidden="true"
				class="size-6 rounded-full object-cover"
				title={assistantName}
			/>
		</header>
	{:else}
		<button
			type="button"
			aria-label="Open the sidebar"
			class="fixed top-[max(0.75rem,env(safe-area-inset-top))] left-3 z-30 flex size-11 items-center justify-center rounded-full bg-black/30 text-white backdrop-blur-md transition-transform active:scale-95 md:hidden"
			onclick={() => (mobileSidebar.open = true)}
		>
			<Icon icon={CHAT_ICONS.menu} size={20} />
		</button>
	{/if}

	<div class="min-h-0 flex-1">
		{@render children()}
	</div>
</section>
