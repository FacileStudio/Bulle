<script lang="ts">
	let { tint }: { tint: readonly [string, string] } = $props();
</script>

<div class="ambient" aria-hidden="true" style:--tint-a={tint[0]} style:--tint-b={tint[1]}>
	<span class="orb orb-a"></span>
	<span class="orb orb-b"></span>
	<span class="orb orb-c"></span>
</div>

<style>
	.ambient {
		position: absolute;
		inset: 0;
		overflow: hidden;
		pointer-events: none;
	}

	.orb {
		position: absolute;
		aspect-ratio: 1;
		border-radius: 9999px;
		filter: blur(90px);
		opacity: 0.3;
		transition: background-color 1.2s var(--ease-fc);
		animation: drift 26s var(--ease-fc) infinite alternate;
	}

	.orb-a {
		width: min(60vw, 520px);
		top: -12%;
		left: -10%;
		background-color: var(--tint-a);
	}

	.orb-b {
		width: min(55vw, 460px);
		bottom: -14%;
		right: -8%;
		background-color: var(--tint-b);
		animation-duration: 32s;
		animation-delay: -8s;
	}

	.orb-c {
		width: min(40vw, 320px);
		top: 38%;
		left: 56%;
		opacity: 0.16;
		background-color: var(--tint-a);
		animation-delay: -14s;
	}

	@keyframes drift {
		from {
			transform: translate3d(0, 0, 0) scale(1);
		}
		to {
			transform: translate3d(6vw, 5vh, 0) scale(1.15);
		}
	}
</style>
