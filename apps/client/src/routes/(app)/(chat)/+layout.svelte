<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { Alert, Drawer, IconButton, Icon } from '@facile/muse';
	import ComputerChip from '$lib/components/ComputerChip.svelte';
	import SessionList from '$lib/components/SessionList.svelte';
	import { chat } from '$lib/chat/chat.svelte';
	import { computer } from '$lib/chat/computer.svelte';
	import { sessions } from '$lib/chat/sessions.svelte';
	import { CHAT_ICONS } from '$lib/icons';

	let { children } = $props();

	let drawerOpen = $state(false);

	const title = $derived(page.params.id ? chat.title || 'Chat' : 'New chat');

	onMount(() => {
		void sessions.refresh();
		void computer.watch();
		return () => computer.unwatch();
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
				aria-label="Open sessions"
				onclick={() => (drawerOpen = true)}
			>
				<Icon icon={CHAT_ICONS.menu} size={18} />
			</IconButton>
			<h1 class="min-w-0 flex-1 truncate text-fc-md font-medium">{title}</h1>
			<ComputerChip />
		</header>

		{#if computer.value?.status === 'creating'}
			<Alert tone="info" class="m-3 mb-0">
				Booting your computer, first run takes a few minutes. You can keep chatting meanwhile.
			</Alert>
		{:else if computer.value?.status === 'error'}
			<Alert tone="danger" title="The computer hit an error" class="m-3 mb-0">
				{computer.value.error || 'It did not start. Try starting it again.'}
			</Alert>
		{:else if computer.error}
			<Alert tone="warning" class="m-3 mb-0">{computer.error}</Alert>
		{/if}

		<div class="min-h-0 flex-1">
			{@render children()}
		</div>
	</section>
</div>

<Drawer bind:open={drawerOpen} title="Sessions" class="md:hidden">
	<div class="-mx-4 h-[60dvh]">
		<SessionList onnavigate={() => (drawerOpen = false)} />
	</div>
</Drawer>
