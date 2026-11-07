<script lang="ts">
	import { Blob } from '@facile/blob';
	import { Alert } from '@facile/muse';
	import type { Decision } from '$lib/api/types';
	import type { ChatItem } from '$lib/chat/reducer';
	import type { OpenUIAction } from '$lib/openui/actions';
	import { session } from '$lib/session.svelte';
	import { artifactCards } from '$lib/artifacts';
	import ApprovalBatch from './ApprovalBatch.svelte';
	import ArtifactCard from './ArtifactCard.svelte';
	import ApprovalCard from './ApprovalCard.svelte';
	import AskCard from './AskCard.svelte';
	import { ASK_TOOL, askQuestions } from '$lib/chat/ask';
	import { groupApprovals } from '$lib/chat/batch';
	import MessageAttachments from './MessageAttachments.svelte';
	import Reply from './Reply.svelte';
	import Thinking from './Thinking.svelte';
	import ToolGroups from './ToolGroups.svelte';

	let {
		item,
		focus = false,
		ondecide,
		ondecideall,
		onaction,
		answered = false,
		busy = false,
		onanswer
	}: {
		item: ChatItem;
		focus?: boolean;
		ondecide: (callId: string, decision: Decision) => void;
		ondecideall: (callIds: string[], decision: Decision) => void;
		onaction?: (action: OpenUIAction) => void;
		answered?: boolean;
		busy?: boolean;
		onanswer?: (text: string) => Promise<boolean>;
	} = $props();

	let showSteps = $state(false);

	const stepsHidden = $derived(focus && !showSteps);
	const hasSteps = $derived(item.tools.length > 0 || !!item.thinking);
	const shownText = $derived(
		stepsHidden ? item.answer || (item.streaming ? '' : item.text) : item.text
	);

	const hasPendingApproval = $derived(
		item.approvals.some((a) => a.state === 'pending' || a.state === 'sending')
	);
	const entries = $derived(groupApprovals(item.approvals));
	const showTyping = $derived(item.streaming && !shownText && !hasPendingApproval);
	const assistantName = $derived(session.user?.assistant?.name || 'Bulle');
	const artifacts = $derived(artifactCards(item.tools));
	const steps = $derived(item.tools.filter((t) => t.name !== ASK_TOOL));
	const asks = $derived(
		item.tools.flatMap((t) => {
			const questions = t.name === ASK_TOOL && t.status !== 'failed' ? askQuestions(t.args) : null;
			return questions ? [{ callId: t.callId, questions }] : [];
		})
	);
</script>

{#if item.role === 'user'}
	<div class="flex flex-col items-end gap-2">
		<MessageAttachments attachments={item.attachments} />
		{#if item.text}
			<p
				class="max-w-[85%] rounded-fc-lg bg-fc-surface px-4 py-2 text-fc-md whitespace-pre-wrap [overflow-wrap:anywhere]"
			>
				{item.text}
			</p>
		{/if}
	</div>
{:else}
	<article class="flex flex-col gap-3">
		<div class="flex items-center gap-2 mb-1">
			<Blob name={assistantName} size={24} class="size-6 shrink-0" />
			<span class="font-medium text-fc-xs text-fc-fg">{assistantName}</span>
		</div>
		{#if focus && hasSteps}
			<button
				type="button"
				class="self-start text-fc-xs text-fc-fg-muted hover:text-fc-fg"
				aria-expanded={showSteps}
				onclick={() => (showSteps = !showSteps)}
			>
				{showSteps ? 'Hide steps' : 'Show steps'}
			</button>
		{/if}
		{#if item.thinking && !stepsHidden}
			<Thinking text={item.thinking} active={item.streaming && !item.text} />
		{/if}
		{#if steps.length && !stepsHidden}
			<ToolGroups tools={steps} />
		{/if}
		{#each entries as entry (entry.kind === 'batch' ? `batch-${entry.tool}` : entry.approval.callId)}
			{#if entry.kind === 'batch'}
				<ApprovalBatch approvals={entry.approvals} ondecideall={ondecideall} />
			{:else}
				<ApprovalCard approval={entry.approval} {ondecide} />
			{/if}
		{/each}
		{#each artifacts as artifact (artifact.id)}
			<ArtifactCard {artifact} />
		{/each}
		{#if shownText}
			<Reply text={shownText} streaming={item.streaming} {onaction} />
		{/if}
		{#each asks as ask (`${ask.callId}-${JSON.stringify(ask.questions)}`)}
			<AskCard
				questions={ask.questions}
				{answered}
				disabled={busy || !onanswer}
				onanswer={(text) => onanswer?.(text) ?? Promise.resolve(false)}
			/>
		{/each}
		{#if showTyping}
			<div class="flex items-center gap-1.5 py-1 text-fc-fg-muted" aria-label="Typing">
				<span class="size-2 rounded-full bg-fc-fg-muted/70 dot"></span>
				<span class="size-2 rounded-full bg-fc-fg-muted/70 dot"></span>
				<span class="size-2 rounded-full bg-fc-fg-muted/70 dot"></span>
				{#if focus}
					<span class="ml-1 text-fc-sm">Working on it</span>
				{/if}
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
