<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { IconButton, Icon } from '@facile/muse';
	import SessionList from '$lib/components/SessionList.svelte';
	import { chat } from '$lib/chat/chat.svelte';
	import { sessions } from '$lib/chat/sessions.svelte';
	import { session } from '$lib/session.svelte';
	import { CHAT_ICONS } from '$lib/icons';
	import { mobileSidebar } from '$lib/mobile-sidebar.svelte';

	let { children } = $props();

	const title = $derived(page.params.id ? chat.title || 'Chat' : 'New chat');
	const assistantName = $derived(session.user?.assistant?.name || 'Bulle');
	const avatarUrl = $derived(
		session.user?.assistant?.avatar_url || 'https://blobatar.dev/avatar/bulle?gen=2'
	);

	onMount(() => {
		void sessions.refresh();
	});
</script>

<div class="flex h-full min-h-0 w-full">
	<aside class="hidden w-72 shrink-0 border-r border-fc-border md:block">
		<SessionList />
	</aside>

	<section class="flex min-w-0 flex-1 flex-col">
		<header class="flex h-14 shrink-0 items-center gap-2 border-b border-fc-border px-2 sm:px-4">
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

		<div class="min-h-0 flex-1">
			{@render children()}
		</div>
	</section>
</div>
