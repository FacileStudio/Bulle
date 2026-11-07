<script lang="ts">
	import { page } from '$app/state';
	import { Button, Icon, IconButton, icons } from '@facile/muse';
	import { artifactPanel, closeArtifact } from '$lib/artifact-panel.svelte';
	import { spaces } from '$lib/spaces.svelte';
	import ArtifactView from './ArtifactView.svelte';

	const titleId = $props.id();

	let header = $state<HTMLElement | null>(null);
	let title = $state('');

	const ref = $derived(artifactPanel.ref);
	const spaceId = $derived(ref?.space_id || spaces.currentId);

	$effect(() => {
		void page.url.pathname;
		artifactPanel.ref = null;
	});

	$effect(() => {
		if (!ref) return;
		title = ref.title;
		queueMicrotask(() => header?.querySelector<HTMLButtonElement>('[data-close]')?.focus());
		const onkey = (event: KeyboardEvent) => {
			if (event.key !== 'Escape') return;
			event.preventDefault();
			closeArtifact();
		};
		window.addEventListener('keydown', onkey);
		return () => window.removeEventListener('keydown', onkey);
	});
</script>

{#if ref}
	<div
		role="dialog"
		aria-labelledby={titleId}
		class="fixed inset-0 z-50 flex flex-col bg-fc-page md:static md:z-auto md:h-full md:w-[min(42rem,48%)] md:shrink-0 md:py-3 md:pr-3"
	>
		<div class="flex min-h-0 flex-1 flex-col bg-fc-component md:rounded-fc-lg">
			<header bind:this={header} class="flex shrink-0 items-center gap-1 py-2 pr-2 pl-4 pt-[max(0.5rem,env(safe-area-inset-top))]">
				<h2 id={titleId} class="min-w-0 flex-1 truncate text-fc-md font-semibold text-fc-fg">{title}</h2>
				<Button
					variant="ghost"
					href={`/artifacts/${encodeURIComponent(ref.id)}`}
					icon={icons.arrow}
					class="h-11 px-3"
					aria-label="Open in Artifacts"
				>
					<span class="hidden sm:inline">Open in Artifacts</span>
				</Button>
				<IconButton variant="ghost" aria-label="Close the artifact" data-close onclick={closeArtifact}>
					<Icon icon={icons.close} size={18} />
				</IconButton>
			</header>
			<div class="min-h-0 flex-1 overflow-y-auto overscroll-contain px-4 pb-[max(1.5rem,env(safe-area-inset-bottom))] sm:px-6">
				{#key ref.id}
					<ArtifactView id={ref.id} {spaceId} onload={(a) => (title = a.title)} onnavigate={closeArtifact} />
				{/key}
			</div>
		</div>
	</div>
{/if}
