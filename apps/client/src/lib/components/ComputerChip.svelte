<script lang="ts">
	import { Button, StatusDot } from '@facile/muse';
	import type { ComputerStatus } from '$lib/api';
	import { computer } from '$lib/chat/computer.svelte';

	type Tone = 'neutral' | 'success' | 'warning' | 'danger';

	const view: Record<ComputerStatus, { label: string; tone: Tone }> = {
		absent: { label: 'No computer', tone: 'neutral' },
		creating: { label: 'Booting', tone: 'warning' },
		running: { label: 'Running', tone: 'success' },
		stopped: { label: 'Stopped', tone: 'neutral' },
		error: { label: 'Error', tone: 'danger' }
	};

	const status = $derived(computer.value?.status);
	const current = $derived(status ? view[status] : undefined);
	const canStart = $derived(status === 'absent' || status === 'stopped' || status === 'error');
	const canStop = $derived(status === 'running');
</script>

{#if current}
	<div class="flex shrink-0 items-center gap-2">
		<StatusDot tone={current.tone} pulse={status === 'creating'} label={current.label} />
		{#if canStart}
			<Button size="sm" variant="outline" disabled={computer.busy} onclick={() => computer.start()}>
				{status === 'absent' ? 'Create' : 'Start'}
			</Button>
		{:else if canStop}
			<Button size="sm" variant="ghost" disabled={computer.busy} onclick={() => computer.stop()}>
				Stop
			</Button>
		{/if}
	</div>
{/if}
