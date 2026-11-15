<script lang="ts">
	import { StatusDot } from '@facile/muse';
	import { t } from '$lib/i18n/index.svelte';
	import { formatDate } from '$lib/i18n/format';
	import type { JobRun } from '$lib/api';

	let { runs }: { runs: JobRun[] } = $props();

	function statusTone(status: string): 'neutral' | 'success' | 'warning' | 'danger' {
		switch (status) {
			case 'finished':
				return 'success';
			case 'running':
				return 'warning';
			case 'error':
				return 'danger';
			default:
				return 'neutral';
		}
	}

	function runStarted(val: string): string {
		try {
			return formatDate(val, {
				month: 'short',
				day: 'numeric',
				hour: '2-digit',
				minute: '2-digit',
				second: '2-digit'
			});
		} catch {
			return val;
		}
	}

	function statusLabel(status: string): string {
		switch (status) {
			case 'finished':
				return t('jobs.status.finished');
			case 'running':
				return t('jobs.status.running');
			case 'error':
				return t('jobs.status.error');
			default:
				return status;
		}
	}
</script>

<div class="flex flex-col gap-1 overflow-hidden rounded-fc-md">
	{#each runs as run (run.id)}
		<div class="flex flex-col gap-2 p-3 sm:flex-row sm:items-center sm:justify-between bg-fc-surface/30">
			<div class="flex items-center gap-3">
				<StatusDot tone={statusTone(run.status)} label={statusLabel(run.status)} />
				<div class="flex flex-col">
					<span class="font-medium text-fc-sm text-fc-fg">{statusLabel(run.status)}</span>
					<span class="text-fc-xs text-fc-fg-muted">{runStarted(run.started_at)}</span>
				</div>
			</div>

			<div class="text-fc-xs text-fc-fg-muted">
				{#if !run.finished_at}
					<span class="text-fc-warning">{t('jobs.history.inProgress')}</span>
				{/if}
			</div>
		</div>
	{/each}
</div>
