<script lang="ts">
	import { Icon, Modal } from '@facile/muse';
	import { attachmentObjectUrl, attachmentUrl } from '$lib/backend';
	import { formatSize, isImageType, type Attachment } from '$lib/chat/attachments';
	import { CHAT_ICONS } from '$lib/icons';
	import { isApp } from '$lib/platform';

	let { attachments }: { attachments: Attachment[] } = $props();

	const app = isApp();
	let sources = $state<Record<string, string>>({});
	let viewing = $state<Attachment | null>(null);
	let open = $state(false);

	const images = $derived(attachments.filter((a) => isImageType(a.media_type)));
	const others = $derived(attachments.filter((a) => !isImageType(a.media_type)));

	$effect(() => {
		if (!app) return;
		const made: string[] = [];
		let live = true;
		for (const a of images) {
			attachmentObjectUrl(a.id).then(
				(url) => {
					if (!live) {
						URL.revokeObjectURL(url);
						return;
					}
					made.push(url);
					sources = { ...sources, [a.id]: url };
				},
				() => undefined
			);
		}
		return () => {
			live = false;
			for (const url of made) URL.revokeObjectURL(url);
		};
	});

	function src(a: Attachment): string | undefined {
		return app ? sources[a.id] : attachmentUrl(a.id);
	}

	function show(a: Attachment) {
		viewing = a;
		open = true;
	}

	async function download(event: MouseEvent, a: Attachment) {
		if (!app) return;
		event.preventDefault();
		const url = await attachmentObjectUrl(a.id);
		const link = document.createElement('a');
		link.href = url;
		link.download = a.name;
		link.click();
		setTimeout(() => URL.revokeObjectURL(url), 60_000);
	}
</script>

{#if attachments.length}
	<ul class="flex max-w-[85%] flex-wrap justify-end gap-2 self-end" aria-label="Attachments">
		{#each images as a (a.id)}
			<li>
				<button
					type="button"
					class="block size-24 overflow-hidden rounded-fc-lg bg-fc-surface focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
					aria-label={`View ${a.name}`}
					onclick={() => show(a)}
				>
					{#if src(a)}
						<img src={src(a)} alt={a.name} loading="lazy" class="size-full object-cover" />
					{/if}
				</button>
			</li>
		{/each}
		{#each others as a (a.id)}
			<li>
				<a
					href={attachmentUrl(a.id)}
					target="_blank"
					rel="noopener"
					download={a.name}
					onclick={(event) => download(event, a)}
					class="flex min-h-11 max-w-64 items-center gap-2 rounded-fc-lg bg-fc-surface py-1 pr-3 pl-1 hover:text-fc-fg focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-fc-ring"
				>
					<span
						class="inline-flex size-10 shrink-0 items-center justify-center rounded-fc-md bg-fc-bg text-fc-fg-muted"
					>
						<Icon icon={CHAT_ICONS.file} size={20} />
					</span>
					<span class="flex min-w-0 flex-col">
						<span class="truncate text-fc-sm text-fc-fg">{a.name}</span>
						<span class="text-fc-xs text-fc-fg-muted">{formatSize(a.size)}</span>
					</span>
				</a>
			</li>
		{/each}
	</ul>
	{#if images.length}
		<Modal bind:open title={viewing?.name} size="lg" showClose onClose={() => (viewing = null)}>
			{#if viewing && src(viewing)}
				<img
					src={src(viewing)}
					alt={viewing.name}
					class="max-h-[75dvh] w-full rounded-fc-md object-contain"
				/>
			{/if}
		</Modal>
	{/if}
{/if}
