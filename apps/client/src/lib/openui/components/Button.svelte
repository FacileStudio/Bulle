<script lang="ts">
	import { Button as MuseButton } from '@facile/muse';
	import { getIsStreaming, getTriggerAction } from '@openuidev/svelte-lang';
	import type { ComponentRenderProps } from '@openuidev/svelte-lang';
	import { BUTTON_ACTION } from '../actions';

	type Variant = 'primary' | 'outline' | 'ghost' | 'danger';

	let {
		props
	}: ComponentRenderProps<{ label?: string; name?: string; variant?: Variant | null }> = $props();

	const trigger = getTriggerAction();
	const streaming = getIsStreaming();
</script>

<MuseButton
	variant={props.variant ?? 'primary'}
	disabled={streaming() || !props.name}
	onclick={() =>
		trigger(props.label ?? '', undefined, {
			type: BUTTON_ACTION,
			params: { name: props.name }
		})}
>
	{props.label ?? ''}
</MuseButton>
