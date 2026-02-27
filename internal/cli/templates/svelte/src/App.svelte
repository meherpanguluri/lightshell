<script>
  let platform = $state('')
  let arch = $state('')

  async function init() {
    platform = await lightshell.system.platform()
    arch = await lightshell.system.arch()
  }

  $effect(() => {
    init()
  })
</script>

<main>
  <h1>{{TITLE}}</h1>
  <p>Your LightShell + Svelte app is running.</p>
  {#if platform}
    <div class="info">{platform}/{arch}</div>
  {/if}
</main>
