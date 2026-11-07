<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { goto } from '$app/navigation';
	import { Alert, PlanLimit, type PlanLimitError } from '@facile/muse';
	import { Blob, gaze } from '@facile/blob';
	import { api } from '$lib/api';
	import { ApiError } from '$lib/backend';
	import type { Attachment } from '$lib/chat/attachments';
	import { sessions } from '$lib/chat/sessions.svelte';
	import { session } from '$lib/session.svelte';
	import Composer from '$lib/components/Composer.svelte';
	import FileDropZone from '$lib/components/FileDropZone.svelte';
	import { caretGazeTarget, type CaretCoordinates } from '$lib/caret';

	let error = $state<string | null>(null);
	let planLimit = $state<PlanLimitError | null>(null);
	let avatarEl = $state<HTMLElement | null>(null);
	const assistantName = $derived(session.user?.assistant?.name?.trim() || 'bulle');
	const eyes = gaze({ travel: 6, target: 'pointer' });

	$effect(() => {
		function onpointermove() {
			eyes.lookAt('pointer');
		}
		window.addEventListener('pointermove', onpointermove, { passive: true });
		return () => {
			window.removeEventListener('pointermove', onpointermove);
		};
	});

	function oncaret(coords: CaretCoordinates | null) {
		if (coords && avatarEl) {
			eyes.lookAt(caretGazeTarget(avatarEl.getBoundingClientRect(), coords));
			return;
		}
		eyes.lookAt('pointer');
	}

	async function start(text: string, attachments: Attachment[]): Promise<boolean> {
		error = null;
		planLimit = null;
		try {
			const id = await sessions.create();
			await api().sendMessage(
				id,
				text.trim(),
				attachments.map((a) => a.id)
			);
			await goto(`/s/${id}`);
			return true;
		} catch (err) {
			if (err instanceof ApiError && err.planLimit) {
				planLimit = err.planLimit;
				return false;
			}
			error = err instanceof Error ? err.message : t('chat.startFailed');
			return false;
		}
	}

	let composer = $state<Composer | null>(null);
</script>

<svelte:head>
	<title>{t('chat.newTitle')}</title>
</svelte:head>

<FileDropZone class="h-full" disabled={false} ondrop={(files) => composer?.addFiles(files)}>
	<div class="mx-auto flex h-full w-full max-w-fc-md flex-col justify-center gap-6 px-4 pb-4">
		<div class="flex flex-col items-center gap-4 py-8 text-center select-none">
			<div
				bind:this={avatarEl}
				class="relative flex items-center justify-center cursor-pointer transition-transform duration-200 hover:scale-105"
			>
				<Blob
					{@attach eyes}
					name={assistantName}
					size={128}
					animate="always"
					class="drop-shadow-sm transition-transform duration-200"
				/>
			</div>
			<div class="flex flex-col gap-1 max-w-fc-sm">
				<h2 class="text-fc-md font-medium text-fc-fg">{t('chat.emptyTitle')}</h2>
				<p class="text-fc-sm text-fc-fg-muted">
					{t('chat.emptyBody')}
				</p>
			</div>
		</div>
		<PlanLimit error={planLimit} />
		{#if error}
			<Alert tone="danger">{error}</Alert>
		{/if}
		<Composer bind:this={composer} onsend={start} {oncaret} placeholder={t('chat.taskPlaceholder')} />
	</div>
</FileDropZone>
