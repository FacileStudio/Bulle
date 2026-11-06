<script lang="ts">
	import { Badge, Button, Card, Icon, icons } from '@facile/muse';
	import { DISMISS_REASONS, type Idea, type IdeaKind, type IdeaPatch } from '$lib/ideas';

	type Props = {
		idea: Idea;
		busy?: boolean;
		onstart: (idea: Idea) => void;
		onpatch: (idea: Idea, patch: IdeaPatch) => void;
	};

	let { idea, busy = false, onstart, onpatch }: Props = $props();

	let expanded = $state(false);
	let dismissing = $state(false);

	const KIND_ICONS: Record<IdeaKind, string> = {
		mail: icons.mail,
		task: icons.check,
		event: icons.calendar,
		follow_up: icons.history,
		research: icons.search,
		other: icons.widget
	};

	const saved = $derived(idea.status === 'saved');
	const detailsId = $derived(`idea-details-${idea.id}`);

	function rate(value: 'up' | 'down') {
		onpatch(idea, { feedback: idea.feedback === value ? '' : value });
	}

	function dismiss(reason = '') {
		dismissing = false;
		onpatch(idea, reason ? { status: 'dismissed', reason } : { status: 'dismissed' });
	}
</script>

<Card class="flex flex-col gap-3 p-4 sm:p-5">
	<div class="flex items-start gap-3">
		<span
			class="flex size-9 shrink-0 items-center justify-center rounded-fc-pill bg-fc-surface text-fc-fg-muted"
			aria-hidden="true"
		>
			<Icon icon={KIND_ICONS[idea.kind] ?? icons.widget} size={18} />
		</span>
		<div class="flex min-w-0 flex-col gap-1">
			<h3 class="text-fc-base font-medium wrap-anywhere text-fc-fg">{idea.title}</h3>
			<p class="text-fc-sm wrap-anywhere text-fc-fg-muted">{idea.why}</p>
		</div>
	</div>

	{#if idea.sources.length > 0}
		<ul class="flex flex-wrap gap-1.5" aria-label="Sources">
			{#each idea.sources as source, index (index)}
				<li class="max-w-full"><Badge tone="neutral" class="wrap-anywhere">{source.label}</Badge></li>
			{/each}
		</ul>
	{/if}

	<Button
		variant="ghost"
		class="-ml-2 h-11 self-start"
		iconRight={expanded ? icons.chevronUp : icons.chevronDown}
		aria-expanded={expanded}
		aria-controls={detailsId}
		onclick={() => (expanded = !expanded)}
	>
		{expanded ? 'Hide what Bulle will do' : 'Show what Bulle will do'}
	</Button>

	<div id={detailsId} class="flex flex-col gap-3" hidden={!expanded}>
		<div class="flex flex-col gap-1">
			<span class="text-fc-xs font-medium text-fc-fg-muted">Prompt</span>
			<p class="rounded-fc-md bg-fc-surface p-3 text-fc-sm whitespace-pre-wrap wrap-anywhere text-fc-fg">
				{idea.prompt}
			</p>
		</div>
		{#if idea.draft}
			<div class="flex flex-col gap-1">
				<span class="text-fc-xs font-medium text-fc-fg-muted">Draft</span>
				<p class="rounded-fc-md bg-fc-surface p-3 text-fc-sm whitespace-pre-wrap wrap-anywhere text-fc-fg">
					{idea.draft}
				</p>
			</div>
		{/if}
	</div>

	{#if dismissing}
		<div class="flex flex-col gap-2" role="group" aria-labelledby={`${detailsId}-dismiss`}>
			<span id={`${detailsId}-dismiss`} class="text-fc-sm text-fc-fg-muted">Why dismiss it? This is optional.</span>
			<div class="flex flex-wrap gap-2">
				{#each DISMISS_REASONS as reason (reason)}
					<Button variant="outline" class="h-11" disabled={busy} onclick={() => dismiss(reason)}>
						{reason}
					</Button>
				{/each}
				<Button variant="ghost-danger" class="h-11" disabled={busy} onclick={() => dismiss()}>
					Just dismiss
				</Button>
				<Button variant="ghost" class="h-11" onclick={() => (dismissing = false)}>Cancel</Button>
			</div>
		</div>
	{:else}
		<div class="flex flex-wrap items-center gap-2">
			<Button class="h-11" icon={icons.bolt} disabled={busy} aria-busy={busy} onclick={() => onstart(idea)}>
				Do it
			</Button>
			<Button
				variant="outline"
				class="h-11"
				disabled={busy}
				aria-pressed={saved}
				onclick={() => onpatch(idea, { status: saved ? 'new' : 'saved' })}
			>
				{saved ? 'Unsave' : 'Save'}
			</Button>
			<Button
				variant={idea.feedback === 'up' ? 'primary' : 'ghost'}
				class="h-11"
				disabled={busy}
				aria-pressed={idea.feedback === 'up'}
				onclick={() => rate('up')}
			>
				Useful
			</Button>
			<Button
				variant={idea.feedback === 'down' ? 'danger' : 'ghost'}
				class="h-11"
				disabled={busy}
				aria-pressed={idea.feedback === 'down'}
				onclick={() => rate('down')}
			>
				Not useful
			</Button>
			<Button
				variant="ghost-danger"
				class="h-11"
				icon={icons.close}
				disabled={busy}
				onclick={() => (dismissing = true)}
			>
				Dismiss
			</Button>
		</div>
	{/if}
</Card>
