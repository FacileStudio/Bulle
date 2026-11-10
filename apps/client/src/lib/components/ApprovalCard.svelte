<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { Button } from '@facile/muse';
	import type { Decision } from '$lib/api/types';
	import type { Approval } from '$lib/chat/reducer';
	import { serverQuestion } from '$lib/chat/confirm';
	import { approvalFields, approvalIsDestructive, approvalLabel, approvalSummary, prettyArgs } from '$lib/chat/tools';

	let {
		approval,
		ondecide
	}: { approval: Approval; ondecide: (callId: string, decision: Decision) => void } = $props();

	const settled = $derived(approval.state === 'approved' || approval.state === 'denied');
	const busy = $derived(approval.state === 'sending');
	const question = $derived(serverQuestion(approval.args));
	const shown = $derived(question ? question.arguments : approval.args);
	const fields = $derived(approvalFields(shown));
	const destructive = $derived(approvalIsDestructive(approval.tool, shown));
	const summary = $derived(question ? question.message : approvalSummary(approval.args));
</script>

{#if settled}
	<p
		class="flex items-baseline gap-2 text-fc-sm text-fc-fg-muted"
		role="group"
		aria-label={t('approval.settledLabel', { tool: approvalLabel(approval.tool) })}
	>
		<span>{approvalLabel(approval.tool)}</span>
		{#if summary}
			<span class="min-w-0 truncate text-fc-xs">{summary}</span>
		{/if}
		<span class="text-fc-xs">{approval.state === 'approved' ? t('approval.approved') : t('approval.denied')}</span>
	</p>
{:else}
	<div
		class="flex flex-col gap-4 py-2"
		role="group"
		aria-label={t('approval.neededLabel', { tool: approvalLabel(approval.tool) })}
	>
		<div class="flex flex-col gap-0.5">
			<span class="text-fc-md font-medium {destructive ? 'text-fc-danger' : 'text-fc-fg'}">
				{approvalLabel(approval.tool)}
			</span>
			<span class="text-fc-xs text-fc-fg-muted">
				{question?.server ? t('approval.serverAsks', { server: question.server }) : t('approval.needed')}
			</span>
		</div>
		{#if question?.message}
			<p class="text-fc-md text-fc-fg break-words whitespace-pre-wrap">{question.message}</p>
		{/if}
		{#if question?.url.startsWith('https://')}
			<a
				class="inline-flex min-h-11 items-center self-start text-fc-sm font-medium text-fc-accent underline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
				href={question.url} target="_blank" rel="noopener noreferrer">
				{t('approval.openLink')}
			</a>
		{/if}
		{#if fields && fields.length}
			<dl class="grid max-h-60 grid-cols-[minmax(0,auto)_minmax(0,1fr)] gap-x-4 gap-y-2 overflow-auto text-fc-sm">
				{#each fields as field (field.key)}
					<dt class="text-fc-fg-muted capitalize">{field.key}</dt>
					<dd class="break-words whitespace-pre-wrap text-fc-fg">{field.value}</dd>
				{/each}
			</dl>
		{:else if !question && prettyArgs(approval.args)}
			<pre
				class="max-h-40 overflow-auto bg-fc-surface p-3 font-fc-mono text-fc-xs whitespace-pre-wrap rounded-fc-md">{prettyArgs(
					approval.args
				)}</pre>
		{/if}
		<div class="flex flex-wrap items-center gap-2">
			<Button
				size="lg"
				variant={destructive ? 'danger' : undefined}
				disabled={busy}
				onclick={() => ondecide(approval.callId, { approve: true })}
			>
				{t('approval.approve')}
			</Button>
			<Button
				size="lg"
				variant="ghost"
				disabled={busy}
				onclick={() => ondecide(approval.callId, { approve: false })}
			>
				{t('approval.deny')}
			</Button>
		</div>
	</div>
{/if}
