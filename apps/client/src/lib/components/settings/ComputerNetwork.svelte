<script lang="ts">
	import { Button, Field, Icon, IconButton, Input, RadioGroup, SettingsRow, icons } from '@facile/muse';
	import { computer } from '$lib/chat/computer.svelte';
	import { MAX_HOSTS, checkHost, sameHosts, type HostProblem } from '$lib/computer-hosts';
	import type { ComputerNetwork } from '$lib/api';
	import { t } from '$lib/i18n/index.svelte';

	let draftMode = $state<ComputerNetwork | null>(null);
	let draftHosts = $state<string[] | null>(null);
	let entry = $state('');
	let problem = $state<string | null>(null);

	const savedMode = $derived<ComputerNetwork>(computer.value?.network ?? 'open');
	const savedHosts = $derived(computer.value?.net_allow ?? []);
	const mode = $derived(draftMode ?? savedMode);
	const hosts = $derived(draftHosts ?? savedHosts);
	const dirty = $derived(
		mode !== savedMode || (mode === 'allowlist' && !sameHosts(hosts, savedHosts))
	);

	const options = $derived([
		{ value: 'open', label: t('computer.network.mode.open') },
		{ value: 'allowlist', label: t('computer.network.mode.allowlist') },
		{ value: 'offline', label: t('computer.network.mode.offline') }
	]);

	const explanation = $derived(
		mode === 'allowlist'
			? t('computer.network.allowlist')
			: mode === 'offline'
				? t('computer.network.offline')
				: t('computer.network.open')
	);

	function message(reason: HostProblem): string {
		return reason === 'empty' ? '' : t(`computer.network.error.${reason}`);
	}

	function pick(value: string) {
		draftMode = value as ComputerNetwork;
		if (draftMode === savedMode) draftMode = null;
	}

	function add() {
		const check = checkHost(entry);
		if (!check.ok) {
			problem = message(check.problem);
			return;
		}
		if (hosts.includes(check.host)) {
			problem = t('computer.network.error.duplicate');
			return;
		}
		if (hosts.length >= MAX_HOSTS) {
			problem = t('computer.network.error.limit', { max: MAX_HOSTS });
			return;
		}
		draftHosts = [...hosts, check.host];
		entry = '';
		problem = null;
	}

	function remove(host: string) {
		draftHosts = hosts.filter((h) => h !== host);
		problem = null;
	}

	function onkeydown(event: KeyboardEvent) {
		if (event.key !== 'Enter') return;
		event.preventDefault();
		add();
	}

	function oninput() {
		const check = entry.trim() ? checkHost(entry) : null;
		const live = check && !check.ok && (check.problem === 'ip' || check.problem === 'chars');
		problem = live ? message(check.problem) : null;
	}

	async function apply() {
		await computer.setNetwork(mode, mode === 'allowlist' ? hosts : undefined);
		if (!computer.error) {
			draftMode = null;
			draftHosts = null;
		}
	}
</script>

<SettingsRow stacked label={t('computer.network.label')} description={t('computer.network.description')}>
	<div class="flex flex-col gap-4">
		<RadioGroup
			orientation="horizontal"
			aria-label={t('computer.network.label')}
			{options}
			value={mode}
			disabled={computer.busy || !computer.value}
			onchange={(event) => pick((event.target as HTMLInputElement).value)}
		/>
		<p class="text-fc-sm text-fc-fg-muted">{explanation}</p>

		{#if mode === 'allowlist'}
			<div class="flex flex-col gap-3">
				<Field
					label={t('computer.network.hosts.label')}
					helper={t('computer.network.hosts.helper')}
					error={problem ?? undefined}
				>
					<div class="flex gap-2">
						<Input
							bind:value={entry}
							class="min-w-0 flex-1"
							placeholder={t('computer.network.hosts.placeholder')}
							autocapitalize="none"
							autocorrect="off"
							spellcheck={false}
							inputmode="url"
							{onkeydown}
							{oninput}
						/>
						<Button variant="secondary" disabled={!entry.trim()} onclick={add}>
							{t('computer.network.hosts.add')}
						</Button>
					</div>
				</Field>

				{#if hosts.length === 0}
					<p class="text-fc-xs text-fc-fg-muted">{t('computer.network.hosts.none')}</p>
				{:else}
					<ul class="flex flex-col gap-2">
						{#each hosts as host (host)}
							<li class="flex items-center justify-between gap-2 rounded-fc-md bg-fc-surface/40 py-1 pr-1 pl-3">
								<span class="min-w-0 break-all font-mono text-fc-sm text-fc-fg">{host}</span>
								<IconButton
									variant="ghost"
									aria-label={t('computer.network.hosts.remove', { host })}
									onclick={() => remove(host)}
								>
									<Icon icon={icons.close} size={16} />
								</IconButton>
							</li>
						{/each}
					</ul>
					<span class="text-fc-xs text-fc-fg-muted">
						{t('computer.network.hosts.count', { count: hosts.length, max: MAX_HOSTS })}
					</span>
				{/if}
			</div>
		{/if}

		<div>
			<Button disabled={!dirty || computer.busy || !computer.value} onclick={apply}>
				{t('computer.network.apply')}
			</Button>
		</div>
	</div>
</SettingsRow>
