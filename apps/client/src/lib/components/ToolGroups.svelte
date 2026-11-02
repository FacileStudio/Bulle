<script lang="ts">
	import { Badge, Icon, Spinner, icons } from '@facile/muse';
	import type { ToolCall } from '$lib/chat/reducer';
	import { failedCount, groupTools, prettyArgs, runningCount, summarizeArgs } from '$lib/chat/tools';
	import type { ToolKind } from '$lib/chat/tools';
	import { formatDuration } from '$lib/chat/format';
	import { CHAT_ICONS } from '$lib/icons';

	let { tools }: { tools: ToolCall[] } = $props();

	let expanded = $state<Record<string, boolean>>({});

	const groups = $derived(groupTools(tools));

	const kindIcon: Record<ToolKind, string> = {
		commands: CHAT_ICONS.terminal,
		files: icons.folder,
		web: icons.globe,
		other: CHAT_ICONS.tool
	};

	function toggle(callId: string) {
		expanded[callId] = !expanded[callId];
	}
</script>

{#if groups.length}
	<div class="flex flex-col gap-3">
		{#each groups as group (group.kind)}
			{@const running = runningCount(group)}
			{@const failed = failedCount(group)}
			<section class="overflow-hidden rounded-fc-lg border border-fc-border" aria-label={group.label}>
				<header class="flex items-center gap-2 bg-fc-surface px-3 py-2 text-fc-sm">
					<Icon icon={kindIcon[group.kind]} size={16} class="text-fc-fg-muted" />
					<span class="font-medium">{group.label}</span>
					<span class="text-fc-fg-muted">{group.calls.length}</span>
					<span class="ml-auto flex items-center gap-2">
						{#if running}
							<Spinner size="sm" />
							<span class="text-fc-xs text-fc-fg-muted">{running} running</span>
						{/if}
						{#if failed}
							<Badge tone="danger">{failed} failed</Badge>
						{/if}
					</span>
				</header>
				<ul class="divide-y divide-fc-border">
					{#each group.calls as call (call.callId)}
						<li>
							<button
								type="button"
								class="flex min-h-11 w-full items-center gap-3 px-3 py-2 text-left text-fc-sm hover:bg-fc-surface focus-visible:outline-2 focus-visible:-outline-offset-2 focus-visible:outline-fc-ring"
								aria-expanded={!!expanded[call.callId]}
								onclick={() => toggle(call.callId)}
							>
								<span class="flex size-4 shrink-0 items-center justify-center">
									{#if call.status === 'running'}
										<Spinner size="sm" />
									{:else if call.status === 'ok'}
										<Icon icon={icons.check} size={16} class="text-fc-success" />
									{:else}
										<Icon icon={icons.error} size={16} class="text-fc-danger" />
									{/if}
								</span>
								<span class="shrink-0 font-medium">{call.name}</span>
								<span class="min-w-0 flex-1 truncate font-fc-mono text-fc-xs text-fc-fg-muted">
									{summarizeArgs(call.args)}
								</span>
								{#if call.durationMs !== undefined}
									<span class="shrink-0 text-fc-xs text-fc-fg-muted">
										{formatDuration(call.durationMs)}
									</span>
								{/if}
							</button>
							{#if expanded[call.callId]}
								<div class="flex flex-col gap-2 bg-fc-component px-3 py-3 text-fc-xs">
									{#if prettyArgs(call.args)}
										<pre
											class="overflow-x-auto font-fc-mono whitespace-pre-wrap text-fc-fg">{prettyArgs(
												call.args
											)}</pre>
									{/if}
									{#if call.outputPreview}
										<pre
											class="max-h-60 overflow-auto font-fc-mono whitespace-pre-wrap text-fc-fg-muted">{call.outputPreview}</pre>
									{/if}
								</div>
							{/if}
						</li>
					{/each}
				</ul>
			</section>
		{/each}
	</div>
{/if}
