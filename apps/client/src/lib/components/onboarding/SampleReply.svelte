<script lang="ts">
	import { SAMPLE_PROMPT } from '$lib/onboarding/sample';

	let {
		lines,
		compact = false,
		live = false
	}: { lines: readonly string[]; compact?: boolean; live?: boolean } = $props();
</script>

<div class="flex w-full flex-col gap-2 text-left">
	<p class="self-end rounded-fc-lg bg-fc-component px-3 py-1.5 text-fc-xs text-fc-fg-muted">
		{SAMPLE_PROMPT}
	</p>
	<div
		class="flex flex-col gap-1.5 rounded-fc-lg border border-fc-border bg-fc-bg/80 text-fc-fg {compact
			? 'p-3 text-fc-xs'
			: 'p-4 text-fc-sm'}"
		aria-live={live ? 'polite' : undefined}
		aria-atomic={live ? 'true' : undefined}
	>
		{#each lines as line, index (index)}
			<p>{line}</p>
		{/each}
	</div>
</div>
