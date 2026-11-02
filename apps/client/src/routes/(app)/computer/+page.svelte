<script lang="ts">
	import { onMount } from 'svelte';
	import {
		Alert,
		Button,
		Icon,
		PageHeader,
		Section,
		Spinner,
		Stack,
		StatusDot,
		icons
	} from '@facile/muse';
	import { computer } from '$lib/chat/computer.svelte';
	import { relativeTime } from '$lib/chat/format';
	import { CHAT_ICONS } from '$lib/icons';
	import { session } from '$lib/session.svelte';

	type Tone = 'neutral' | 'success' | 'warning' | 'danger';

	const assistantName = $derived(session.user?.assistant?.name || 'Bulle');
	const status = $derived(computer.value?.status);
	const isBooting = $derived(computer.busy || computer.booting || status === 'creating');

	const statusMeta = $derived<{ label: string; tone: Tone; desc: string }>(
		isBooting
			? computer.waking && status !== 'creating'
				? { label: 'Waking', tone: 'warning', desc: 'The virtual machine is waking up. This takes a few seconds.' }
				: { label: 'Booting', tone: 'warning', desc: 'The virtual machine is being created. The first run takes a few minutes.' }
			: status === 'running'
				? { label: 'Running', tone: 'success', desc: `${assistantName}'s virtual machine is online and ready for tasks.` }
				: status === 'stopped'
					? { label: 'Stopped', tone: 'neutral', desc: 'The virtual machine is currently stopped to conserve resources.' }
					: status === 'error'
						? { label: 'Error', tone: 'danger', desc: 'The machine encountered an issue during execution.' }
						: { label: 'Absent', tone: 'neutral', desc: 'No virtual machine allocated yet. It will be created on start.' }
	);

	onMount(() => {
		void computer.watch();
		return () => computer.unwatch();
	});
</script>

<svelte:head>
	<title>Computer - Bulle</title>
</svelte:head>

<Stack gap="section">
	<PageHeader
		title="Computer"
		description={`The dedicated Linux environment for ${assistantName}.`}
	/>

	<Section
		card
		title="Machine Status"
		description="Live state and power controls for the virtual workspace."
	>
		<div class="flex flex-col gap-5">
			{#if computer.error}
				<Alert tone="danger">{computer.error}</Alert>
			{/if}
			{#if computer.value?.error}
				<Alert tone="danger" title="System Error">{computer.value.error}</Alert>
			{/if}

			<div class="flex flex-col gap-4 rounded-fc-md border border-fc-border bg-fc-surface/40 p-4 sm:flex-row sm:items-center sm:justify-between">
				<div class="flex items-center gap-3">
					<div class="flex size-10 shrink-0 items-center justify-center rounded-fc-md bg-fc-surface border border-fc-border text-fc-fg">
						<Icon icon={icons.monitor} size={20} />
					</div>
					<div class="flex flex-col">
						<div class="flex items-center gap-2">
							<span class="font-medium text-fc-md text-fc-fg">Virtual Machine</span>
							<StatusDot tone={statusMeta.tone} pulse={isBooting} label={statusMeta.label} />
						</div>
						<span class="text-fc-xs text-fc-fg-muted">{statusMeta.desc}</span>
					</div>
				</div>

				<div class="flex items-center gap-2">
					{#if isBooting}
						<div class="flex items-center gap-2 text-fc-xs text-fc-fg-muted">
							<Spinner size="sm" />
							<span>Starting...</span>
						</div>
					{:else if status === 'running'}
						<Button
							variant="outline"
							size="sm"
							icon={icons.refresh}
							disabled={computer.busy}
							onclick={() => computer.restart()}
						>
							Restart
						</Button>
						<Button
							variant="ghost"
							size="sm"
							icon={CHAT_ICONS.stop}
							disabled={computer.busy}
							onclick={() => computer.stop()}
						>
							Stop
						</Button>
					{:else}
						<Button
							variant="primary"
							size="sm"
							icon={icons.bolt}
							disabled={computer.busy}
							onclick={() => computer.start()}
						>
							{status === 'absent' ? 'Provision Machine' : 'Start Computer'}
						</Button>
					{/if}
				</div>
			</div>
		</div>
	</Section>

	<Section
		card
		title="Hardware & Resources"
		description="Allocated virtual CPU cores, memory, and disk capacity."
	>
		<div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
			<div class="flex flex-col gap-1 rounded-fc-md border border-fc-border bg-fc-surface/40 p-3">
				<span class="text-fc-xs text-fc-fg-muted">Virtual CPUs</span>
				<span class="font-mono text-fc-md font-semibold text-fc-fg">
					{computer.value?.cpus || 2} vCPU
				</span>
			</div>

			<div class="flex flex-col gap-1 rounded-fc-md border border-fc-border bg-fc-surface/40 p-3">
				<span class="text-fc-xs text-fc-fg-muted">Memory (RAM)</span>
				<span class="font-mono text-fc-md font-semibold text-fc-fg">
					{computer.value?.memory_mb ? `${computer.value.memory_mb} MB (${Math.round(computer.value.memory_mb / 1024)} GB)` : '4096 MB (4 GB)'}
				</span>
			</div>

			<div class="flex flex-col gap-1 rounded-fc-md border border-fc-border bg-fc-surface/40 p-3">
				<span class="text-fc-xs text-fc-fg-muted">Disk Storage</span>
				<span class="font-mono text-fc-md font-semibold text-fc-fg">
					{computer.value?.disk_gb ? `${computer.value.disk_gb} GB SSD` : '20 GB SSD'}
				</span>
			</div>

			<div class="flex flex-col gap-1 rounded-fc-md border border-fc-border bg-fc-surface/40 p-3">
				<span class="text-fc-xs text-fc-fg-muted">Operating System</span>
				<span class="font-mono text-fc-md font-semibold text-fc-fg">Linux x86_64</span>
			</div>

			<div class="flex flex-col gap-1 rounded-fc-md border border-fc-border bg-fc-surface/40 p-3">
				<span class="text-fc-xs text-fc-fg-muted">Workspace Directory</span>
				<span class="font-mono text-fc-md font-semibold text-fc-fg">/workspace</span>
			</div>

			<div class="flex flex-col gap-1 rounded-fc-md border border-fc-border bg-fc-surface/40 p-3">
				<span class="text-fc-xs text-fc-fg-muted">Last Activity</span>
				<span class="text-fc-md font-semibold text-fc-fg">
					{computer.value?.last_active_at ? relativeTime(computer.value.last_active_at) : 'None recorded'}
				</span>
			</div>
		</div>
	</Section>

	<Section
		card
		title="Capabilities & Tools"
		description={`Native tools available to ${assistantName} within this workspace.`}
	>
		<div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
			<div class="flex items-start gap-3 rounded-fc-md border border-fc-border bg-fc-surface/30 p-3">
				<div class="mt-0.5 text-fc-fg-muted">
					<Icon icon={CHAT_ICONS.terminal} size={18} />
				</div>
				<div class="flex flex-col gap-0.5">
					<span class="font-medium text-fc-sm text-fc-fg">Shell Execution</span>
					<span class="text-fc-xs text-fc-fg-muted">
						Runs commands inside the workspace container with git, build tools, and package managers.
					</span>
				</div>
			</div>

			<div class="flex items-start gap-3 rounded-fc-md border border-fc-border bg-fc-surface/30 p-3">
				<div class="mt-0.5 text-fc-fg-muted">
					<Icon icon={CHAT_ICONS.tool} size={18} />
				</div>
				<div class="flex flex-col gap-0.5">
					<span class="font-medium text-fc-sm text-fc-fg">File Manipulation</span>
					<span class="text-fc-xs text-fc-fg-muted">
						Inspects code, creates source files, and applies targeted diffs with verification.
					</span>
				</div>
			</div>

			<div class="flex items-start gap-3 rounded-fc-md border border-fc-border bg-fc-surface/30 p-3">
				<div class="mt-0.5 text-fc-fg-muted">
					<Icon icon={icons.globe} size={18} />
				</div>
				<div class="flex flex-col gap-0.5">
					<span class="font-medium text-fc-sm text-fc-fg">Web Inspection</span>
					<span class="text-fc-xs text-fc-fg-muted">
						Fetches documentation and reference content securely while treating pages as data.
					</span>
				</div>
			</div>

			<div class="flex items-start gap-3 rounded-fc-md border border-fc-border bg-fc-surface/30 p-3">
				<div class="mt-0.5 text-fc-fg-muted">
					<Icon icon={icons.shield} size={18} />
				</div>
				<div class="flex flex-col gap-0.5">
					<span class="font-medium text-fc-sm text-fc-fg">Human Approval Gate</span>
					<span class="text-fc-xs text-fc-fg-muted">
						High-risk commands and external mutations pause for your explicit consent.
					</span>
				</div>
			</div>
		</div>
	</Section>
</Stack>
