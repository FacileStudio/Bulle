<script lang="ts">
	import { Button, StatusDot } from '@facile/muse';
	import type { ComputerStatus } from '$lib/api';
	import { computer } from '$lib/chat/computer.svelte';

	type Tone = 'neutral' | 'success' | 'warning' | 'danger';

	const view: Record<ComputerStatus, { label: string; tone: Tone }> = {
		absent: { label: 'Asleep', tone: 'neutral' },
		creating: { label: 'Booting', tone: 'warning' },
		running: { label: 'Running', tone: 'success' },
		stopped: { label: 'Asleep', tone: 'neutral' },
		error: { label: 'Error', tone: 'danger' }
	};

	const status = $derived(computer.value?.status);
	const isBooting = $derived(computer.booting || status === 'creating');
	const current = $derived(status ? view[status] : undefined);
	const canStop = $derived(status === 'running');
</script>

{#if isBooting}
	<div class="flex shrink-0 items-center gap-2">
		<StatusDot tone="warning" pulse label="Starting..." />
	</div>
{:else if current}
	<div class="flex shrink-0 items-center gap-2">
		<StatusDot tone={current.tone} pulse={status === 'creating'} label={current.label} />
		{#if canStop}
			<Button size="sm" variant="ghost" disabled={computer.busy} onclick={() => computer.stop()}>
				Stop
			</Button>
		{/if}
	</div>
{/if}
