<script lang="ts">
	import { Button } from '@facile/muse';
	import type { Decision } from '$lib/api/types';
	import type { Approval } from '$lib/chat/reducer';
	import { approvalFields, approvalIsDestructive, approvalLabel } from '$lib/chat/tools';

	let {
		approvals,
		ondecideall
	}: { approvals: Approval[]; ondecideall: (callIds: string[], decision: Decision) => void } =
		$props();

	const tool = $derived(approvals[0].tool);
	const count = $derived(approvals.length);
	const busy = $derived(approvals.some((a) => a.state === 'sending'));
	const destructive = $derived(approvals.some((a) => approvalIsDestructive(a.tool, a.args)));
	const callIds = $derived(approvals.map((a) => a.callId));

	function summary(approval: Approval): string {
		const fields = approvalFields(approval.args);
		if (!fields || !fields.length) return '';
		return fields
			.slice(0, 3)
			.map((f) => f.value)
			.join(' / ');
	}
</script>

<div
	class="flex flex-col gap-4 py-2"
	role="group"
	aria-label="Approval needed: {count} x {approvalLabel(tool)}"
>
	<div class="flex flex-col gap-0.5">
		<span class="text-fc-md font-medium {destructive ? 'text-fc-danger' : 'text-fc-fg'}">
			{approvalLabel(tool)} x {count}
		</span>
		<span class="text-fc-xs text-fc-fg-muted">Needs your approval</span>
	</div>
	<ul class="flex max-h-72 flex-col gap-2 overflow-auto text-fc-sm">
		{#each approvals as approval (approval.callId)}
			<li class="truncate text-fc-fg">{summary(approval) || approvalLabel(approval.tool)}</li>
		{/each}
	</ul>
	<div class="flex flex-wrap items-center gap-2">
		<Button
			size="lg"
			variant={destructive ? 'danger' : undefined}
			disabled={busy}
			onclick={() => ondecideall(callIds, { approve: true })}
		>
			Approve all {count}
		</Button>
		<Button
			size="lg"
			variant="ghost"
			disabled={busy}
			onclick={() => ondecideall(callIds, { approve: false })}
		>
			Deny all
		</Button>
	</div>
	<button
		type="button"
		class="min-h-11 self-start text-fc-xs text-fc-fg-muted underline underline-offset-4 disabled:opacity-50"
		disabled={busy}
		onclick={() => ondecideall(callIds, { approve: true, remember: true })}
	>
		Approve all and skip the rest of this run
	</button>
</div>
