<script lang="ts">
	import { onMount } from 'svelte';
	import { Alert, Button, Icon, Section, SettingsRow, StatusDot, Switch, icons } from '@facile/muse';
	import { computer } from '$lib/chat/computer.svelte';
	import { relativeTime } from '$lib/chat/format';
	import { CHAT_ICONS } from '$lib/icons';
	import { assistantName } from '$lib/session.svelte';
	import { t } from '$lib/i18n/index.svelte';

	type Tone = 'neutral' | 'success' | 'warning' | 'danger';

	const name = $derived(assistantName());
	const status = $derived(computer.value?.status);
	const isBooting = $derived(computer.booting || status === 'creating');

	const statusMeta = $derived<{ label: string; tone: Tone; desc: string }>(
		isBooting
			? { label: t('computer.status.starting'), tone: 'warning', desc: t('computer.status.startingDesc', { name }) }
			: status === 'running'
				? { label: t('computer.status.running'), tone: 'success', desc: t('computer.status.runningDesc') }
				: status === 'error'
					? { label: t('computer.status.error'), tone: 'danger', desc: t('computer.status.errorDesc') }
					: { label: t('computer.status.asleep'), tone: 'neutral', desc: t('computer.status.asleepDesc', { name }) }
	);

	const offline = $derived(computer.value?.network === 'offline');
	let networkAttempt = $state(0);

	async function toggleNetwork() {
		try {
			await computer.setNetwork(offline ? 'open' : 'offline');
		} finally {
			networkAttempt++;
		}
	}

	const specs = $derived([
		{ label: t('computer.spec.cpu'), value: t('computer.spec.cpuValue', { count: computer.value?.cpus || 2 }) },
		{
			label: t('computer.spec.memory'),
			value: t('computer.spec.memoryValue', { count: Math.round((computer.value?.memory_mb || 4096) / 1024) })
		},
		{ label: t('computer.spec.disk'), value: t('computer.spec.diskValue', { count: computer.value?.disk_gb || 20 }) },
		{
			label: t('computer.spec.lastActivity'),
			value: computer.value?.last_active_at ? relativeTime(computer.value.last_active_at) : t('computer.spec.never')
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
	title={t('computer.title')}
	description={t('computer.description', { name })}
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
					variant="secondary"
					size="sm"
					icon={CHAT_ICONS.stop}
					disabled={computer.busy}
					onclick={() => computer.stop()}
				>
					{t('computer.stop')}
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

		<SettingsRow label={t('computer.network.label')} description={t('computer.network.description')}>
			{#key networkAttempt}
				<Switch
					class="min-h-11 min-w-11"
					checked={offline}
					disabled={computer.busy || !computer.value}
					aria-label={t('computer.network.label')}
					onchange={toggleNetwork}
				/>
			{/key}
		</SettingsRow>
	</div>
</Section>
