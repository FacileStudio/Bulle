<script lang="ts">
	import { Renderer } from '@openuidev/svelte-lang';
	import type { ActionEvent } from '@openuidev/lang-core';
	import { Skeleton } from '@facile/muse';
	import { toOpenUIAction, type OpenUIAction } from './actions';
	import { createInspector } from './inspect';
	import { library } from './library';

	let {
		source,
		streaming = false,
		onAction
	}: {
		source: string;
		streaming?: boolean;
		onAction?: (action: OpenUIAction) => void;
	} = $props();

	const inspect = createInspector(library);
	const verdict = $derived(inspect(source, streaming));

	function handle(event: ActionEvent) {
		const action = toOpenUIAction(event);
		if (action) onAction?.(action);
	}
</script>

{#if verdict.status === 'ready'}
	<Renderer response={source} {library} isStreaming={streaming} onAction={handle} />
{:else if verdict.status === 'pending'}
	<Skeleton class="h-16 w-full" />
{:else}
	<pre
		class="overflow-x-auto rounded-fc-md bg-fc-component p-4 text-fc-xs text-fc-fg"
		data-openui-fallback={verdict.reason}><code>{source}</code></pre>
{/if}
