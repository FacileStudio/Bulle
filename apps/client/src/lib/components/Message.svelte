<script lang="ts">
	import { Alert, Spinner } from '@facile/muse';
	import type { ChatItem } from '$lib/chat/reducer';
	import type { OpenUIAction } from '$lib/openui/actions';
	import ApprovalCard from './ApprovalCard.svelte';
	import Reply from './Reply.svelte';
	import Thinking from './Thinking.svelte';
	import ToolGroups from './ToolGroups.svelte';

	let {
		item,
		ondecide,
		onaction
	}: {
		item: ChatItem;
		ondecide: (callId: string, approve: boolean) => void;
		onaction?: (action: OpenUIAction) => void;
	} = $props();

	const idle = $derived(
		item.streaming && !item.text && !item.thinking && !item.tools.length && !item.approvals.length
	);
</script>

{#if item.role === 'user'}
	<div class="flex justify-end">
		<p
			class="max-w-[85%] rounded-fc-lg bg-fc-surface px-4 py-2 text-fc-md whitespace-pre-wrap [overflow-wrap:anywhere]"
		>
			{item.text}
		</p>
	</div>
{:else}
	<article class="flex flex-col gap-3">
		{#if item.thinking}
			<Thinking text={item.thinking} active={item.streaming && !item.text} />
		{/if}
		{#if item.tools.length}
			<ToolGroups tools={item.tools} />
		{/if}
		{#each item.approvals as approval (approval.callId)}
			<ApprovalCard {approval} {ondecide} />
		{/each}
		{#if item.text}
			<Reply text={item.text} streaming={item.streaming} {onaction} />
		{/if}
		{#if idle}
			<div class="flex items-center gap-2 text-fc-sm text-fc-fg-muted">
				<Spinner />
				<span>Working</span>
			</div>
		{/if}
		{#if item.error}
			<Alert tone="danger" title="The run failed">{item.error}</Alert>
		{/if}
	</article>
{/if}
