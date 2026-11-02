# Plan: fullscreen assistant onboarding

Status: implemented on `feat/fullscreen-onboarding`, 2026-11-02.

## Problem

`OnboardingModal.svelte` put the whole setup (name, four archetypes, five traits) in one dense modal over the chat page. The user chose a personality from two-line descriptions, never heard the result, and saw the app behind the overlay. The save button read "Save and Start Coding" although the assistant also handles mail, calendar and tasks.

## Research

- Dot (New Computer): one question per screen, a tone-setting welcome, a "one more thing" finale, and a moment where the assistant reacts to what you wrote. The review also flags what to avoid: steps the product does not need yet (pronouns, birth date) and permission asks before the user knows what the product does. https://diklein.com/writing/on-the-design-of-dot
- Replika: identity first, then look, then the name, then a "Meet <name>" button as the last action. The whole flow is five taps. https://help.replika.com/hc/en-us/articles/115001095892-How-do-I-begin-Replika
- Nielsen Norman Group on wizards: show the steps and the current one, enforce the order, label buttons by what happens next, let people leave and resume, and do not hide the app behind a modal. https://www.nngroup.com/articles/wizards

## Design

A fullscreen route, `/welcome`, outside the sidebar layout. Five screens, one idea each.

| Step | Screen | Primary button |
|---|---|---|
| 1 | Welcome. Large live avatar, one line on what the assistant does (mail, calendar, tasks, code). | Give it a name |
| 2 | Name. Input, dice, avatar redraws as you type and its eyes follow the caret. Skip keeps "Bulle". | Choose a personality |
| 3 | Personality. Four cards. Each shows the reply that personality would give to "What's on my plate today?". | Fine-tune the style |
| 4 | Traits. Five chips. The same sample reply changes as you toggle them. Skip keeps "Concise". | Meet {name} |
| 5 | Meet {name}. Avatar scales in. | Say hello |

Rules:

- The sample replies are static strings composed on the client. There is no model call during onboarding.
- Progress is a labelled segmented bar ("Step 2 of 6") with `role="progressbar"`. Back is on every step after the first. Skip exists only where a default is safe.
- The draft lives in `sessionStorage` and survives a reload. Every storage access sits in try/catch.
- Only step 5 saves, with one `PUT /api/assistant`. The API and the `AssistantProfile` shape do not change. Archetype and trait ids stay the same.
- Focus moves to the step heading on every change. A polite live region announces the step. Enter continues from the name field and from the heading.
- GSAP drives the transitions. With `prefers-reduced-motion` the slide becomes a short crossfade.
- Colors are muse tokens only. The ambient glow takes its tint from the chosen archetype.
- The account page reuses the name field, the personality cards and the trait chips, so there is one implementation.

## Gate

`(app)/+layout.svelte` redirects a signed-in user with no assistant name to `/welcome` and shows a spinner instead of the chat while it does. `/welcome` sends signed-out users to `/login` and users who already have an assistant to `/sessions`.

## Where a "Connect your apps" step would slot in

The `feat/suite-delegation` branch adds `SuiteConnections.svelte`. After both branches merge, add a step between Traits and Meet: put `{ id: 'apps', label: 'Apps', next: 'Meet {name}' }` in `lib/onboarding/steps.ts`, change the Traits label to "Connect your apps", and render the component in a new `AppsStep.svelte`. Only the `STEPS` list and the step switch in `Onboarding.svelte` change.
