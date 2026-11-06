<script lang="ts">
	import { Button } from '@facile/muse';
	import type { Decision } from '$lib/api/types';
	import type { Approval } from '$lib/chat/reducer';
	import MailApprovalCard from './MailApprovalCard.svelte';
	import { mailRecord } from '$lib/chat/mail';
	import { approvalFields, approvalIsDestructive, approvalLabel, approvalSummary, prettyArgs } from '$lib/chat/tools';

	let {
		approval,
		ondecide
	}: { approval: Approval; ondecide: (callId: string, decision: Decision) => void } = $props();

	const settled = $derived(approval.state === 'approved' || approval.state === 'denied');
	const busy = $derived(approval.state === 'sending');
	const fields = $derived(approvalFields(approval.args));
	const destructive = $derived(approvalIsDestructive(approval.tool, approval.args));
	const mail = $derived(mailRecord(approval.args));
	const summary = $derived(approvalSummary(approval.args));
</script>

{#if mail}
	<MailApprovalCard {approval} {mail} {ondecide} />
{:else if settled}
	<p
		class="flex items-baseline gap-2 text-fc-sm text-fc-fg-muted"
		role="group"
		aria-label="Approval: {approvalLabel(approval.tool)}"
	>
		<span>{approvalLabel(approval.tool)}</span>
		{#if summary}
			<span class="min-w-0 truncate text-fc-xs">{summary}</span>
		{/if}
		<span class="text-fc-xs">{approval.state === 'approved' ? 'Approved' : 'Denied'}</span>
	</p>
{:else}
	<div
		class="flex flex-col gap-4 py-2"
		role="group"
		aria-label="Approval needed: {approvalLabel(approval.tool)}"
	>
		<div class="flex flex-col gap-0.5">
			<span class="text-fc-md font-medium {destructive ? 'text-fc-danger' : 'text-fc-fg'}">
				{approvalLabel(approval.tool)}
			</span>
			<span class="text-fc-xs text-fc-fg-muted">Needs your approval</span>
		</div>
		{#if fields && fields.length}
			<dl class="grid max-h-60 grid-cols-[minmax(0,auto)_minmax(0,1fr)] gap-x-4 gap-y-2 overflow-auto text-fc-sm">
				{#each fields as field (field.key)}
					<dt class="text-fc-fg-muted capitalize">{field.key}</dt>
					<dd class="break-words whitespace-pre-wrap text-fc-fg">{field.value}</dd>
				{/each}
			</dl>
		{:else if prettyArgs(approval.args)}
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
				Approve
			</Button>
			<Button
				size="lg"
				variant="ghost"
				disabled={busy}
				onclick={() => ondecide(approval.callId, { approve: false })}
			>
				Deny
			</Button>
		</div>
	</div>
{/if}
