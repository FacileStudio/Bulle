<script lang="ts">
	import { Button, Checkbox, Icon, IconButton, Input, Select, icons, toast } from '@facile/muse';
	import {
		GOAL_CATEGORIES,
		errorText,
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
	let category = $state<GoalCategory>('productivity');
	let adding = $state(false);
	let pending = $state<Record<string, boolean>>({});

	const sorted = $derived(sortGoals(goals));
	const openCount = $derived(goals.filter((goal) => !goal.done).length);
	const full = $derived(openCount >= MAX_OPEN);

	function label(value: string): string {
		return value.charAt(0).toUpperCase() + value.slice(1);
	}

	function setPending(id: string, value: boolean) {
		pending = { ...pending, [id]: value };
	}

	async function add(event: SubmitEvent) {
		event.preventDefault();
		const text = title.trim();
		const target = spaceId;
		if (!text || full || adding) return;
		adding = true;
		try {
			const goal = await createGoal(target, text, category);
			if (spaceId === target) onchange([...goals, goal]);
			title = '';
		} catch (err) {
			toast.danger(errorText(err, 'Could not add the goal.'));
		} finally {
			adding = false;
		}
	}

	async function toggle(event: Event, goal: Goal) {
		const input = event.currentTarget as HTMLInputElement;
		if (pending[goal.id]) {
			input.checked = goal.done;
			return;
		}
		setPending(goal.id, true);
		try {
			onchange(replaceById(goals, await updateGoal(goal.id, { done: !goal.done })));
		} catch (err) {
			input.checked = goal.done;
			toast.danger(errorText(err, 'Could not update the goal.'));
		} finally {
			setPending(goal.id, false);
		}
	}

	async function remove(goal: Goal) {
		if (pending[goal.id]) return;
		setPending(goal.id, true);
		try {
			await deleteGoal(goal.id);
			onchange(goals.filter((item) => item.id !== goal.id));
		} catch (err) {
			toast.danger(errorText(err, 'Could not delete the goal.'));
		} finally {
			setPending(goal.id, false);
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
				<ul aria-label="Goals" class="flex flex-col gap-1">
					{#each sorted as goal (goal.id)}
						<li class="flex items-center gap-2 rounded-fc-md bg-fc-surface pl-3">
							<Checkbox
								class="min-h-11 min-w-0 flex-1 py-2 wrap-anywhere {goal.done ? 'text-fc-fg-muted line-through' : ''}"
								checked={goal.done}
								disabled={pending[goal.id]}
								label={`${goal.title} · ${label(goal.category)}`}
								onchange={(event) => toggle(event, goal)}
							/>
							<IconButton
								variant="danger"
								aria-label={`Delete goal ${goal.title}`}
								disabled={pending[goal.id]}
								onclick={() => remove(goal)}
							>
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
