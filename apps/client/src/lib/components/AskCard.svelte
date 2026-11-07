<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { untrack } from 'svelte';
	import { Button, Input } from '@facile/muse';
	import { askComplete, askReply, askShortcut, type AskAnswer, type AskQuestion } from '$lib/chat/ask';

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
		if (event.isComposing) return;
		if (event.key === 'Enter' && !event.shiftKey) {
			event.preventDefault();
			submit();
		}
	}

	/** Picks an option with the digit keys 1 to 4, as the terminal does, unless the user is typing. */
	function onshortcut(event: KeyboardEvent) {
		if (closed || disabled || event.isComposing || event.defaultPrevented) return;
		const hit = askShortcut(event, questions, answers);
		if (!hit) return;
		event.preventDefault();
		pick(hit.question, hit.label);
	}
</script>

<svelte:window onkeydown={onshortcut} />

<div class="flex flex-col gap-5 rounded-fc-lg bg-fc-surface p-4" role="group" aria-label={t('ask.label')}>
	{#each questions as q, index (index)}
		<fieldset class="flex flex-col gap-2" disabled={closed || disabled}>
			<legend class="flex flex-col gap-0.5 pb-2">
				{#if q.header}
					<span class="text-fc-xs text-fc-fg-muted">{q.header}</span>
				{/if}
				<span class="text-fc-md font-medium text-fc-fg">{q.question}</span>
				{#if q.multiSelect}
					<span class="text-fc-xs text-fc-fg-muted">{t('ask.pickAny')}</span>
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
				placeholder={t('ask.typePlaceholder')}
				aria-label={t('ask.typeLabel', { question: q.question })}
			/>
		</fieldset>
	{/each}
	{#if closed}
		<p class="text-fc-xs text-fc-fg-muted" role="status">{t('ask.answered')}</p>
	{:else if !instant || answers.some((a) => a.typed.trim())}
		<div>
			<Button disabled={!ready || disabled} onclick={submit}>{t('ask.send', { count: questions.length })}</Button>
		</div>
	{/if}
</div>
