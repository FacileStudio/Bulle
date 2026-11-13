<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { page } from '$app/state';
	import { Blob } from '@facile/blob';
	import { IconButton, Icon, Tooltip } from '@facile/muse';
	import { chat } from '$lib/chat/chat.svelte';
	import { sessions } from '$lib/chat/sessions.svelte';
	import ArtifactPanel from '$lib/components/ArtifactPanel.svelte';
	import SessionStatus from '$lib/components/SessionStatus.svelte';
	import { assistantName } from '$lib/session.svelte';
	import { CHAT_ICONS } from '$lib/icons';
	import { desktopSidebar, openSidebar } from '$lib/mobile-sidebar.svelte';
	import { hasNativeTrafficLights } from '$lib/platform';

	let { children } = $props();

	const title = $derived(chat.title || t('chat.defaultTitle'));
	const name = $derived(assistantName());
	const listed = $derived(sessions.statusOf(page.params.id));
	const status = $derived(listed === 'idle' && chat.state.running ? 'running' : listed);
</script>

<div class="flex h-full min-w-0">
	<section class="relative flex h-full min-w-0 flex-1 flex-col">
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
					aria-label={t('sidebar.open')}
					onclick={openSidebar}
				>
					<Icon icon={CHAT_ICONS.menu} size={18} />
				</IconButton>
				<h1 class="min-w-0 flex-1 truncate text-fc-md font-medium">{title}</h1>
				<SessionStatus {status} showLabel class="shrink-0" />
				<Blob {name} size={24} title={name} class="size-6 shrink-0" />
			</header>
		{:else}
			<Tooltip text={t('sidebar.open')} describe={false} class="contents">
				<button
					type="button"
					aria-label={t('sidebar.open')}
					class="fixed top-[max(0.75rem,env(safe-area-inset-top))] z-30 flex size-11 items-center justify-center rounded-full bg-fc-surface text-fc-fg backdrop-blur-md transition-transform active:scale-95 {desktopSidebar.collapsed
						? hasNativeTrafficLights()
							? 'left-3 md:left-[76px]'
							: 'left-3'
						: 'left-3 md:hidden'}"
					onclick={openSidebar}
				>
					<Icon icon={CHAT_ICONS.menu} size={20} />
				</button>
			</Tooltip>
		{/if}

		<div class="min-h-0 flex-1">
			{@render children()}
		</div>
	</section>
	<ArtifactPanel />
</div>
