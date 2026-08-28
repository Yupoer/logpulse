(() => {
  const display = document.querySelector('#timer-display');
  if (!display) return;

  let remaining = 20 * 60;
  let interval = null;

  const render = () => {
    const minutes = String(Math.floor(remaining / 60)).padStart(2, '0');
    const seconds = String(remaining % 60).padStart(2, '0');
    display.textContent = `${minutes}:${seconds}`;
  };

  document.querySelector('[data-action="start"]')?.addEventListener('click', () => {
    if (interval) return;
    interval = setInterval(() => {
      remaining = Math.max(0, remaining - 1);
      render();
      if (remaining === 0) {
        clearInterval(interval);
        interval = null;
      }
    }, 1000);
  });

  document.querySelector('[data-action="pause"]')?.addEventListener('click', () => {
    clearInterval(interval);
    interval = null;
  });

  document.querySelector('[data-action="reset"]')?.addEventListener('click', () => {
    clearInterval(interval);
    interval = null;
    remaining = 20 * 60;
    render();
  });

  render();
})();
