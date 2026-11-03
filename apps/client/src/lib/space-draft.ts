import type { Space, SpaceInput } from '$lib/spaces.svelte';

export type SpaceDraft = {
	name: string;
	courrier_account: string;
	agenda_calendar: string;
	opus_workspace: string;
};

/** Turns a space into editable text fields, showing an unset id as empty. */
export function draftFor(space: Space): SpaceDraft {
	return {
		name: space.name,
		courrier_account: space.courrier_account > 0 ? String(space.courrier_account) : '',
		agenda_calendar: space.agenda_calendar > 0 ? String(space.agenda_calendar) : '',
		opus_workspace: space.opus_workspace
	};
}

function toId(text: string): number {
	const value = Number(String(text).trim());
	return Number.isInteger(value) && value > 0 ? value : 0;
}

/** Turns the edited fields back into an update, where an empty or invalid id clears the default. */
export function inputFromDraft(draft: SpaceDraft): SpaceInput {
	return {
		name: draft.name.trim(),
		courrier_account: toId(draft.courrier_account),
		agenda_calendar: toId(draft.agenda_calendar),
		opus_workspace: draft.opus_workspace.trim()
	};
}
