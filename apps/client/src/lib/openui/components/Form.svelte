<script lang="ts">
	import type { ComponentRenderProps } from '@openuidev/svelte-lang';
	import FormBody from './FormBody.svelte';

	let { props, renderNode }: ComponentRenderProps<{
		name?: string;
		submitLabel?: string;
		fields?: unknown;
	}> = $props();
</script>

{#key props.name}
	<FormBody name={props.name} submitLabel={props.submitLabel} fields={props.fields} {renderNode} />
{/key}
