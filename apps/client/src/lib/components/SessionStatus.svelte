<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { StatusDot } from '@facile/muse';
	import type { SessionStatus } from '$lib/api';

	let {
		status,
		showLabel = false,
		class: className = ''
	}: { status: SessionStatus | undefined; showLabel?: boolean; class?: string } = $props();

	const label = $derived(status === 'waiting' ? t('session.needsApproval') : t('session.working'));
</script>

{#if status === 'running' || status === 'waiting'}
	<span class="inline-flex items-center {className}" role={showLabel ? 'status' : 'img'} aria-label={label} title={label}>
		<StatusDot
			tone={status === 'waiting' ? 'accent' : 'warning'}
			pulse={status === 'running'}
			label={showLabel ? label : undefined}
			aria-hidden="true"
		/>
	</span>
{/if}
