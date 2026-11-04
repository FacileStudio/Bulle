<script lang="ts">
	import { Badge, Button, Icon, icons } from '@facile/muse';
	import type { Decision } from '$lib/api/types';
	import type { Approval } from '$lib/chat/reducer';
	import MailApprovalCard from './MailApprovalCard.svelte';
	import { mailRecord } from '$lib/chat/mail';
	import { approvalFields, approvalIsDestructive, approvalLabel, prettyArgs } from '$lib/chat/tools';

	let {
		approval,
		ondecide
	}: { approval: Approval; ondecide: (callId: string, decision: Decision) => void } = $props();

	const settled = $derived(approval.state === 'approved' || approval.state === 'denied');
	const busy = $derived(approval.state === 'sending');
	const fields = $derived(approvalFields(approval.args));
	const destructive = $derived(approvalIsDestructive(approval.tool, approval.args));
	const mail = $derived(mailRecord(approval.args));
</script>

{#if mail}
	<MailApprovalCard {approval} {mail} {ondecide} />
{:else}

	<div
		class="flex flex-col gap-3 rounded-fc-lg border border-fc-border bg-fc-component p-4"
		role="group"
		aria-label="Approval needed: {approvalLabel(approval.tool)}"
	>
		<div class="flex items-center gap-2 text-fc-sm">
			<Icon icon={icons.shield} size={16} class={destructive ? 'text-fc-danger' : 'text-fc-warning'} />
			<span class="font-medium">{approvalLabel(approval.tool)}</span>
			<span class="text-fc-fg-muted">needs your approval</span>
			{#if settled}
				<span class="ml-auto">
					<Badge tone={approval.state === 'approved' ? 'success' : 'neutral'}>
						{approval.state === 'approved' ? 'Approved' : 'Denied'}
					</Badge>
				</span>
			{/if}
		</div>
		{#if fields && fields.length}
			<dl class="grid max-h-60 grid-cols-[auto_1fr] gap-x-3 gap-y-1 overflow-auto text-fc-sm">
				{#each fields as field (field.key)}
					<dt class="text-fc-fg-muted capitalize">{field.key}</dt>
					<dd class="break-words whitespace-pre-wrap">{field.value}</dd>
				{/each}
			</dl>
		{:else if prettyArgs(approval.args)}
			<pre
				class="max-h-40 overflow-auto rounded-fc-md bg-fc-surface p-3 font-fc-mono text-fc-xs whitespace-pre-wrap">{prettyArgs(
					approval.args
				)}</pre>
		{/if}
		{#if !settled}
			<div class="flex flex-wrap gap-2">
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
					variant="outline"
					disabled={busy}
					onclick={() => ondecide(approval.callId, { approve: false })}
				>
					Deny
				</Button>
			</div>
		{/if}
	</div>
{/if}
