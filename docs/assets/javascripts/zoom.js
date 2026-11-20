// Click an image or Mermaid diagram in a page to open it in a dialog sized to it,
// where the wheel or a pinch zooms (the dialog grows until it fills the screen,
// then zooming continues inside it), dragging pans, double-click (or double-tap)
// resets and a click or Esc closes.
//
// Material renders Mermaid into a closed shadow root, so the diagram can't be
// copied from the page. This script runs before Material's async Mermaid render,
// keeps each diagram's source, and renders it again into the dialog.
(() => {
  const mermaidSources = [...document.querySelectorAll("pre.mermaid")].map(
    (pre) => pre.textContent,
  );
  let renders = 0;

  const dialog = document.createElement("dialog");
  dialog.className = "zoom-dialog";
  dialog.innerHTML =
    '<div class="zoom-stage"><div class="zoom-content"></div></div>' +
    '<p class="zoom-hint">' +
    '<span class="zoom-hint-mouse">Scroll to zoom · drag to pan · double-click to reset · click or Esc to close</span>' +
    '<span class="zoom-hint-touch">Pinch to zoom · drag to pan · double-tap to reset · tap to close</span>' +
    "</p>";
  document.body.append(dialog);
  const stage = dialog.querySelector(".zoom-stage");
  const content = dialog.querySelector(".zoom-content");

  let scale = 1;
  let x = 0;
  let y = 0;
  // The content's size at scale 1, and the dialog's size around the stage.
  let base = { width: 0, height: 0 };
  let chrome = { width: 0, height: 0 };
  const clampPan = (v, overflow) => Math.min(Math.max(v, -overflow / 2), overflow / 2);
  // The stage grows with the zoom until the dialog fills the viewport; past
  // that, zooming happens inside it. Panning stops at the content's edges.
  const apply = () => {
    const width = base.width * scale;
    const height = base.height * scale;
    const stageWidth = Math.min(width, window.innerWidth - chrome.width);
    const stageHeight = Math.min(height, window.innerHeight - chrome.height);
    stage.style.width = `${stageWidth}px`;
    stage.style.height = `${stageHeight}px`;
    x = clampPan(x, Math.max(0, width - stageWidth));
    y = clampPan(y, Math.max(0, height - stageHeight));
    content.style.transform = `translate(${x}px, ${y}px) scale(${scale})`;
  };
  const reset = () => {
    scale = 1;
    x = 0;
    y = 0;
    apply();
  };
  const clampScale = (s) => Math.min(10, Math.max(0.5, s));

  // Show the node at its natural size, shrunk to fit the viewport. The margins
  // leave room for the dialog padding and the hint line (see zoom.css).
  const fit = (node, width, height) => {
    const f = Math.min(1, (window.innerWidth - 48) / width, (window.innerHeight - 96) / height);
    base = { width: width * f, height: height * f };
    node.style.width = `${base.width}px`;
    node.style.height = `${base.height}px`;
  };

  const open = (node) => {
    content.replaceChildren(node);
    dialog.showModal();
    stage.style.width = `${base.width}px`;
    stage.style.height = `${base.height}px`;
    chrome = {
      width: dialog.offsetWidth - stage.offsetWidth,
      height: dialog.offsetHeight - stage.offsetHeight,
    };
    reset();
  };

  document.addEventListener("click", async (event) => {
    const target = event.target;
    if (!(target instanceof Element) || !target.closest(".md-content")) return;

    const img = target.closest("img");
    if (img && !img.closest("a")) {
      const copy = img.cloneNode();
      copy.removeAttribute("width");
      copy.removeAttribute("height");
      fit(copy, img.naturalWidth || img.width, img.naturalHeight || img.height);
      open(copy);
      return;
    }

    const diagram = target.closest("div.mermaid");
    if (diagram && window.mermaid) {
      const index = [...document.querySelectorAll(".md-content .mermaid")].indexOf(diagram);
      const source = mermaidSources[index];
      if (source === undefined) return;
      const { svg } = await window.mermaid.render(`__zoom_${renders++}`, source);
      const wrapper = document.createElement("div");
      wrapper.className = "zoom-mermaid";
      wrapper.innerHTML = svg;
      const el = wrapper.querySelector("svg");
      const { width, height } = el.viewBox.baseVal;
      el.removeAttribute("width");
      el.style.maxWidth = "none";
      fit(el, width, height);
      open(wrapper);
    }
  });

  // Zoom to `next`, keeping the stage point (px, py), relative to its centre, fixed.
  const zoomAt = (next, px, py) => {
    x = px - ((px - x) * next) / scale;
    y = py - ((py - y) * next) / scale;
    scale = next;
    apply();
  };
  const fromCentre = (clientX, clientY) => {
    const rect = stage.getBoundingClientRect();
    return [clientX - rect.left - rect.width / 2, clientY - rect.top - rect.height / 2];
  };

  stage.addEventListener(
    "wheel",
    (event) => {
      event.preventDefault();
      zoomAt(clampScale(scale * Math.exp(-event.deltaY * 0.002)), ...fromCentre(event.clientX, event.clientY));
    },
    { passive: false },
  );

  // One pointer pans; two pointers pinch-zoom around their midpoint.
  const pointers = new Map();
  let drag = null;
  let pinch = null;
  let moved = false;
  const startDrag = () => {
    const [p] = pointers.values();
    drag = { startX: p.x - x, startY: p.y - y, downX: p.x, downY: p.y };
  };
  const pinchState = () => {
    const [a, b] = pointers.values();
    return { dist: Math.hypot(a.x - b.x, a.y - b.y), mid: fromCentre((a.x + b.x) / 2, (a.y + b.y) / 2) };
  };

  stage.addEventListener("pointerdown", (event) => {
    stage.setPointerCapture(event.pointerId);
    pointers.set(event.pointerId, { x: event.clientX, y: event.clientY });
    if (pointers.size === 1) {
      moved = false;
      startDrag();
    } else if (pointers.size === 2) {
      drag = null;
      moved = true;
      pinch = { ...pinchState(), scale, x, y };
    }
  });
  stage.addEventListener("pointermove", (event) => {
    if (!pointers.has(event.pointerId)) return;
    pointers.set(event.pointerId, { x: event.clientX, y: event.clientY });
    if (pinch && pointers.size >= 2) {
      const { dist, mid } = pinchState();
      // Restart from the pinch's start so the content follows the fingers.
      scale = pinch.scale;
      x = pinch.x + mid[0] - pinch.mid[0];
      y = pinch.y + mid[1] - pinch.mid[1];
      zoomAt(clampScale((pinch.scale * dist) / pinch.dist), ...mid);
    } else if (drag) {
      x = event.clientX - drag.startX;
      y = event.clientY - drag.startY;
      if (Math.hypot(event.clientX - drag.downX, event.clientY - drag.downY) > 3) moved = true;
      apply();
    }
  });
  const release = (event) => {
    pointers.delete(event.pointerId);
    pinch = null;
    // Lifting one finger of a pinch carries on as a drag with the other.
    if (pointers.size === 1) startDrag();
    else drag = null;
  };
  stage.addEventListener("pointerup", release);
  stage.addEventListener("pointercancel", release);

  // A click on the content closes after a short wait, so a second click can
  // reset instead; this also covers double-tap, where mobile browsers don't
  // reliably fire dblclick. The end of a drag or pinch doesn't close. A click
  // outside the content (padding or backdrop) closes at once.
  let closeTimer;
  let lastClick = 0;
  dialog.addEventListener("click", (event) => {
    if (!stage.contains(event.target)) {
      dialog.close();
      return;
    }
    if (moved) return;
    clearTimeout(closeTimer);
    if (event.timeStamp - lastClick < 300) {
      lastClick = 0;
      reset();
    } else {
      lastClick = event.timeStamp;
      closeTimer = setTimeout(() => dialog.close(), 300);
    }
  });
})();
