<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, Button, Icon, Section, StatusDot, icons } from '@facile/muse';
	import { computer } from '$lib/chat/computer.svelte';
	import { relativeTime } from '$lib/chat/format';
	import { CHAT_ICONS } from '$lib/icons';
	import { session } from '$lib/session.svelte';

	type Tone = 'neutral' | 'success' | 'warning' | 'danger';

	const assistantName = $derived(session.user?.assistant?.name || 'Bulle');
	const status = $derived(computer.value?.status);
	const isBooting = $derived(computer.booting || status === 'creating');

	const statusMeta = $derived<{ label: string; tone: Tone; desc: string }>(
		isBooting
			? { label: 'Starting', tone: 'warning', desc: `${assistantName} is starting its machine for a task.` }
			: status === 'running'
				? { label: 'Running', tone: 'success', desc: `Online. It stops on its own after a while without use.` }
				: status === 'error'
					? { label: 'Error', tone: 'danger', desc: 'The machine hit a problem. The next task that needs it retries.' }
					: { label: 'Asleep', tone: 'neutral', desc: `${assistantName} starts it on its own when a task needs it.` }
	);

	const specs = $derived([
		{ label: 'CPU', value: `${computer.value?.cpus || 2} vCPU` },
		{
			label: 'Memory',
			value: `${Math.round((computer.value?.memory_mb || 4096) / 1024)} GB`
		},
		{ label: 'Disk', value: `${computer.value?.disk_gb || 20} GB` },
		{
			label: 'Last activity',
			value: computer.value?.last_active_at ? relativeTime(computer.value.last_active_at) : 'Never'
		}
	]);

	onMount(() => {
		void computer.watch();
		return () => computer.unwatch();
	});
</script>

<Section
	card
	id="computer"
	title="Computer"
	description={`The private Linux machine where ${assistantName} runs code and shell commands.`}
>
	<div class="flex flex-col gap-5">
		{#if computer.error}
			<Alert tone="danger">{computer.error}</Alert>
		{/if}
		{#if computer.value?.error}
			<Alert tone="danger">{computer.value.error}</Alert>
		{/if}

		<div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
			<div class="flex items-center gap-3">
				<div
					class="flex size-10 shrink-0 items-center justify-center rounded-fc-md bg-fc-surface text-fc-fg"
				>
					<Icon icon={icons.monitor} size={20} />
				</div>
				<div class="flex flex-col">
					<StatusDot tone={statusMeta.tone} pulse={isBooting} label={statusMeta.label} />
					<span class="text-fc-xs text-fc-fg-muted">{statusMeta.desc}</span>
				</div>
			</div>
			{#if status === 'running' && !isBooting}
				<Button
					variant="outline"
					size="sm"
					icon={CHAT_ICONS.stop}
					disabled={computer.busy}
					onclick={() => computer.stop()}
				>
					Stop now
				</Button>
			{/if}
		</div>

		<dl class="grid grid-cols-2 gap-3 sm:grid-cols-4">
			{#each specs as spec (spec.label)}
				<div class="flex flex-col gap-1 rounded-fc-md bg-fc-surface/40 p-3">
					<dt class="text-fc-xs text-fc-fg-muted">{spec.label}</dt>
					<dd class="font-mono text-fc-sm font-semibold text-fc-fg">{spec.value}</dd>
				</div>
			{/each}
		</dl>
	</div>
</Section>
