<script lang="ts">
	import { StatCard } from '@facile/muse';
	import type { ComponentRenderProps } from '@openuidev/svelte-lang';

	let {
		props
	}: ComponentRenderProps<{
		label?: string;
		value?: string | number;
		delta?: string | null;
	}> = $props();
</script>

<StatCard label={props.label ?? ''} value={props.value ?? ''} delta={props.delta ?? undefined} />
