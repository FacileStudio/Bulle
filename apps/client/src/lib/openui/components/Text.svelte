<script lang="ts">
	import type { ComponentRenderProps } from '@openuidev/svelte-lang';

	type Tone = 'neutral' | 'muted' | 'accent' | 'success' | 'warning' | 'danger';

	let {
		props
	}: ComponentRenderProps<{
		content?: string;
		tone?: Tone | null;
	}> = $props();

	const toneClass = $derived(
		props.tone === 'muted'
			? 'text-fc-fg-muted'
			: props.tone === 'accent'
				? 'text-fc-accent'
				: props.tone === 'success'
					? 'text-fc-success'
					: props.tone === 'warning'
						? 'text-fc-warning'
						: props.tone === 'danger'
							? 'text-fc-danger'
							: 'text-fc-fg'
	);
</script>

<p class="text-fc-sm leading-relaxed {toneClass}">{props.content ?? ''}</p>
