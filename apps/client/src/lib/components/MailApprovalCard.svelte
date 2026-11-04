<script lang="ts">
	import { untrack } from 'svelte';
	import { Badge, Button, Field, Icon, Input, Textarea, icons } from '@facile/muse';
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

	function decide(as: 'send' | 'draft') {
		ondecide(approval.callId, { approve: true, as, args: edits });
	}
</script>

<div
	class="flex flex-col gap-3 rounded-fc-lg border border-fc-border bg-fc-component p-4"
	role="group"
	aria-label="Review the email before it is sent"
>
	<div class="flex flex-wrap items-center gap-2 text-fc-sm">
		<Icon icon={icons.shield} size={16} class="text-fc-warning" />
		<span class="font-medium">Send an email</span>
		<span class="text-fc-fg-muted">review it, edit it, then send</span>
		{#if settled}
			<span class="ml-auto">
				<Badge tone={approval.state === 'approved' ? 'success' : 'neutral'}>
					{approval.state === 'approved' ? 'Done' : 'Cancelled'}
				</Badge>
			</span>
		{/if}
	</div>

	<dl class="grid grid-cols-[auto_1fr] items-center gap-x-3 gap-y-1 text-fc-sm">
		<dt class="text-fc-fg-muted">From</dt>
		<dd class="break-all">{mail.from || 'your default mailbox'}</dd>
		{#if mail.threaded}
			<dt class="text-fc-fg-muted">Thread</dt>
			<dd><Badge tone="neutral">Reply in the existing conversation</Badge></dd>
		{/if}
	</dl>

	<fieldset class="flex flex-col gap-3" disabled={settled || busy}>
		<Field label="To" for="{id}-to">
			<Input id="{id}-to" bind:value={draft.to} autocomplete="off" />
		</Field>
		{#if showCopies}
			<Field label="Cc" for="{id}-cc">
				<Input id="{id}-cc" bind:value={draft.cc} autocomplete="off" />
			</Field>
			<Field label="Bcc" for="{id}-bcc">
				<Input id="{id}-bcc" bind:value={draft.bcc} autocomplete="off" />
			</Field>
		{:else if !settled}
			<button
				type="button"
				class="self-start text-fc-xs text-fc-fg-muted hover:text-fc-fg"
				onclick={() => (showCopies = true)}
			>
				Add Cc or Bcc
			</button>
		{/if}
		<Field label="Subject" for="{id}-subject">
			<Input id="{id}-subject" bind:value={draft.subject} />
		</Field>
		<Field label="Message" for="{id}-body">
			<Textarea id="{id}-body" bind:value={draft.body} rows={8} />
		</Field>
	</fieldset>

	{#if !settled}
		<div class="flex flex-wrap gap-2">
			<Button size="lg" disabled={busy || !draft.to.trim()} onclick={() => decide('send')}>
				Send
			</Button>
			<Button size="lg" variant="outline" disabled={busy} onclick={() => decide('draft')}>
				Save as draft
			</Button>
			<Button
				size="lg"
				variant="ghost"
				disabled={busy}
				onclick={() => ondecide(approval.callId, { approve: false })}
			>
				Cancel
			</Button>
		</div>
	{/if}
</div>
