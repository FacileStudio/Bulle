<script lang="ts">
	import { StatusDot } from '@facile/muse';
	import type { SessionStatus } from '$lib/api';

	let {
		status,
		showLabel = false,
		class: className = ''
	}: { status: SessionStatus | undefined; showLabel?: boolean; class?: string } = $props();

	const label = $derived(status === 'waiting' ? 'Needs approval' : 'Working');
</script>

{#if status === 'running' || status === 'waiting'}
	<span class="inline-flex items-center {className}" role={showLabel ? 'status' : 'img'} aria-label={label} title={label}>
		<StatusDot
			tone={status === 'waiting' ? 'warning' : 'accent'}
			pulse
			label={showLabel ? label : undefined}
			aria-hidden="true"
		/>
	</span>
{/if}
