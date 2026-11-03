<script lang="ts">
	import { page } from '$app/state';
	import { Blob } from '@facile/blob';
	import { IconButton, Icon } from '@facile/muse';
	import { chat } from '$lib/chat/chat.svelte';
	import { session } from '$lib/session.svelte';
	import { CHAT_ICONS } from '$lib/icons';
	import { desktopSidebar, openSidebar } from '$lib/mobile-sidebar.svelte';
	import { hasNativeTrafficLights } from '$lib/platform';

	let { children } = $props();

	const title = $derived(chat.title || 'Chat');
	const assistantName = $derived(session.user?.assistant?.name || 'Bulle');
</script>

<section class="relative flex h-full min-w-0 flex-col">
	{#if page.params.id}
		<header
			class="flex h-14 shrink-0 items-center gap-2 px-2 sm:px-4 {desktopSidebar.collapsed &&
			hasNativeTrafficLights()
				? 'md:pl-[76px]'
				: ''}"
			data-tauri-drag-region
		>
			<IconButton
				variant="ghost"
				class={desktopSidebar.collapsed ? '' : 'md:hidden'}
				aria-label="Open the sidebar"
				onclick={openSidebar}
			>
				<Icon icon={CHAT_ICONS.menu} size={18} />
			</IconButton>
			<h1 class="min-w-0 flex-1 truncate text-fc-md font-medium">{title}</h1>
			<Blob name={assistantName} size={24} title={assistantName} class="size-6 shrink-0" />
		</header>
	{:else}
		<button
			type="button"
			aria-label="Open the sidebar"
			class="fixed top-[max(0.75rem,env(safe-area-inset-top))] z-30 flex size-11 items-center justify-center rounded-full bg-black/30 text-white backdrop-blur-md transition-transform active:scale-95 {desktopSidebar.collapsed
				? hasNativeTrafficLights()
					? 'left-3 md:left-[76px]'
					: 'left-3'
				: 'left-3 md:hidden'}"
			onclick={openSidebar}
		>
			<Icon icon={CHAT_ICONS.menu} size={20} />
		</button>
	{/if}

	<div class="min-h-0 flex-1">
		{@render children()}
	</div>
</section>
