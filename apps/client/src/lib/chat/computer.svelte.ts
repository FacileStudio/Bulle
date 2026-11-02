import { api, type Computer } from '$lib/api';

const FAST_MS = 3000;
const SLOW_MS = 30000;

class ComputerStore {
	value = $state.raw<Computer | null>(null);
	error = $state<string | null>(null);
	busy = $state(false);
	booting = $state(false);
	private timer: ReturnType<typeof setTimeout> | undefined;
	private active = false;
	private stoppedByUser = false;

	async refresh(): Promise<void> {
		try {
			this.value = await api().getComputer();
			this.error = null;
			if (this.value.status === 'running') {
				this.booting = false;
			} else if (this.value.status === 'creating') {
				this.booting = true;
			} else if (this.value.status === 'error') {
				this.booting = false;
			}
		} catch (err) {
			this.error = err instanceof Error ? err.message : 'Could not read the computer status';
		}
	}

	private schedule(): void {
		if (!this.active) return;
		clearTimeout(this.timer);
		const fast = this.booting || this.value?.status === 'creating' || this.busy;
		const delay = fast ? FAST_MS : SLOW_MS;
		this.timer = setTimeout(async () => {
			await this.refresh();
			this.schedule();
		}, delay);
	}

	async watch(): Promise<void> {
		this.active = true;
		await this.refresh();
		if (!this.stoppedByUser && (this.value?.status === 'stopped' || this.value?.status === 'absent')) {
			void this.start();
		}
		this.schedule();
	}

	unwatch(): void {
		this.active = false;
		clearTimeout(this.timer);
	}

	async start(): Promise<void> {
		this.stoppedByUser = false;
		this.booting = true;
		await this.act(() => api().startComputer());
	}

	async stop(): Promise<void> {
		this.stoppedByUser = true;
		this.booting = false;
		await this.act(() => api().stopComputer());
	}

	async restart(): Promise<void> {
		this.stoppedByUser = false;
		this.booting = true;
		await this.act(async () => {
			await api().stopComputer();
			await api().startComputer();
		});
	}

	private async act(call: () => Promise<void>): Promise<void> {
		this.busy = true;
		try {
			await call();
			this.error = null;
		} catch (err) {
			this.error = err instanceof Error ? err.message : 'The request failed';
			this.booting = false;
		}
		await this.refresh();
		this.busy = false;
		this.schedule();
	}
}

export const computer = new ComputerStore();
