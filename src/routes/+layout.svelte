<script lang="ts">
	import favicon from "$lib/assets/favicon.ico";
	import favicon16 from "$lib/assets/favicon-16x16.png";
	import favicon32 from "$lib/assets/favicon-32x32.png";
	import favicon96 from "$lib/assets/favicon-96x96.png";
	import appleIcon57 from "$lib/assets/apple-icon-57x57.png";
	import appleIcon60 from "$lib/assets/apple-icon-60x60.png";
	import appleIcon72 from "$lib/assets/apple-icon-72x72.png";
	import appleIcon76 from "$lib/assets/apple-icon-76x76.png";
	import appleIcon114 from "$lib/assets/apple-icon-114x114.png";
	import appleIcon120 from "$lib/assets/apple-icon-120x120.png";
	import appleIcon144 from "$lib/assets/apple-icon-144x144.png";
	import appleIcon152 from "$lib/assets/apple-icon-152x152.png";
	import appleIcon180 from "$lib/assets/apple-icon-180x180.png";
	import msIcon144 from "$lib/assets/ms-icon-144x144.png";

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
	<link rel="icon" type="image/png" sizes="16x16" href={favicon16} />
	<link rel="icon" type="image/png" sizes="32x32" href={favicon32} />
	<link rel="icon" type="image/png" sizes="96x96" href={favicon96} />
	<link rel="apple-touch-icon" sizes="57x57" href={appleIcon57} />
	<link rel="apple-touch-icon" sizes="60x60" href={appleIcon60} />
	<link rel="apple-touch-icon" sizes="72x72" href={appleIcon72} />
	<link rel="apple-touch-icon" sizes="76x76" href={appleIcon76} />
	<link rel="apple-touch-icon" sizes="114x114" href={appleIcon114} />
	<link rel="apple-touch-icon" sizes="120x120" href={appleIcon120} />
	<link rel="apple-touch-icon" sizes="144x144" href={appleIcon144} />
	<link rel="apple-touch-icon" sizes="152x152" href={appleIcon152} />
	<link rel="apple-touch-icon" sizes="180x180" href={appleIcon180} />
	<meta name="msapplication-TileImage" content={msIcon144} />
</svelte:head>

{@render children()}
