<script lang="ts">
	import { Alert } from '@facile/muse';
	import type { ComponentRenderProps } from '@openuidev/svelte-lang';

	type Tone = 'neutral' | 'info' | 'success' | 'warning' | 'danger';

	let {
		props
	}: ComponentRenderProps<{ text?: string; tone?: Tone | null; title?: string | null }> = $props();
</script>

<Alert tone={props.tone ?? 'info'} title={props.title ?? undefined}>{props.text ?? ''}</Alert>
