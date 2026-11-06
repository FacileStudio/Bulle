<script lang="ts">
	import { Button, Checkbox, Icon, IconButton, Input, Select, icons, toast } from '@facile/muse';
	import {
		GOAL_CATEGORIES,
		createGoal,
		deleteGoal,
		replaceById,
		sortGoals,
		updateGoal,
		type Goal,
		type GoalCategory
	} from '$lib/ideas';

	type Props = { spaceId: string; goals: Goal[]; onchange: (goals: Goal[]) => void };

	let { spaceId, goals, onchange }: Props = $props();

	const MAX_OPEN = 20;

	let open = $state(false);
	let title = $state('');
	let category = $state<string>('productivity');
	let adding = $state(false);

	const sorted = $derived(sortGoals(goals));
	const openCount = $derived(goals.filter((goal) => !goal.done).length);
	const full = $derived(openCount >= MAX_OPEN);

	function label(value: string): string {
		return value.charAt(0).toUpperCase() + value.slice(1);
	}

	async function add(event: SubmitEvent) {
		event.preventDefault();
		const text = title.trim();
		if (!text || full) return;
		adding = true;
		try {
			const goal = await createGoal(spaceId, text, category as GoalCategory);
			onchange([...goals, goal]);
			title = '';
		} catch (err) {
			toast.danger(err instanceof Error ? err.message : 'Could not add the goal');
		} finally {
			adding = false;
		}
	}

	async function toggle(goal: Goal) {
		try {
			onchange(replaceById(goals, await updateGoal(goal.id, { done: !goal.done })));
		} catch {
			toast.danger('Could not update the goal');
		}
	}

	async function remove(goal: Goal) {
		try {
			await deleteGoal(goal.id);
			onchange(goals.filter((item) => item.id !== goal.id));
		} catch {
			toast.danger('Could not delete the goal');
		}
	}
</script>

<section class="flex flex-col gap-3" aria-labelledby="goals-heading">
	<Button
		variant="ghost"
		class="-ml-2 h-11 self-start"
		iconRight={open ? icons.chevronUp : icons.chevronDown}
		aria-expanded={open}
		aria-controls="goals-body"
		onclick={() => (open = !open)}
	>
		<span id="goals-heading">Goals ({openCount})</span>
	</Button>

	{#if open}
		<div id="goals-body" class="flex flex-col gap-3">
			<p class="text-fc-sm text-fc-fg-muted">Bulle reads these when it thinks of ideas.</p>

			{#if sorted.length > 0}
				<ul class="flex flex-col gap-1">
					{#each sorted as goal (goal.id)}
						<li class="flex items-center gap-2 rounded-fc-md bg-fc-surface pl-3">
							<Checkbox
								class="min-h-11 min-w-0 flex-1 {goal.done ? 'text-fc-fg-muted line-through' : ''}"
								checked={goal.done}
								label={`${goal.title} · ${label(goal.category)}`}
								onchange={() => toggle(goal)}
							/>
							<IconButton variant="danger" aria-label={`Delete goal ${goal.title}`} onclick={() => remove(goal)}>
								<Icon icon={icons.remove} size={16} />
							</IconButton>
						</li>
					{/each}
				</ul>
			{/if}

			<form class="flex flex-col gap-2 sm:flex-row" onsubmit={add}>
				<Select bind:value={category} aria-label="Goal category" class="sm:w-44" disabled={full}>
					{#each GOAL_CATEGORIES as value (value)}
						<option {value}>{label(value)}</option>
					{/each}
				</Select>
				<Input
					bind:value={title}
					aria-label="New goal"
					placeholder={full ? 'You have 20 open goals' : 'Add a goal'}
					maxlength={120}
					disabled={full}
					class="min-w-0 flex-1"
				/>
				<Button type="submit" class="h-11" icon={icons.plus} disabled={adding || full || !title.trim()}>
					Add
				</Button>
			</form>
		</div>
	{/if}
</section>
