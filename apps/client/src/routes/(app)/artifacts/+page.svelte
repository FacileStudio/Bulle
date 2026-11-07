<script lang="ts">
	import { Alert, Button, EmptyState, Icon, Input, PageHeader, Skeleton, Stack, icons } from '@facile/muse';
	import {
		artifactError,
		filterArtifacts,
		groupByMonth,
		listArtifacts,
		type ArtifactList
	} from '$lib/artifacts';
	import { relativeTime } from '$lib/chat/format';
	import MyceliumConnect from '$lib/components/MyceliumConnect.svelte';
	import { spaces } from '$lib/spaces.svelte';

	let list = $state<ArtifactList | null>(null);
	let error = $state<string | null>(null);
	let query = $state('');
	let attempt = $state(0);

	const months = $derived(groupByMonth(filterArtifacts(list?.artifacts ?? [], query)));

	$effect(() => {
		void attempt;
		const space = spaces.currentId;
		let stale = false;
		error = null;
		listArtifacts(space)
			.then((result) => {
				if (!stale) list = result;
			})
			.catch((err: unknown) => {
				if (stale) return;
				const failure = artifactError(err);
				if (failure.connect) list = { artifacts: [], available: false };
				else error = failure.text;
			});
		return () => {
			stale = true;
		};
	});
</script>

<svelte:head>
	<title>Artifacts - Bulle</title>
</svelte:head>

<Stack gap="section">
	<PageHeader
		title="Artifacts"
		description={spaces.current
			? `Documents your assistant wrote in ${spaces.current.name}.`
			: 'Documents your assistant wrote for you.'}
	/>

	{#if error}
		<Alert tone="danger" title="Could not load artifacts">
			<div class="flex flex-col items-start gap-3">
				<span>{error}</span>
				<Button variant="ghost" icon={icons.refresh} class="h-11" onclick={() => (attempt += 1)}>Try again</Button>
			</div>
		</Alert>
	{:else if !list}
		<div class="flex flex-col gap-2" aria-busy="true" aria-label="Loading artifacts">
			<Skeleton class="h-11 w-full" />
			<Skeleton class="h-14 w-full" />
			<Skeleton class="h-14 w-full" />
		</div>
	{:else if !list.available}
		<MyceliumConnect />
	{:else if list.artifacts.length === 0}
		<EmptyState
			icon={icons.layers}
			title="No artifacts yet"
			description="Ask your assistant for a report, a plan or a summary worth keeping, and it lands here."
		/>
	{:else}
		<div class="flex flex-col gap-6">
			<Input
				type="search"
				bind:value={query}
				class="border-0 bg-fc-component"
				placeholder="Filter by title"
				aria-label="Filter artifacts by title"
			/>
			{#each months as month (month.key)}
				<section class="flex flex-col gap-2" aria-labelledby="month-{month.key}">
					<h2 id="month-{month.key}" class="px-1 text-fc-xs font-semibold tracking-wide text-fc-fg-muted uppercase">
						{month.label}
					</h2>
					<ul class="flex flex-col gap-1">
						{#each month.items as artifact (artifact.id)}
							<li>
								<a
									href={`/artifacts/${encodeURIComponent(artifact.id)}`}
									class="flex min-h-14 items-center gap-3 rounded-fc-md bg-fc-component px-4 py-2 transition-colors hover:bg-fc-surface focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
								>
									<Icon icon={icons.layers} size={18} class="text-fc-accent" />
									<span class="flex min-w-0 flex-1 flex-col">
										<span class="truncate text-fc-sm font-medium text-fc-fg">{artifact.title}</span>
										<span class="truncate text-fc-xs text-fc-fg-muted">
											{relativeTime(artifact.updated)}{artifact.machine ? `, from ${artifact.machine}` : ''}
										</span>
									</span>
									<Icon icon={icons.arrow} size={16} class="text-fc-fg-muted" />
								</a>
							</li>
						{/each}
					</ul>
				</section>
			{:else}
				<p class="px-1 text-fc-sm text-fc-fg-muted">No artifact matches "{query}".</p>
			{/each}
		</div>
	{/if}
</Stack>
