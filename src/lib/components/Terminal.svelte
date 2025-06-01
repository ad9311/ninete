<script lang="ts">
	let statusText = 'Ready';

	function showStatus(text: string) {
		statusText = text;
		setTimeout(() => {
			statusText = 'Ready';
		}, 3000);
	}
</script>

<div
	class="mx-auto my-8 w-full max-w-3xl rounded-sm border-2 border-t-zinc-50 border-r-zinc-400 border-b-zinc-400 border-l-zinc-50 bg-zinc-200 font-sans text-sm shadow-md select-none"
>
	<!-- Windows 98 Style Title Bar -->
	<div class="flex cursor-default items-center justify-between bg-blue-700 px-2 py-1 text-white">
		<div class="flex items-center">
			<!-- Placeholder for an icon -->
			<span class="mr-2 inline-block h-4 w-4 border border-zinc-400 bg-zinc-300"></span>
			<span>My Vintage Web Page - Nostalgia Explorer</span>
		</div>
		<div class="flex space-x-1">
			<button
				class="win98-button flex h-5 w-5 items-center justify-center"
				aria-label="Minimize"
				title="Minimize"
				on:click={() => showStatus('Window Minimized (simulation)')}
			>
				<span class="-mt-1 text-xs font-black">_</span>
			</button>
			<button
				class="win98-button flex h-5 w-5 items-center justify-center"
				aria-label="Maximize"
				title="Maximize"
				on:click={() => showStatus('Window Maximized (simulation)')}
			>
				<svg viewBox="0 0 10 10" class="h-2.5 w-2.5 fill-current"
					><path d="M0 0h10v1H0zm0 1h1v8H0zm9-1h1v10H9zm-9 8h10v1H0zM1 1h8v8H1z" /></svg
				>
			</button>
			<button
				class="win98-button flex h-5 w-5 items-center justify-center"
				aria-label="Close"
				title="Close"
				on:click={() => showStatus('Window Closed (simulation)')}
			>
				<span class="text-xs font-black">×</span>
			</button>
		</div>
	</div>

	<!-- Menu Bar -->
	<div
		class="cursor-default border-b-2 border-t-zinc-50 border-r-zinc-400 border-b-zinc-400 border-l-zinc-50 bg-zinc-200 px-1 py-0.5 text-xs"
	>
		<span
			class="px-2 py-0.5 hover:bg-blue-600 hover:text-white"
			on:click={() => showStatus('File menu clicked')}><u>F</u>ile</span
		>
		<span
			class="px-2 py-0.5 hover:bg-blue-600 hover:text-white"
			on:click={() => showStatus('Edit menu clicked')}><u>E</u>dit</span
		>
		<span
			class="px-2 py-0.5 hover:bg-blue-600 hover:text-white"
			on:click={() => showStatus('View menu clicked')}><u>V</u>iew</span
		>
		<span
			class="px-2 py-0.5 hover:bg-blue-600 hover:text-white"
			on:click={() => showStatus('Favorites menu clicked')}>F<u>a</u>vorites</span
		>
		<span
			class="px-2 py-0.5 hover:bg-blue-600 hover:text-white"
			on:click={() => showStatus('Help menu clicked')}><u>H</u>elp</span
		>
	</div>

	<!-- Address Bar (like a browser) -->
	<div
		class="flex items-center border-b-2 border-t-zinc-50 border-r-zinc-400 border-b-zinc-400 border-l-zinc-50 bg-zinc-200 px-2 py-1 text-xs"
	>
		<span class="mr-1 text-zinc-600"><u>A</u>ddress:</span>
		<input
			type="text"
			value="C:\My Documents\Homepage.html"
			class="flex-grow border border-zinc-500 border-t-zinc-700 border-l-zinc-700 bg-white px-1 py-0.5 text-xs focus:outline-none"
			readonly
			on:focus={() => showStatus('Address bar focused')}
		/>
		<button
			class="win98-button ml-1 px-2 py-0.5 text-xs"
			on:click={() => showStatus('Go button clicked')}>Go</button
		>
	</div>

	<!-- Main Content Area - styled like a document/webpage -->
	<div class="flex h-[400px] bg-zinc-100 p-1">
		<!-- Optional: "Folders" or "Navigation" Pane -->
		<div
			class="mr-1 w-1/4 overflow-y-auto border-r-2 border-b-2 border-r-zinc-400 border-b-zinc-400 bg-zinc-200 p-2"
		>
			<h3 class="mb-1 text-xs font-bold text-zinc-700">Explorer</h3>
			<ul class="space-y-0.5 text-xs">
				<li
					class="cursor-default p-0.5 hover:bg-blue-600 hover:text-white"
					on:click={() => showStatus('My Computer clicked')}
				>
					My Computer
				</li>
				<li
					class="cursor-default p-0.5 hover:bg-blue-600 hover:text-white"
					on:click={() => showStatus('Network Neighborhood clicked')}
				>
					Network Neighborhood
				</li>
				<li
					class="cursor-default bg-blue-700 p-0.5 text-white hover:bg-blue-600 hover:text-white"
					on:click={() => showStatus('My Documents clicked')}
				>
					My Documents
				</li>
				<li
					class="ml-2 cursor-default p-0.5 hover:bg-blue-600 hover:text-white"
					on:click={() => showStatus('Photos clicked')}
				>
					- Photos
				</li>
				<li
					class="ml-2 cursor-default p-0.5 hover:bg-blue-600 hover:text-white"
					on:click={() => showStatus('Web Pages clicked')}
				>
					- Web Pages
				</li>
				<li
					class="cursor-default p-0.5 hover:bg-blue-600 hover:text-white"
					on:click={() => showStatus('Recycle Bin clicked')}
				>
					Recycle Bin
				</li>
			</ul>
		</div>

		<!-- Document View -->
		<div
			class="flex-grow overflow-y-auto border-2 border-t-zinc-700 border-r-zinc-300 border-b-zinc-300 border-l-zinc-700 bg-white p-4"
		>
			<h1 class="mb-4 text-2xl font-bold text-zinc-800">Welcome to My Vintage Page!</h1>
			<p class="mb-2 text-zinc-700">
				This is a paragraph of text on your new vintage-style web page. It's designed to look like
				something you might have seen in the late 90s. You can put any content here. For example,
				here's a <a
					href="#"
					class="text-blue-600 underline hover:text-blue-800"
					on:click|preventDefault={() => showStatus('Link clicked!')}>hyperlink</a
				>.
			</p>
			<p class="mb-4 text-zinc-700">
				Interacting with elements will give you feedback in the status bar below. Try clicking menu
				items or links!
			</p>

			<div class="my-4 border border-zinc-300 bg-zinc-50 p-2">
				<h2 class="mb-1 font-semibold text-zinc-700">A Section with Content</h2>
				<p class="text-xs text-zinc-600">
					You can have different sections on your page. Maybe some <span
						class="font-semibold text-green-600">important updates</span
					>
					or <span class="text-red-600">warnings</span> if needed.
				</p>
				<button
					class="win98-button mt-2 px-3 py-1 text-xs"
					on:click={() => showStatus('Action Button clicked successfully!')}>Do Something</button
				>
			</div>

			<p class="mt-6 text-xs text-zinc-500">
				This page uses Tailwind CSS to achieve its retro look and feel. The scrollbars are also
				styled to match!
			</p>
		</div>
	</div>

	<!-- Status Bar -->
	<div
		class="flex justify-between border-t-2 border-t-zinc-400 bg-zinc-200 px-2 py-0.5 text-xs text-zinc-700"
	>
		<div>{statusText}</div>
		<div class="text-zinc-500">My Computer</div>
	</div>
</div>

<style lang="postcss">
	/* For Webkit-based browsers (Chrome, Safari, Edge) */
	.overflow-y-auto::-webkit-scrollbar {
		width: 16px;
	}
	.overflow-y-auto::-webkit-scrollbar-track {
		background: var(--color-zinc-300);
	}
	.overflow-y-auto::-webkit-scrollbar-thumb {
		background-color: var(--color-zinc-200);
		border: 1px solid;
		border-color: var(--color-zinc-50) var(--color-zinc-400) var(--color-zinc-400)
			var(--color-zinc-50);
	}
	.overflow-y-auto::-webkit-scrollbar-thumb:hover {
		background-color: var(--color-zinc-100);
	}
	.overflow-y-auto::-webkit-scrollbar-button {
		background-color: var(--color-zinc-200);
		border: 1px solid;
		border-color: var(--color-zinc-50) var(--color-zinc-400) var(--color-zinc-400)
			var(--color-zinc-50);
		display: block;
		height: 13px;
	}
	.overflow-y-auto::-webkit-scrollbar-button:vertical:decrement {
		background-image: url("data:image/svg+xml;charset=utf-8,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' fill='black'%3E%3Cpath d='M5 3l4 4H1z'/%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: center;
	}
	.overflow-y-auto::-webkit-scrollbar-button:vertical:increment {
		background-image: url("data:image/svg+xml;charset=utf-8,%3Csvg xmlns='http://www.w3.org/2000/svg' width='10' height='10' fill='black'%3E%3Cpath d='M5 7L1 3h8z'/%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: center;
	}

	/* For Firefox */
	.overflow-y-auto {
		scrollbar-width: auto;
		scrollbar-color: var(--color-zinc-200) var(--color-zinc-300);
	}

	/* Basic button bevel for Win98 style title bar buttons */
	.win98-button {
		background-color: var(--color-zinc-200);
		color: var(--color-black);
		border-top: 1px solid var(--color-zinc-50);
		border-left: 1px solid var(--color-zinc-50);
		border-right: 1px solid var(--color-zinc-400);
		border-bottom: 1px solid var(--color-zinc-400);
		/* box-shadow: 1px 1px 0 var(--color-zinc-500); Removed, borders give enough effect */
	}
	.win98-button:active {
		border-top: 1px solid var(--color-zinc-400);
		border-left: 1px solid var(--color-zinc-400);
		border-right: 1px solid var(--color-zinc-50);
		border-bottom: 1px solid var(--color-zinc-50);
		/* box-shadow: inset 1px 1px 0 var(--color-zinc-500); */
		background-color: var(--color-zinc-300); /* Slightly darker on press */
	}
</style>
