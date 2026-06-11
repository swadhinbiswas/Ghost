import anime from 'animejs';

const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches;

export function initMotion() {
  navScroll();
  scrollProgress();
  initIsoHero();
  if (reduce) {
    document.querySelectorAll<HTMLElement>('[data-reveal]').forEach((el) => {
      el.style.opacity = '1';
      el.style.transform = 'none';
    });
    return;
  }
  heroIntro();
  scrollReveals();
  ghostFloat();
  ghostEyes();
  magneticButtons();
  cardGlow();
  typeTerminals();
  counters();
}

/* ---- nav background on scroll ---- */
function navScroll() {
  const nav = document.querySelector('[data-nav]');
  if (!nav) return;
  const onScroll = () => nav.classList.toggle('scrolled', window.scrollY > 12);
  onScroll();
  window.addEventListener('scroll', onScroll, { passive: true });
}

/* ---- top scroll-progress bar ---- */
function scrollProgress() {
  const bar = document.querySelector<HTMLElement>('[data-progress]');
  if (!bar) return;
  const update = () => {
    const h = document.documentElement;
    const max = h.scrollHeight - h.clientHeight;
    const pct = max > 0 ? (h.scrollTop / max) * 100 : 0;
    bar.style.width = pct + '%';
  };
  update();
  window.addEventListener('scroll', update, { passive: true });
  window.addEventListener('resize', update, { passive: true });
}

/* ---- hero entrance choreography ---- */
function heroIntro() {
  const hero = document.querySelector('[data-hero]');
  if (!hero) return;
  const items = hero.querySelectorAll('[data-hero-item]');
  anime({
    targets: items,
    opacity: [0, 1],
    translateY: [28, 0],
    filter: ['blur(8px)', 'blur(0px)'],
    delay: anime.stagger(110, { start: 150 }),
    duration: 900,
    easing: 'easeOutExpo',
  });
  const ghost = hero.querySelector('[data-ghost]');
  if (ghost) {
    anime({ targets: ghost, opacity: [0, 1], scale: [0.86, 1], duration: 1100, delay: 250, easing: 'easeOutExpo' });
  }
}

/* ---- scroll-triggered reveals with stagger per section ---- */
function scrollReveals() {
  const groups = new Map<Element, Element[]>();
  document.querySelectorAll<HTMLElement>('[data-reveal]').forEach((el) => {
    const parent = el.closest('[data-reveal-group]') || el.parentElement || el;
    if (!groups.has(parent)) groups.set(parent, []);
    groups.get(parent)!.push(el);
  });

  const io = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (!entry.isIntersecting) return;
        const parent = entry.target;
        const targets = groups.get(parent) || [parent];
        anime({
          targets,
          opacity: [0, 1],
          translateY: [26, 0],
          delay: anime.stagger(80),
          duration: 760,
          easing: 'easeOutCubic',
        });
        io.unobserve(parent);
      });
    },
    { threshold: 0.16, rootMargin: '0px 0px -8% 0px' }
  );
  groups.forEach((_, parent) => io.observe(parent));
}

/* ---- organic ghost float (richer than a CSS loop) ---- */
function ghostFloat() {
  document.querySelectorAll<HTMLElement>('[data-ghost] .ghost-body-wrap').forEach((el) => {
    anime({ targets: el, translateY: [-10, 14], rotate: [-2.2, 2.2], duration: 4200, direction: 'alternate', loop: true, easing: 'easeInOutSine' });
  });
  document.querySelectorAll<HTMLElement>('[data-ghost] .ghost-shadow').forEach((el) => {
    anime({ targets: el, scaleX: [1.08, 0.82], opacity: [0.55, 0.3], duration: 4200, direction: 'alternate', loop: true, easing: 'easeInOutSine' });
  });
}

/* ---- pupils follow the cursor ---- */
function ghostEyes() {
  const pupils = document.querySelectorAll<SVGCircleElement>('[data-ghost] .pupil');
  if (!pupils.length) return;
  window.addEventListener('mousemove', (e) => {
    const dx = (e.clientX / window.innerWidth - 0.5) * 6;
    const dy = (e.clientY / window.innerHeight - 0.5) * 5;
    pupils.forEach((p) => {
      p.style.transform = `translate(${dx}px, ${dy}px)`;
    });
  }, { passive: true });
}

/* ---- magnetic buttons ---- */
function magneticButtons() {
  document.querySelectorAll<HTMLElement>('[data-magnetic]').forEach((btn) => {
    btn.addEventListener('mousemove', (e) => {
      const r = btn.getBoundingClientRect();
      const x = e.clientX - r.left - r.width / 2;
      const y = e.clientY - r.top - r.height / 2;
      btn.style.transform = `translate(${x * 0.22}px, ${y * 0.3}px)`;
    });
    btn.addEventListener('mouseleave', () => { btn.style.transform = ''; });
  });
}

/* ---- card cursor glow ---- */
function cardGlow() {
  document.querySelectorAll<HTMLElement>('.card').forEach((card) => {
    card.addEventListener('mousemove', (e) => {
      const r = card.getBoundingClientRect();
      card.style.setProperty('--mx', `${e.clientX - r.left}px`);
      card.style.setProperty('--my', `${e.clientY - r.top}px`);
    });
  });
}

/* ---- typewriter terminals ---- */
function typeTerminals() {
  document.querySelectorAll<HTMLElement>('[data-type]').forEach((el) => {
    const lines = JSON.parse(el.dataset.type || '[]') as string[];
    const out = el.querySelector('[data-type-out]') as HTMLElement | null;
    if (!out) return;
    const io = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (!entry.isIntersecting) return;
        io.disconnect();
        runType(out, lines);
      });
    }, { threshold: 0.4 });
    io.observe(el);
  });
}

function runType(out: HTMLElement, lines: string[]) {
  out.innerHTML = '';
  let li = 0;
  const caret = document.createElement('span');
  caret.className = 'caret';
  const next = () => {
    if (li >= lines.length) { caret.remove(); return; }
    const line = document.createElement('div');
    out.appendChild(line);
    line.appendChild(caret);
    const text = lines[li];
    let ci = 0;
    const tick = () => {
      if (ci <= text.length) {
        line.innerHTML = text.slice(0, ci);
        line.appendChild(caret);
        ci++;
        setTimeout(tick, 16 + Math.random() * 26);
      } else {
        li++;
        setTimeout(next, 260);
      }
    };
    tick();
  };
  next();
}

/* ---- number counters ---- */
function counters() {
  document.querySelectorAll<HTMLElement>('[data-count]').forEach((el) => {
    const target = parseFloat(el.dataset.count || '0');
    const suffix = el.dataset.suffix || '';
    const io = new IntersectionObserver((entries) => {
      entries.forEach((entry) => {
        if (!entry.isIntersecting) return;
        io.disconnect();
        const obj = { v: 0 };
        anime({
          targets: obj, v: target, duration: 1600, easing: 'easeOutExpo',
          update: () => {
            el.textContent = (Number.isInteger(target) ? Math.round(obj.v) : obj.v.toFixed(1)) + suffix;
          },
        });
      });
    }, { threshold: 0.6 });
    io.observe(el);
  });
}

/* ---- isometric control-plane hero ---- */
function initIsoHero() {
  const scene = document.querySelector<SVGElement>('[data-iso] .iso-scene');
  if (!scene) return;

  // Draw the connector links (and keep them visible even with reduced motion).
  const links = scene.querySelectorAll<SVGPathElement>('.iso-link');
  links.forEach((p) => {
    const len = p.getTotalLength();
    p.style.strokeDasharray = `${len}`;
    p.style.strokeDashoffset = reduce ? '0' : `${len}`;
  });

  if (reduce) return;

  // Links draw in.
  anime({
    targets: '.iso-link',
    strokeDashoffset: [anime.setDashoffset, 0],
    duration: 1100,
    delay: anime.stagger(120, { start: 300 }),
    easing: 'easeInOutSine',
  });

  // Slabs fade/scale in, then float forever (staggered).
  const slabs = scene.querySelectorAll<SVGGElement>('.iso-slab');
  anime({
    targets: slabs,
    opacity: [0, 1],
    duration: 700,
    delay: anime.stagger(90, { start: 200 }),
    easing: 'easeOutQuad',
  });
  slabs.forEach((slab, i) => {
    anime({
      targets: slab,
      translateY: [-5, 7],
      duration: 3200 + i * 180,
      delay: i * 160,
      direction: 'alternate',
      loop: true,
      easing: 'easeInOutSine',
    });
  });

  // Ghost bob + shadow breathing.
  anime({ targets: '.iso-ghost-body', translateY: [-6, 8], duration: 3600, direction: 'alternate', loop: true, easing: 'easeInOutSine' });
  anime({ targets: '.iso-ghost-shadow', scaleX: [1.05, 0.8], opacity: [0.3, 0.16], duration: 3600, direction: 'alternate', loop: true, easing: 'easeInOutSine' });

  // Core glow pulse.
  anime({ targets: '.iso-core-glow', scale: [1, 1.12], opacity: [0.7, 1], duration: 2600, direction: 'alternate', loop: true, easing: 'easeInOutSine' });

  // Core emblem slow spin + shimmer.
  anime({ targets: '.iso-emblem', rotate: '1turn', duration: 14000, loop: true, easing: 'linear' });

  // Radar rings expanding from the core.
  scene.querySelectorAll<SVGEllipseElement>('.iso-ring').forEach((ring, i) => {
    anime({
      targets: ring,
      scale: [0.5, 1.4],
      opacity: [0.7, 0],
      transformOrigin: ['50% 50%', '50% 50%'],
      duration: 3000,
      delay: i * 1500,
      loop: true,
      easing: 'easeOutSine',
    });
  });

  // Data pulses travel along the links from the core outward.
  scene.querySelectorAll<SVGCircleElement>('.iso-pulse').forEach((dot, i) => {
    const x1 = Number(dot.dataset.x1), y1 = Number(dot.dataset.y1);
    const x2 = Number(dot.dataset.x2), y2 = Number(dot.dataset.y2);
    anime({
      targets: dot,
      cx: [x1, x2],
      cy: [y1, y2],
      opacity: [{ value: 1, duration: 200 }, { value: 1, duration: 800 }, { value: 0, duration: 300 }],
      duration: 1800,
      delay: i * 320,
      loop: true,
      easing: 'easeInOutQuad',
    });
  });

  // Subtle parallax tied to the pointer.
  window.addEventListener('mousemove', (e) => {
    const dx = (e.clientX / window.innerWidth - 0.5) * 16;
    const dy = (e.clientY / window.innerHeight - 0.5) * 12;
    scene.style.transform = `translate(${dx}px, ${dy}px)`;
  }, { passive: true });
}
