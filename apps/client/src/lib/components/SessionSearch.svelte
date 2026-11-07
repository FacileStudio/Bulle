<script lang="ts">
	import { onDestroy } from 'svelte';
	import { Input, Spinner } from '@facile/muse';
	import { relativeTime } from '$lib/chat/format';
	import { SEARCH_DELAY_MS, debounce, searchSessions, shouldSearch, type SessionHit } from '$lib/chat/search';
	import { errorText } from '$lib/ideas';
	import { spaces } from '$lib/spaces.svelte';

	let {
		query = $bindable(''),
		onnavigate
	}: { query?: string; onnavigate?: () => void } = $props();

	let hits = $state<SessionHit[]>([]);
	let searching = $state(false);
	let error = $state<string | null>(null);
	let latest = 0;

	const active = $derived(shouldSearch(query));

	const run = debounce(async (text: string, space: string | null) => {
		const mine = ++latest;
		try {
			const found = await searchSessions(text, space);
			if (mine === latest) hits = found;
		} catch (err) {
			if (mine === latest) error = errorText(err, 'Search is not available right now.');
		} finally {
			if (mine === latest) searching = false;
		}
	}, SEARCH_DELAY_MS);

	$effect(() => {
		const text = query;
		const space = spaces.currentId;
		error = null;
		if (!shouldSearch(text)) {
			run.cancel();
			latest += 1;
			hits = [];
			searching = false;
			return;
		}
		searching = true;
		run.call(text, space);
	});

	onDestroy(() => run.cancel());

	function onkeydown(event: KeyboardEvent) {
		if (event.key !== 'Escape' || !query) return;
		event.preventDefault();
		event.stopPropagation();
		query = '';
	}
</script>

<div class="flex flex-col gap-2">
	<Input
		type="search"
		bind:value={query}
		{onkeydown}
		class="border-0 bg-fc-surface"
		placeholder="Search chats"
		aria-label="Search past chats"
		aria-controls="session-search-results"
		autocomplete="off"
	/>
	{#if active}
		<div id="session-search-results" aria-live="polite" class="flex flex-col gap-1">
			{#if searching && !hits.length}
				<div class="flex min-h-11 items-center gap-2 px-3 text-fc-xs text-fc-fg-muted">
					<Spinner size="sm" label="Searching" />
					Searching
				</div>
			{:else if error}
				<p class="px-3 py-2 text-fc-xs text-fc-danger" role="alert">{error}</p>
			{:else if !hits.length}
				<p class="px-3 py-2 text-fc-xs text-fc-fg-muted">No chat matches "{query.trim()}".</p>
			{:else}
				<ul class="flex flex-col gap-1" aria-label="Search results">
					{#each hits as hit (hit.session_id)}
						<li>
							<a
								href={`/s/${encodeURIComponent(hit.session_id)}`}
								onclick={() => onnavigate?.()}
								class="flex min-h-11 flex-col justify-center gap-0.5 rounded-fc-md px-3 py-1.5 hover:bg-fc-surface focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
							>
								<span class="truncate text-fc-sm font-medium">{hit.title || 'New chat'}</span>
								{#if hit.snippet}
									<span class="line-clamp-2 text-fc-xs text-fc-fg-muted [overflow-wrap:anywhere]">{hit.snippet}</span>
								{/if}
								<span class="text-fc-xs text-fc-fg-muted">{relativeTime(hit.updated_at)}</span>
							</a>
						</li>
					{/each}
				</ul>
			{/if}
		</div>
	{/if}
</div>
