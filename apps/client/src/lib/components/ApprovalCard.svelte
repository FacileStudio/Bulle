<script lang="ts">
	import { Badge, Button, Icon, icons } from '@facile/muse';
	import type { Approval } from '$lib/chat/reducer';
	import { prettyArgs } from '$lib/chat/tools';

	let {
		approval,
		ondecide
	}: { approval: Approval; ondecide: (callId: string, approve: boolean) => void } = $props();

	const settled = $derived(approval.state === 'approved' || approval.state === 'denied');
	const busy = $derived(approval.state === 'sending');
</script>

<div
	class="flex flex-col gap-3 rounded-fc-lg border border-fc-border bg-fc-component p-4"
	role="group"
	aria-label="Approval needed for {approval.tool}"
>
	<div class="flex items-center gap-2 text-fc-sm">
		<Icon icon={icons.shield} size={16} class="text-fc-warning" />
		<span class="font-medium">Approval needed</span>
		<span class="text-fc-fg-muted">{approval.tool}</span>
		{#if settled}
			<span class="ml-auto">
				<Badge tone={approval.state === 'approved' ? 'success' : 'neutral'}>
					{approval.state === 'approved' ? 'Approved' : 'Denied'}
				</Badge>
			</span>
		{/if}
	</div>
	{#if prettyArgs(approval.args)}
		<pre
			class="max-h-40 overflow-auto rounded-fc-md bg-fc-surface p-3 font-fc-mono text-fc-xs whitespace-pre-wrap">{prettyArgs(
				approval.args
			)}</pre>
	{/if}
	{#if !settled}
		<div class="flex flex-wrap gap-2">
			<Button size="lg" disabled={busy} onclick={() => ondecide(approval.callId, true)}>
				Approve
			</Button>
			<Button
				size="lg"
				variant="outline"
				disabled={busy}
				onclick={() => ondecide(approval.callId, false)}
			>
				Deny
			</Button>
		</div>
	{/if}
</div>
