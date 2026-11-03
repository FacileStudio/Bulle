<script lang="ts">
	import { Icon, icons } from '@facile/muse';
	import { api } from '$lib/api';
	import { CHAT_ICONS } from '$lib/icons';
	import { getCaretCoordinates, type CaretCoordinates } from '$lib/caret';
	import { isNative } from '$lib/platform';
	import {
		ATTACHMENT_ACCEPT,
		admit,
		errorText,
		formatSize,
		isImageType,
		type Attachment,
		type AttachmentError
	} from '$lib/chat/attachments';

	type Pending = {
		key: number;
		name: string;
		size: number;
		progress: number;
		status: 'uploading' | 'done' | 'error';
		error?: AttachmentError;
		attachment?: Attachment;
		preview?: string;
		abort?: () => void;
	};

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
		onsend: (text: string, attachments: Attachment[]) => Promise<boolean>;
		onstop?: () => void;
		oncaret?: (coords: CaretCoordinates | null) => void;
	} = $props();

	let draft = $state('');
	let sending = $state(false);
	let field = $state<HTMLTextAreaElement | null>(null);
	let picker = $state<HTMLInputElement | null>(null);
	let camera = $state<HTMLInputElement | null>(null);
	let library = $state<HTMLInputElement | null>(null);
	let sourcesOpen = $state(false);
	const native = isNative();
	let files = $state<Pending[]>([]);
	let nextKey = 0;

	const uploading = $derived(files.some((f) => f.status === 'uploading'));
	const failed = $derived(files.some((f) => f.status === 'error'));
	const ready = $derived(
		files.flatMap((f) => (f.status === 'done' && f.attachment ? [f.attachment] : []))
	);
	const canSend = $derived(
		(!!draft.trim() || ready.length > 0) &&
			!uploading &&
			!failed &&
			!running &&
			!disabled &&
			!sending
	);
	const hint = $derived(
		uploading ? 'Waiting for uploads to finish' : failed ? 'Remove the files that failed to send' : ''
	);

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
		emitCaret();
	});

	$effect(() => () => {
		for (const f of files) release(f);
	});

	function release(f: Pending) {
		f.abort?.();
		if (f.preview) URL.revokeObjectURL(f.preview);
	}

	function patch(key: number, next: Partial<Pending>) {
		files = files.map((f) => (f.key === key ? { ...f, ...next } : f));
	}

	function upload(key: number, file: File) {
		const handle = api().uploadAttachment(file, (fraction) => patch(key, { progress: fraction }));
		patch(key, { abort: handle.abort });
		handle.done.then(
			(attachment) => patch(key, { status: 'done', progress: 1, attachment, abort: undefined }),
			() => patch(key, { status: 'error', error: 'failed', abort: undefined })
		);
	}

	function add(list: File[]) {
		if (!list.length) return;
		const taken = files.filter((f) => f.status !== 'error').length;
		const added = admit(list, taken).map(({ file, error }) => ({
			file,
			entry: {
				key: nextKey++,
				name: file.name || 'pasted file',
				size: file.size,
				progress: 0,
				status: error ? 'error' : 'uploading',
				error: error ?? undefined,
				preview: !error && isImageType(file.type) ? URL.createObjectURL(file) : undefined
			} satisfies Pending
		}));
		files = [...files, ...added.map((a) => a.entry)];
		for (const { file, entry } of added) {
			if (entry.status === 'uploading') upload(entry.key, file);
		}
	}

	/** Adds files from outside the composer, such as a drop on the whole conversation. */
	export function addFiles(list: File[]) {
		if (disabled) return;
		add(list);
		field?.focus();
	}

	function choose(input: HTMLInputElement | null) {
		sourcesOpen = false;
		input?.click();
	}

	function remove(key: number) {
		const target = files.find((f) => f.key === key);
		if (target) release(target);
		files = files.filter((f) => f.key !== key);
		field?.focus();
	}

	function clearFiles() {
		for (const f of files) if (f.preview) URL.revokeObjectURL(f.preview);
		files = [];
	}

	async function submit() {
		if (!canSend) return;
		const text = draft;
		const attachments = ready;
		sending = true;
		draft = '';
		oncaret?.(null);
		const ok = await onsend(text, attachments);
		sending = false;
		if (ok) clearFiles();
		else draft = text;
		field?.focus();
	}

	function onkeydown(event: KeyboardEvent) {
		if (event.key !== 'Enter' || event.shiftKey || event.isComposing) return;
		event.preventDefault();
		void submit();
	}

	function onpaste(event: ClipboardEvent) {
		const pasted = Array.from(event.clipboardData?.files ?? []);
		if (!pasted.length) return;
		event.preventDefault();
		add(pasted);
	}

	function onpick(event: Event) {
		const input = event.currentTarget as HTMLInputElement;
		add(Array.from(input.files ?? []));
		input.value = '';
	}

	function status(f: Pending): string {
		if (f.status === 'error' && f.error) return errorText(f.error);
		if (f.status === 'uploading') return `${Math.round(f.progress * 100)}%`;
		return formatSize(f.size);
	}
</script>

<form
	class="flex flex-col gap-2 rounded-[31px] border border-fc-border bg-fc-bg p-2 focus-within:border-fc-ring"
	aria-label="Message composer"
	onsubmit={(event) => {
		event.preventDefault();
		void submit();
	}}
>
	{#if files.length}
		<ul class="flex flex-wrap gap-2 px-1 pt-1" aria-label="Attached files">
			{#each files as f (f.key)}
				<li
					class={[
						'flex max-w-full min-w-0 items-center gap-2 rounded-fc-lg py-1 pl-1 sm:max-w-64',
						f.status === 'error' ? 'bg-fc-danger/10' : 'bg-fc-surface'
					]}
				>
					{#if f.preview}
						<img src={f.preview} alt="" class="size-10 shrink-0 rounded-fc-md object-cover" />
					{:else}
						<span
							class="inline-flex size-10 shrink-0 items-center justify-center rounded-fc-md bg-fc-bg text-fc-fg-muted"
						>
							<Icon icon={CHAT_ICONS.file} size={20} />
						</span>
					{/if}
					<span class="flex min-w-0 flex-1 flex-col">
						<span class="truncate text-fc-sm text-fc-fg">{f.name}</span>
						<span
							class={[
								'text-fc-xs',
								f.status === 'error' ? 'text-fc-danger' : 'text-fc-fg-muted'
							]}
							role={f.status === 'error' ? 'alert' : undefined}
						>
							{status(f)}
						</span>
						{#if f.status === 'uploading'}
							<span
								class="mt-1 h-1 w-full overflow-hidden rounded-fc-pill bg-fc-border"
								role="progressbar"
								aria-label={`Uploading ${f.name}`}
								aria-valuemin={0}
								aria-valuemax={100}
								aria-valuenow={Math.round(f.progress * 100)}
							>
								<span
									class="block h-full bg-fc-accent transition-[width] motion-reduce:transition-none"
									style:width={`${Math.round(f.progress * 100)}%`}
								></span>
							</span>
						{/if}
					</span>
					<button
						type="button"
						aria-label={`Remove ${f.name}`}
						class="inline-flex size-11 shrink-0 items-center justify-center rounded-fc-pill text-fc-fg-muted hover:bg-fc-bg hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
						onclick={() => remove(f.key)}
					>
						<Icon icon={icons.close} size={16} />
					</button>
				</li>
			{/each}
		</ul>
	{/if}
	{#if native && sourcesOpen}
		<div class="flex flex-wrap gap-2 px-1 pt-1" role="menu" aria-label="Attach from">
			{#each [{ label: 'Take a photo', input: camera }, { label: 'Photo library', input: library }, { label: 'File', input: picker }] as source (source.label)}
				<button
					type="button"
					role="menuitem"
					class="inline-flex min-h-11 items-center rounded-fc-pill bg-fc-surface px-4 text-fc-sm text-fc-fg hover:bg-fc-component focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
					onclick={() => choose(source.input)}
				>
					{source.label}
				</button>
			{/each}
		</div>
	{/if}
	<div class="flex items-end gap-2">
		<input
			bind:this={picker}
			type="file"
			multiple
			accept={ATTACHMENT_ACCEPT}
			class="hidden"
			tabindex="-1"
			aria-hidden="true"
			onchange={onpick}
		/>
		{#if native}
			<input
				bind:this={camera}
				type="file"
				accept="image/*"
				capture="environment"
				class="hidden"
				tabindex="-1"
				aria-hidden="true"
				onchange={onpick}
			/>
			<input
				bind:this={library}
				type="file"
				multiple
				accept="image/*"
				class="hidden"
				tabindex="-1"
				aria-hidden="true"
				onchange={onpick}
			/>
		{/if}
		<button
			type="button"
			aria-label="Attach files"
			{disabled}
			class="inline-flex size-11 shrink-0 items-center justify-center rounded-fc-pill text-fc-fg-muted hover:bg-fc-surface hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring disabled:cursor-not-allowed disabled:opacity-40"
			aria-haspopup={native ? 'menu' : undefined}
			aria-expanded={native ? sourcesOpen : undefined}
			onclick={() => (native ? (sourcesOpen = !sourcesOpen) : picker?.click())}
		>
			<Icon icon={icons.plus} size={20} />
		</button>
		<textarea
			bind:this={field}
			bind:value={draft}
			{onkeydown}
			{onpaste}
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
				title={hint || undefined}
				disabled={!canSend}
				class="inline-flex size-11 shrink-0 items-center justify-center rounded-fc-pill bg-fc-accent text-fc-accent-fg hover:opacity-90 focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring disabled:cursor-not-allowed disabled:opacity-40"
			>
				<Icon icon={CHAT_ICONS.send} size={18} />
			</button>
		{/if}
	</div>
	{#if hint}
		<p class="sr-only" role="status">{hint}</p>
	{/if}
</form>
