<script lang="ts">
	import { dev } from '$app/environment';
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { Alert, Button, EmptyState, Icon, IconButton, PlanLimit, Skeleton, Spinner, icons, type PlanLimitError } from '@facile/muse';
	import { Blob } from '@facile/blob';
	import { t } from '$lib/i18n/index.svelte';
	import { api, type PageContext } from '$lib/api';
	import { ApiError } from '$lib/backend';
	import type { Attachment } from '$lib/chat/attachments';
	import { chat } from '$lib/chat/chat.svelte';
	import Composer from '$lib/components/Composer.svelte';
	import Message from '$lib/components/Message.svelte';
	import { actionMessage } from '$lib/openui/actions';
	import { needsOnboarding } from '$lib/onboarding/steps';
	import { SUBSCRIBE_URL } from '$lib/plan';
	import { assistantName, loadSession, session } from '$lib/session.svelte';
	import { CHAT_ICONS } from '$lib/icons';
	import {
		askFrom,
		conversationKey,
		isAllowedOrigin,
		pageFrom,
		parentOrigin,
		quoteSelection,
		samePage,
		type EmbedAsk
	} from '$lib/embed';

	let current = $state<PageContext | null>(pageFrom((name) => page.url.searchParams.get(name)));
	let sessionId = $state<string | null>(null);
	let resuming = $state(true);
	let error = $state<string | null>(null);
	let startLimit = $state<PlanLimitError | null>(null);
	let composer = $state<Composer | null>(null);
	let scroller = $state<HTMLElement | null>(null);
	let pinned = true;
	let noted: PageContext | null = null;
	let parent: string | null = null;

	const name = $derived(assistantName());
	const view = $derived(chat.state);
	const ready = $derived(session.status === 'signed-in' && !needsOnboarding(session.user));
	const unpaid = $derived<PlanLimitError | null>(
		session.user?.plan === 'free' ? { message: t('embed.paid'), upgrade_url: SUBSCRIBE_URL } : null
	);
	const limit = $derived(startLimit ?? chat.planLimit ?? unpaid);

	$effect(() => {
		void view;
		if (pinned) scroller?.scrollTo({ top: scroller.scrollHeight });
	});

	function onscroll() {
		if (!scroller) return;
		pinned = scroller.scrollHeight - scroller.scrollTop - scroller.clientHeight < 80;
	}

	function stored(key: string): string | null {
		try {
			return localStorage.getItem(key);
		} catch {
			return null;
		}
	}

	function store(key: string, id: string | null) {
		try {
			if (id) localStorage.setItem(key, id);
			else localStorage.removeItem(key);
		} catch {
			return;
		}
	}

	async function resume() {
		resuming = true;
		chat.close();
		sessionId = null;
		noted = null;
		const key = session.user && current ? conversationKey(session.user.id, current) : null;
		const id = key ? stored(key) : null;
		if (key && id) {
			const found = await api()
				.getSession(id)
				.then(() => true)
				.catch((err) => !(err instanceof ApiError && err.status === 404));
			if (found) {
				sessionId = id;
				await chat.open(id);
			} else {
				store(key, null);
			}
		}
		resuming = false;
	}

	async function conversation(): Promise<string> {
		if (sessionId) return sessionId;
		if (!session.user || !current) throw new Error(t('embed.noPage'));
		const created = await api().createSession(current.title || undefined);
		store(conversationKey(session.user.id, current), created.id);
		sessionId = created.id;
		await chat.open(created.id);
		return created.id;
	}

	async function send(text: string, attachments: Attachment[] = []): Promise<boolean> {
		if (!current || unpaid) return false;
		error = null;
		startLimit = null;
		try {
			await conversation();
		} catch (err) {
			if (err instanceof ApiError && err.planLimit) startLimit = err.planLimit;
			else error = err instanceof Error ? err.message : t('chat.startFailed');
			return false;
		}
		const page = samePage(noted, current) ? null : current;
		const sent = await chat.send(text, attachments, page);
		if (sent) noted = current;
		return sent;
	}

	async function ask(request: EmbedAsk) {
		if (request.context && (request.context.app !== current?.app || request.context.id !== current?.id)) {
			current = request.context;
			await resume();
		} else if (request.context) {
			current = request.context;
		}
		const text = quoteSelection(t('embed.aboutSelection'), request.text);
		if (!ready || unpaid || view.running || resuming || !(await send(text))) composer?.prefill(text);
	}

	function post(message: unknown) {
		if (parent) window.parent.postMessage(message, parent);
	}

	async function signIn() {
		await loadSession();
		if (ready && current) await resume();
	}

	onMount(() => {
		parent = parentOrigin(window.location, document.referrer, dev);
		function onmessage(event: MessageEvent) {
			if (event.source !== window.parent || !isAllowedOrigin(event.origin, dev)) return;
			const request = askFrom(event.data);
			if (!request) return;
			parent = event.origin;
			void ask(request);
		}
		function onfocus() {
			if (!ready) void signIn();
		}
		window.addEventListener('message', onmessage);
		window.addEventListener('focus', onfocus);
		void (async () => {
			await signIn();
			resuming = false;
			post({ type: 'bulle:ready', assistant: { name: assistantName() } });
		})();
		return () => {
			window.removeEventListener('message', onmessage);
			window.removeEventListener('focus', onfocus);
			chat.close();
		};
	});
</script>

<svelte:head>
	<title>{name} - Bulle</title>
	<base target="_blank" />
</svelte:head>

<div class="flex h-dvh min-h-0 flex-col bg-fc-bg text-fc-fg">
	<header class="flex h-14 shrink-0 items-center gap-2 px-3">
		<Blob {name} size={28} animate={view.running ? 'always' : undefined} class="size-7 shrink-0" />
		<h1 class="min-w-0 flex-1 truncate text-fc-md font-medium">{name}</h1>
		<IconButton variant="ghost" aria-label={t('embed.close')} onclick={() => post({ type: 'bulle:close' })}>
			<Icon icon={icons.close} size={18} />
		</IconButton>
	</header>

	{#if session.status === 'loading'}
		<div class="flex flex-1 items-center justify-center"><Spinner /></div>
	{:else if session.status === 'signed-out'}
		<EmptyState bare icon={CHAT_ICONS.chat} title={t('embed.signedOutTitle')} description={t('embed.signedOutBody')} class="flex-1">
			<Button href="/login?next=%2Fsessions" target="_blank" rel="external noopener" icon={icons.external}>{t('embed.signIn')}</Button>
		</EmptyState>
	{:else if session.status === 'unreachable'}
		<EmptyState bare icon={icons.error} title={t('embed.unreachable')} class="flex-1">
			<Button icon={icons.refresh} onclick={signIn}>{t('common.tryAgain')}</Button>
		</EmptyState>
	{:else if !ready}
		<EmptyState bare icon={CHAT_ICONS.chat} title={t('embed.setupTitle')} class="flex-1">
			<Button href="/welcome" target="_blank" rel="external noopener" icon={icons.external}>{t('embed.openBulle')}</Button>
		</EmptyState>
	{:else if !current}
		<EmptyState bare icon={CHAT_ICONS.chat} title={t('embed.noPage')} class="flex-1" />
	{:else}
		<div bind:this={scroller} {onscroll} class="min-h-0 flex-1 overflow-y-auto overscroll-contain">
			<div class="flex flex-col gap-4 px-3 py-4" role="log" aria-live="off" aria-label={t('chat.conversation')}>
				{#if resuming || chat.loading}
					<Skeleton class="h-10 w-2/3 self-end" />
					<Skeleton class="h-20 w-full" />
				{:else if chat.loadError}
					<EmptyState bare icon={icons.error} title={t('chat.loadFailed')} description={chat.loadError}>
						<Button icon={icons.refresh} onclick={resume}>{t('common.tryAgain')}</Button>
					</EmptyState>
				{:else if !view.items.length}
					<EmptyState
						bare
						icon={CHAT_ICONS.chat}
						title={t('embed.emptyTitle')}
						description={t('embed.emptyBody', { name })}
					/>
				{:else}
					{#each view.items as item, index (item.id)}
						<Message
							{item}
							ondecide={(callId, decision) => chat.decide(callId, decision)}
							ondecideall={(callIds, decision) => chat.decideAll(callIds, decision)}
							onaction={(action) => send(actionMessage(action))}
							answered={index < view.items.length - 1}
							busy={view.running}
							onanswer={(text) => send(text)}
						/>
					{/each}
				{/if}
			</div>
		</div>

		<div class="flex shrink-0 flex-col gap-2 px-3 pb-3">
			<PlanLimit error={limit} />
			{#if error}
				<Alert tone="danger">{error}</Alert>
			{/if}
			{#if view.error && !view.items.some((i) => i.error)}
				<Alert tone="danger" title={t('chat.errorTitle')}>
					<div class="flex items-center justify-between gap-3">
						<span>{view.error}</span>
						<Button size="sm" variant="ghost" onclick={() => chat.dismissError()}>{t('common.dismiss')}</Button>
					</div>
				</Alert>
			{/if}
			<Composer
				bind:this={composer}
				running={view.running}
				disabled={!!unpaid || resuming || chat.loading}
				placeholder={t('embed.placeholder')}
				onsend={send}
				onstop={() => chat.stop()}
			/>
		</div>
	{/if}
</div>
