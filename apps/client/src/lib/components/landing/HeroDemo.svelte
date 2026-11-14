<script lang="ts">
	import { tick } from 'svelte';
	import { Blob } from '@facile/blob';
	import { Bar, Cursor, Stage, Window, cursorAt, type Act } from '@facile/mise-en-scene';
	import { t } from '$lib/i18n/index.svelte';
	import HeroThread, { type HeroScene } from './HeroThread.svelte';

	const HISTORY = ['72%', '58%', '80%', '46%', '64%', '70%', '40%'];

	const prompt = $derived(t('landing.demo.prompt'));
	const replyWords = $derived(t('landing.demo.reply').split(' ').length);

	let scene = $state<HeroScene>(blank());
	let cursor = $state(cursorAt());
	let approveEl = $state<HTMLElement>();

	function blank(): HeroScene {
		return { draft: '', sent: false, assistant: false, steps: ['hidden', 'hidden', 'hidden'], approval: 'hidden', words: 0 };
	}

	function reset() {
		scene = blank();
		cursor.visible = false;
		cursor.down = false;
	}

	function poster() {
		scene = { draft: '', sent: true, assistant: true, steps: ['ok', 'ok', 'ok'], approval: 'approved', words: replyWords };
	}

	async function runSteps(act: Act) {
		for (let i = 0; i < scene.steps.length; i++) {
			await act.wait(250);
			scene.steps[i] = 'running';
			await act.wait(850);
			scene.steps[i] = 'ok';
		}
	}

	async function approve(act: Act) {
		scene.approval = 'pending';
		await tick();
		const rest = act.point(approveEl, 2.4, 1.8);
		await act.move(cursor, rest, 900);
		cursor.visible = true;
		await act.wait(500);
		await act.move(cursor, act.point(approveEl, 0.6, 0.6), 1000);
		await act.click(cursor, () => (scene.approval = 'pressed'));
		scene.approval = 'approved';
		await act.move(cursor, rest, 700);
		cursor.visible = false;
	}

	async function script(act: Act) {
		await act.wait(700);
		await act.type(prompt, (v) => (scene.draft = v));
		await act.wait(350);
		scene.sent = true;
		scene.draft = '';
		await act.wait(500);
		scene.assistant = true;
		await act.wait(600);
		await runSteps(act);
		await act.wait(400);
		await approve(act);
		await act.count(replyWords, (v) => (scene.words = v));
		await act.wait(2600);
		reset();
		await act.wait(700);
	}
</script>

{#snippet sidebar()}
	<div class="flex flex-col gap-1 px-3 pt-4">
		<div class="mb-3 flex items-center gap-2 px-2">
			<Blob name="Bulle" size={22} />
			<span class="text-fc-sm font-bold text-fc-fg">Bulle</span>
		</div>
		{#each HISTORY as w, i (i)}
			<div class="rounded-fc-md px-2 py-2.5" class:bg-fc-surface={i === 0}><Bar {w} /></div>
		{/each}
	</div>
{/snippet}

<Stage
	width={960}
	height={640}
	compact={{ below: 640, width: 358, height: 620 }}
	{script}
	{reset}
	{poster}
	label={t('landing.hero.body')}
>
	{#snippet children({ compact })}
		{#if compact}
			<Window style="top:16px;left:16px;width:326px;height:588px">
				<HeroThread {scene} compact bind:approveEl />
			</Window>
		{:else}
			<Window style="top:28px;left:72px;width:816px;height:576px" {sidebar}>
				<HeroThread {scene} compact={false} bind:approveEl />
			</Window>
		{/if}
		<Cursor {cursor} />
	{/snippet}
</Stage>
