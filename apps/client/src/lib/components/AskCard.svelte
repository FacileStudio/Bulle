<script lang="ts">
	import { untrack } from 'svelte';
	import { Button, Input } from '@facile/muse';
	import { askComplete, askReply, type AskAnswer, type AskQuestion } from '$lib/chat/ask';

	let {
		questions,
		answered,
		disabled = false,
		onanswer
	}: {
		questions: AskQuestion[];
		answered: boolean;
		disabled?: boolean;
		onanswer: (text: string) => Promise<boolean>;
	} = $props();

	let answers = $state<AskAnswer[]>(untrack(() => questions.map(() => ({ picked: [], typed: '' }))));
	let sent = $state(false);

	const instant = $derived(questions.length === 1 && !questions[0].multiSelect);
	const closed = $derived(answered || sent);
	const ready = $derived(askComplete(answers));

	async function submit() {
		if (closed || disabled || !askComplete(answers)) return;
		sent = true;
		sent = await onanswer(askReply(questions, answers));
	}

	function pick(index: number, label: string) {
		const answer = answers[index];
		if (questions[index].multiSelect) {
			answer.picked = answer.picked.includes(label)
				? answer.picked.filter((l) => l !== label)
				: [...answer.picked, label];
			return;
		}
		answer.picked = [label];
		answer.typed = '';
		if (instant) submit();
	}

	function onkeydown(event: KeyboardEvent) {
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault();
			submit();
		}
	}
</script>

<div class="flex flex-col gap-5 rounded-fc-lg bg-fc-surface p-4" role="group" aria-label="Questions from the assistant">
	{#each questions as q, index (index)}
		<fieldset class="flex flex-col gap-2" disabled={closed}>
			<legend class="flex flex-col gap-0.5 pb-2">
				{#if q.header}
					<span class="text-fc-xs text-fc-fg-muted">{q.header}</span>
				{/if}
				<span class="text-fc-md font-medium text-fc-fg">{q.question}</span>
				{#if q.multiSelect}
					<span class="text-fc-xs text-fc-fg-muted">Pick any that apply</span>
				{/if}
			</legend>
			{#each q.options as option, n (option.label)}
				{@const chosen = answers[index].picked.includes(option.label)}
				<button
					type="button"
					class="flex min-h-11 items-start gap-3 rounded-fc-md px-3 py-2 text-left transition-colors focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring disabled:cursor-default {chosen
						? 'bg-fc-accent text-fc-accent-fg'
						: 'bg-fc-component text-fc-fg hover:bg-fc-page'}"
					aria-pressed={chosen}
					onclick={() => pick(index, option.label)}
				>
					<span class="pt-0.5 text-fc-xs opacity-70">{n + 1}</span>
					<span class="flex min-w-0 flex-col">
						<span class="text-fc-sm font-medium">{option.label}</span>
						{#if option.description}
							<span class="text-fc-xs opacity-80">{option.description}</span>
						{/if}
					</span>
				</button>
			{/each}
			<Input
				bind:value={answers[index].typed}
				{onkeydown}
				oninput={() => {
					if (!q.multiSelect && answers[index].typed.trim()) answers[index].picked = [];
				}}
				class="border-0 bg-fc-component"
				placeholder="Type your answer"
				aria-label="Type your own answer to: {q.question}"
			/>
		</fieldset>
	{/each}
	{#if closed}
		<p class="text-fc-xs text-fc-fg-muted" role="status">Answered</p>
	{:else if !instant || answers.some((a) => a.typed.trim())}
		<div>
			<Button disabled={!ready || disabled} onclick={submit}>Send answer{questions.length > 1 ? 's' : ''}</Button>
		</div>
	{/if}
</div>
