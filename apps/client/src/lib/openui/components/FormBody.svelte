<script lang="ts">
	import { untrack, type Snippet } from 'svelte';
	import { Button } from '@facile/muse';
	import { getIsStreaming, getTriggerAction, setFormNameContext } from '@openuidev/svelte-lang';
	import { FORM_ACTION } from '../actions';

	let {
		name,
		submitLabel,
		fields,
		renderNode
	}: {
		name?: string;
		submitLabel?: string;
		fields?: unknown;
		renderNode: Snippet<[unknown]>;
	} = $props();

	const trigger = getTriggerAction();
	const streaming = getIsStreaming();

	setFormNameContext(untrack(() => name));

	function submit(event: SubmitEvent) {
		event.preventDefault();
		if (!name) return;
		trigger(submitLabel ?? name, name, { type: FORM_ACTION, params: { name } });
	}
</script>

<form class="flex flex-col gap-4" onsubmit={submit}>
	{@render renderNode(fields)}
	<div>
		<Button type="submit" disabled={streaming() || !name}>{submitLabel ?? 'Submit'}</Button>
	</div>
</form>
