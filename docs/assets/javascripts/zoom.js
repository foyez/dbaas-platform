// Click an image or Mermaid diagram in a page to open it full screen, where the
// wheel zooms, dragging pans, double-click resets and a click or Esc closes.
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
    '<p class="zoom-hint">Scroll to zoom · drag to pan · double-click to reset · click or Esc to close</p>';
  document.body.append(dialog);
  const stage = dialog.querySelector(".zoom-stage");
  const content = dialog.querySelector(".zoom-content");

  let scale = 1;
  let x = 0;
  let y = 0;
  const apply = () => {
    content.style.transform = `translate(${x}px, ${y}px) scale(${scale})`;
  };
  const reset = () => {
    scale = 1;
    x = 0;
    y = 0;
    apply();
  };

  const open = (node) => {
    content.replaceChildren(node);
    reset();
    dialog.showModal();
  };

  document.addEventListener("click", async (event) => {
    const target = event.target;
    if (!(target instanceof Element) || !target.closest(".md-content")) return;

    const img = target.closest("img");
    if (img && !img.closest("a")) {
      const copy = img.cloneNode();
      copy.removeAttribute("width");
      copy.removeAttribute("height");
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
      open(wrapper);
    }
  });

  stage.addEventListener(
    "wheel",
    (event) => {
      event.preventDefault();
      const rect = stage.getBoundingClientRect();
      // Zoom around the cursor: keep the point under it fixed.
      const px = event.clientX - rect.left - rect.width / 2;
      const py = event.clientY - rect.top - rect.height / 2;
      const next = Math.min(10, Math.max(0.5, scale * Math.exp(-event.deltaY * 0.002)));
      x = px - ((px - x) * next) / scale;
      y = py - ((py - y) * next) / scale;
      scale = next;
      apply();
    },
    { passive: false },
  );

  let drag = null;
  let moved = false;
  stage.addEventListener("pointerdown", (event) => {
    drag = { startX: event.clientX - x, startY: event.clientY - y, downX: event.clientX, downY: event.clientY };
    moved = false;
    stage.setPointerCapture(event.pointerId);
  });
  stage.addEventListener("pointermove", (event) => {
    if (!drag) return;
    x = event.clientX - drag.startX;
    y = event.clientY - drag.startY;
    if (Math.hypot(event.clientX - drag.downX, event.clientY - drag.downY) > 3) moved = true;
    apply();
  });
  stage.addEventListener("pointerup", () => {
    drag = null;
  });
  stage.addEventListener("dblclick", reset);

  // A click closes, after a short wait so a double-click can reset instead.
  // The end of a drag doesn't close.
  let closeTimer;
  stage.addEventListener("click", (event) => {
    if (moved) return;
    clearTimeout(closeTimer);
    if (event.detail === 1) closeTimer = setTimeout(() => dialog.close(), 250);
  });
})();
