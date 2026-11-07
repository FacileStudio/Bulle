<script lang="ts">
	import { t } from '$lib/i18n/index.svelte';
	import { untrack } from 'svelte';
	import { Button } from '@facile/muse';
	import type { Decision } from '$lib/api/types';
	import type { Approval } from '$lib/chat/reducer';
	import { draftOf, mailEdits, type MailRecord } from '$lib/chat/mail';

	let {
		approval,
		mail,
		ondecide
	}: {
		approval: Approval;
		mail: MailRecord;
		ondecide: (callId: string, decision: Decision) => void;
	} = $props();

	let draft = $state(untrack(() => draftOf(mail)));
	let showCopies = $state(untrack(() => mail.cc.length > 0 || mail.bcc.length > 0));

	const settled = $derived(approval.state === 'approved' || approval.state === 'denied');
	const busy = $derived(approval.state === 'sending');
	const edits = $derived(mailEdits(mail, draft));
	const id = $derived(`mail-${approval.callId}`);
	const line =
		'w-full min-h-11 rounded-fc-md bg-fc-field px-3 py-2 text-fc-md text-fc-fg outline-none focus-visible:outline-2 focus-visible:outline-fc-ring disabled:text-fc-fg-muted';
	const label = 'text-fc-xs text-fc-fg-muted';

	function decide(as: 'send' | 'draft') {
		ondecide(approval.callId, { approve: true, as, args: edits });
	}
</script>

{#if settled}
	<p
		class="flex items-baseline gap-2 text-fc-sm text-fc-fg-muted"
		role="group"
		aria-label={t('mail.review')}
	>
		<span>{t('mail.to', { to: draft.to || t('mail.noRecipient') })}</span>
		<span class="text-fc-xs">{approval.state === 'approved' ? t('mail.done') : t('mail.cancelled')}</span>
	</p>
{:else}
	<div
		class="flex flex-col gap-4 py-2"
		role="group"
		aria-label={t('mail.review')}
	>
		<div class="flex flex-col gap-0.5">
			<span class="text-fc-md font-medium text-fc-fg">{t('mail.send')}</span>
			<span class="text-fc-xs text-fc-fg-muted">
				{mail.threaded ? t('mail.fromThreaded', { from: mail.from || t('mail.defaultMailbox') }) : t('mail.from', { from: mail.from || t('mail.defaultMailbox') })}
			</span>
		</div>

		<fieldset class="flex flex-col gap-3" disabled={busy}>
			<label class="flex flex-col" for="{id}-to">
				<span class={label}>{t('mail.fieldTo')}</span>
				<input id="{id}-to" class={line} bind:value={draft.to} autocomplete="off" />
			</label>
			{#if showCopies}
				<label class="flex flex-col" for="{id}-cc">
					<span class={label}>{t('mail.fieldCc')}</span>
					<input id="{id}-cc" class={line} bind:value={draft.cc} autocomplete="off" />
				</label>
				<label class="flex flex-col" for="{id}-bcc">
					<span class={label}>{t('mail.fieldBcc')}</span>
					<input id="{id}-bcc" class={line} bind:value={draft.bcc} autocomplete="off" />
				</label>
			{:else}
				<button
					type="button"
					class="min-h-11 self-start px-1 text-fc-xs text-fc-fg-muted underline underline-offset-4"
					onclick={() => (showCopies = true)}
				>
					{t('mail.addCopies')}
				</button>
			{/if}
			<label class="flex flex-col" for="{id}-subject">
				<span class={label}>{t('mail.subject')}</span>
				<input id="{id}-subject" class={line} bind:value={draft.subject} />
			</label>
			<label class="flex flex-col" for="{id}-body">
				<span class={label}>{t('mail.message')}</span>
				<textarea
					id="{id}-body"
					class="{line} resize-y leading-relaxed"
					rows={12}
					style="min-height: 12rem"
					bind:value={draft.body}
				></textarea>
			</label>
		</fieldset>

		<div class="sticky bottom-0 flex flex-wrap items-center gap-x-4 gap-y-1 bg-fc-component py-2">
			<Button size="lg" disabled={busy || !draft.to.trim()} onclick={() => decide('send')}>
				{t('mail.sendButton')}
			</Button>
			<button
				type="button"
				class="min-h-11 min-w-11 px-2 text-fc-sm text-fc-fg-muted underline underline-offset-4 disabled:opacity-50"
				disabled={busy}
				onclick={() => decide('draft')}
			>
				{t('mail.saveDraft')}
			</button>
			<button
				type="button"
				class="min-h-11 min-w-11 px-2 text-fc-sm text-fc-fg-muted underline underline-offset-4 disabled:opacity-50"
				disabled={busy}
				onclick={() => ondecide(approval.callId, { approve: false })}
			>
				{t('common.cancel')}
			</button>
		</div>
	</div>
{/if}
