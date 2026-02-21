// Draggable window cards for LightShell landing page
(function () {
  const workspace = document.getElementById('hw-workspace');
  if (!workspace) return;
  const cards = workspace.querySelectorAll('.hw-card');
  if (!cards.length) return;

  var topZ = 10;
  var activeCard = null;
  var startX = 0;
  var startY = 0;
  var cardStartX = 0;
  var cardStartY = 0;

  // Parse initial positions from CSS custom properties and set them
  cards.forEach(function (card, i) {
    var style = getComputedStyle(card);
    var x = parseFloat(style.getPropertyValue('--hw-x')) || 0;
    var y = parseFloat(style.getPropertyValue('--hw-y')) || 0;
    card._hwX = x;
    card._hwY = y;
    card.style.transform =
      'translate(' + x + 'px, ' + y + 'px) rotate(var(--hw-rotate, 0deg))';
    card.style.zIndex = String(i + 1);
  });

  function getPointer(e) {
    if (e.touches && e.touches.length > 0) {
      return { x: e.touches[0].clientX, y: e.touches[0].clientY };
    }
    return { x: e.clientX, y: e.clientY };
  }

  function onStart(e) {
    var card = e.target.closest('.hw-card');
    if (!card) return;

    // Don't interfere with dot clicks
    if (e.target.classList.contains('hw-dot')) return;

    activeCard = card;
    var p = getPointer(e);
    startX = p.x;
    startY = p.y;
    cardStartX = card._hwX || 0;
    cardStartY = card._hwY || 0;

    // Bring to front
    topZ++;
    card.style.zIndex = String(topZ);

    // Add dragging state
    card.classList.add('hw-dragging');

    // Remove rotation during drag for cleaner feel, add slight scale
    card.style.transform =
      'translate(' +
      cardStartX +
      'px, ' +
      cardStartY +
      'px) rotate(0deg) scale(1.04)';

    e.preventDefault();
  }

  function onMove(e) {
    if (!activeCard) return;

    var p = getPointer(e);
    var dx = p.x - startX;
    var dy = p.y - startY;
    var newX = cardStartX + dx;
    var newY = cardStartY + dy;

    activeCard._hwX = newX;
    activeCard._hwY = newY;
    activeCard.style.transform =
      'translate(' + newX + 'px, ' + newY + 'px) rotate(0deg) scale(1.04)';

    e.preventDefault();
  }

  function onEnd() {
    if (!activeCard) return;

    // Restore rotation
    var rotate =
      getComputedStyle(activeCard).getPropertyValue('--hw-rotate') || '0deg';
    activeCard.style.transform =
      'translate(' +
      activeCard._hwX +
      'px, ' +
      activeCard._hwY +
      'px) rotate(' +
      rotate +
      ')';
    activeCard.classList.remove('hw-dragging');

    activeCard = null;
  }

  // Mouse events
  workspace.addEventListener('mousedown', onStart);
  document.addEventListener('mousemove', onMove);
  document.addEventListener('mouseup', onEnd);

  // Touch events
  workspace.addEventListener('touchstart', onStart, { passive: false });
  document.addEventListener('touchmove', onMove, { passive: false });
  document.addEventListener('touchend', onEnd);

  // Entrance animation — stagger cards in from below with fade
  cards.forEach(function (card, i) {
    card.style.opacity = '0';
    card.style.transform =
      'translate(' +
      card._hwX +
      'px, ' +
      (card._hwY + 40) +
      'px) rotate(var(--hw-rotate, 0deg))';

    setTimeout(function () {
      card.style.transition =
        'opacity 0.6s cubic-bezier(0.16, 1, 0.3, 1), transform 0.6s cubic-bezier(0.16, 1, 0.3, 1)';
      card.style.opacity = '1';
      card.style.transform =
        'translate(' +
        card._hwX +
        'px, ' +
        card._hwY +
        'px) rotate(var(--hw-rotate, 0deg))';

      // After entrance animation finishes, set a faster transition for dragging
      setTimeout(function () {
        card.style.transition =
          'box-shadow 0.25s cubic-bezier(0.16, 1, 0.3, 1)';
      }, 650);
    }, 200 + i * 120);
  });
})();
