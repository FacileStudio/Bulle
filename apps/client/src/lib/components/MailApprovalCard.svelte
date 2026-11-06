<script lang="ts">
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
		'w-full min-h-11 border-0 border-b border-fc-border bg-transparent px-0 py-2 text-fc-md text-fc-fg outline-none focus-visible:border-b-2 focus-visible:border-fc-fg disabled:text-fc-fg-muted';
	const label = 'text-fc-xs text-fc-fg-muted';

	function decide(as: 'send' | 'draft') {
		ondecide(approval.callId, { approve: true, as, args: edits });
	}
</script>

{#if settled}
	<p
		class="flex items-baseline gap-2 text-fc-sm text-fc-fg-muted"
		role="group"
		aria-label="Review the email before it is sent"
	>
		<span>Email to {draft.to || 'no recipient'}</span>
		<span class="text-fc-xs">{approval.state === 'approved' ? 'Done' : 'Cancelled'}</span>
	</p>
{:else}
	<div
		class="flex flex-col gap-4 py-2"
		role="group"
		aria-label="Review the email before it is sent"
	>
		<div class="flex flex-col gap-0.5">
			<span class="text-fc-md font-medium text-fc-fg">Send an email</span>
			<span class="text-fc-xs text-fc-fg-muted">
				From {mail.from || 'your default mailbox'}{mail.threaded ? ', reply in the existing conversation' : ''}
			</span>
		</div>

		<fieldset class="flex flex-col gap-3" disabled={busy}>
			<label class="flex flex-col" for="{id}-to">
				<span class={label}>To</span>
				<input id="{id}-to" class={line} bind:value={draft.to} autocomplete="off" />
			</label>
			{#if showCopies}
				<label class="flex flex-col" for="{id}-cc">
					<span class={label}>Cc</span>
					<input id="{id}-cc" class={line} bind:value={draft.cc} autocomplete="off" />
				</label>
				<label class="flex flex-col" for="{id}-bcc">
					<span class={label}>Bcc</span>
					<input id="{id}-bcc" class={line} bind:value={draft.bcc} autocomplete="off" />
				</label>
			{:else}
				<button
					type="button"
					class="min-h-11 self-start px-1 text-fc-xs text-fc-fg-muted underline underline-offset-4"
					onclick={() => (showCopies = true)}
				>
					Add Cc or Bcc
				</button>
			{/if}
			<label class="flex flex-col" for="{id}-subject">
				<span class={label}>Subject</span>
				<input id="{id}-subject" class={line} bind:value={draft.subject} />
			</label>
			<label class="flex flex-col" for="{id}-body">
				<span class={label}>Message</span>
				<textarea
					id="{id}-body"
					class="{line} resize-y leading-relaxed"
					rows={12}
					style="min-height: 12rem"
					bind:value={draft.body}
				></textarea>
			</label>
		</fieldset>

		<div class="sticky bottom-0 flex flex-wrap items-center gap-x-4 gap-y-1 bg-fc-bg py-2">
			<Button size="lg" disabled={busy || !draft.to.trim()} onclick={() => decide('send')}>
				Send
			</Button>
			<button
				type="button"
				class="min-h-11 min-w-11 px-2 text-fc-sm text-fc-fg-muted underline underline-offset-4 disabled:opacity-50"
				disabled={busy}
				onclick={() => decide('draft')}
			>
				Save as draft
			</button>
			<button
				type="button"
				class="min-h-11 min-w-11 px-2 text-fc-sm text-fc-fg-muted underline underline-offset-4 disabled:opacity-50"
				disabled={busy}
				onclick={() => ondecide(approval.callId, { approve: false })}
			>
				Cancel
			</button>
		</div>
	</div>
{/if}
