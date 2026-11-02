<script lang="ts">
	import { Alert } from '@facile/muse';
	import type { ChatItem } from '$lib/chat/reducer';
	import type { OpenUIAction } from '$lib/openui/actions';
	import { session } from '$lib/session.svelte';
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

	const hasPendingApproval = $derived(
		item.approvals.some((a) => a.state === 'pending' || a.state === 'sending')
	);
	const showTyping = $derived(item.streaming && !item.text && !hasPendingApproval);
	const assistantName = $derived(session.user?.assistant?.name || 'Bulle');
	const avatarUrl = $derived(
		session.user?.assistant?.avatar_url || 'https://blobatar.dev/avatar/bulle?gen=2'
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
		<div class="flex items-center gap-2 mb-1">
			<img
				src={avatarUrl}
				alt=""
				aria-hidden="true"
				class="size-6 rounded-full object-cover"
			/>
			<span class="font-medium text-fc-xs text-fc-fg">{assistantName}</span>
		</div>
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
		{#if showTyping}
			<div class="flex items-center gap-1.5 py-1 text-fc-fg-muted" aria-label="Typing">
				<span class="size-2 rounded-full bg-fc-fg-muted/70 dot"></span>
				<span class="size-2 rounded-full bg-fc-fg-muted/70 dot"></span>
				<span class="size-2 rounded-full bg-fc-fg-muted/70 dot"></span>
			</div>
		{/if}
		{#if item.error}
			<Alert tone="danger" title="The run failed">{item.error}</Alert>
		{/if}
	</article>
{/if}

<style>
	@keyframes typing-bounce {
		0%,
		80%,
		100% {
			transform: scale(0.6);
			opacity: 0.35;
		}
		40% {
			transform: scale(1);
			opacity: 1;
		}
	}
	.dot {
		animation: typing-bounce 1.2s infinite ease-in-out both;
	}
	.dot:nth-child(1) {
		animation-delay: -0.32s;
	}
	.dot:nth-child(2) {
		animation-delay: -0.16s;
	}
</style>
