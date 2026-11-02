<script lang="ts">
	import { Icon } from '@facile/muse';
	import { CHAT_ICONS } from '$lib/icons';
	import { getCaretCoordinates, type CaretCoordinates } from '$lib/caret';

	let {
		running = false,
		disabled = false,
		placeholder = 'Message the agent',
		onsend,
		onstop,
		oncaret
	}: {
		running?: boolean;
		disabled?: boolean;
		placeholder?: string;
		onsend: (text: string) => Promise<boolean>;
		onstop?: () => void;
		oncaret?: (coords: CaretCoordinates | null) => void;
	} = $props();

	let draft = $state('');
	let sending = $state(false);
	let field = $state<HTMLTextAreaElement | null>(null);

	const canSend = $derived(!!draft.trim() && !running && !disabled && !sending);

	function fit() {
		if (!field) return;
		field.style.height = 'auto';
		field.style.height = `${Math.min(field.scrollHeight, 200)}px`;
	}

	function emitCaret() {
		if (!field || !oncaret) return;
		oncaret(getCaretCoordinates(field, field.selectionStart));
	}

	$effect(() => {
		void draft;
		fit();
	});

	async function submit() {
		if (!canSend) return;
		const text = draft;
		sending = true;
		draft = '';
		oncaret?.(null);
		const ok = await onsend(text);
		sending = false;
		if (!ok) draft = text;
		field?.focus();
	}

	function onkeydown(event: KeyboardEvent) {
		if (event.key !== 'Enter' || event.shiftKey || event.isComposing) return;
		event.preventDefault();
		void submit();
	}
</script>

<form
	class="flex items-end gap-2 rounded-fc-lg border border-fc-border bg-fc-bg p-2 pl-4 focus-within:border-fc-ring"
	onsubmit={(event) => {
		event.preventDefault();
		void submit();
	}}
>
	<textarea
		bind:this={field}
		bind:value={draft}
		{onkeydown}
		oninput={emitCaret}
		onkeyup={emitCaret}
		onclick={emitCaret}
		onselect={emitCaret}
		onfocus={emitCaret}
		onblur={() => oncaret?.(null)}
		{placeholder}
		rows="1"
		aria-label="Message"
		class="max-h-[200px] min-h-11 flex-1 resize-none bg-transparent py-2.5 text-fc-md text-fc-fg placeholder:text-fc-fg-muted focus:outline-none"
	></textarea>
	{#if running}
		<button
			type="button"
			aria-label="Stop the run"
			class="inline-flex size-11 shrink-0 items-center justify-center rounded-fc-pill bg-fc-accent text-fc-accent-fg hover:opacity-90 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
			onclick={() => onstop?.()}
		>
			<Icon icon={CHAT_ICONS.stop} size={18} />
		</button>
	{:else}
		<button
			type="submit"
			aria-label="Send message"
			disabled={!canSend}
			class="inline-flex size-11 shrink-0 items-center justify-center rounded-fc-pill bg-fc-accent text-fc-accent-fg hover:opacity-90 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring disabled:cursor-not-allowed disabled:opacity-40"
		>
			<Icon icon={CHAT_ICONS.send} size={18} />
		</button>
	{/if}
</form>
