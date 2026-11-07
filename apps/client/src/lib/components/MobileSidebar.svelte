<script lang="ts">
	import { page } from '$app/state';
	import AppSidebar from '$lib/components/AppSidebar.svelte';
	import { listenForSwipes, mobileSidebar } from '$lib/mobile-sidebar.svelte';

	type Item = { href: string; label: string; icon?: string; active?: boolean; dot?: boolean };

	let {
		items,
		userName,
		userEmail,
		userAvatar,
		userActive = false,
		jobsRunning = false
	}: {
		items: Item[];
		userName?: string;
		userEmail?: string;
		userAvatar?: string;
		userActive?: boolean;
		jobsRunning?: boolean;
	} = $props();

	const close = () => (mobileSidebar.open = false);

	$effect(() => listenForSwipes());

	$effect(() => {
		void page.url.pathname;
		close();
	});

	$effect(() => {
		if (!mobileSidebar.open) return;
		const onkey = (event: KeyboardEvent) => event.key === 'Escape' && close();
		window.addEventListener('keydown', onkey);
		return () => window.removeEventListener('keydown', onkey);
	});
</script>

<div class="fixed inset-0 z-40 md:hidden {mobileSidebar.open ? '' : 'pointer-events-none'}">
	<button
		type="button"
		tabindex="-1"
		aria-label="Close the sidebar"
		class="absolute inset-0 bg-black/40 transition-opacity duration-300 {mobileSidebar.open
			? 'opacity-100'
			: 'opacity-0'}"
		onclick={close}
	></button>
	<aside
		inert={!mobileSidebar.open}
		aria-label="Navigation"
		class="absolute inset-y-0 left-0 flex w-[85%] max-w-80 flex-col bg-fc-page shadow-xl transition-transform duration-300 ease-out {mobileSidebar.open
			? 'translate-x-0'
			: '-translate-x-full'}"
	>
		<AppSidebar {items} {userName} {userEmail} {userAvatar} {userActive} {jobsRunning} onnavigate={close} />
	</aside>
</div>
