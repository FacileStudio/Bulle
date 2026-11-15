import type { UsageBudget } from '$lib/backend';

export const ALLOWANCE_WARN_AT = 0.8;

export type Allowance = {
	percent: number;
	tone: 'ok' | 'warn' | 'over';
};

/**
 * Reads where a user stands against their monthly AI budget. It answers null
 * when no meter applies: no budget (limit 0) or an exempt user.
 */
export function allowance(budget: UsageBudget | undefined): Allowance | null {
	if (!budget || budget.exempt || budget.limit_usd <= 0) return null;
	const ratio = budget.spent_usd / budget.limit_usd;
	const tone = ratio >= 1 ? 'over' : ratio >= ALLOWANCE_WARN_AT ? 'warn' : 'ok';
	return { percent: Math.min(100, Math.round(ratio * 100)), tone };
}
