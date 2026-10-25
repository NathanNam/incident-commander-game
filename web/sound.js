// Incident Commander sound: chiptune music and effects, synthesised with Web Audio
// (no audio files). Exposes window.icSound:
//
//   icSound.sfx(name)        play an effect: start, pause, resume, collect, levelup, gameover, won
//   icSound.music(on, level) start (or keep playing at the level's tempo) / stop the music
//   icSound.mode             'all' | 'fx' | 'off'; icSound.setMode(mode)
//
// The AudioContext is created on the first call, which always follows a key press
// or tap, so browsers allow it to play.
(() => {
  let ac, master, musicBus, fxBus, noise;
  let mode = 'all';

  function ctx() {
    if (ac) { if (ac.state === 'suspended') ac.resume(); return ac; }
    ac = new (window.AudioContext || window.webkitAudioContext)();
    master = ac.createGain(); master.gain.value = 0.9; master.connect(ac.destination);
    musicBus = ac.createGain(); musicBus.connect(master);
    fxBus = ac.createGain(); fxBus.connect(master);
    // One second of white noise, reused for drums and the crash.
    noise = ac.createBuffer(1, ac.sampleRate, ac.sampleRate);
    const d = noise.getChannelData(0);
    for (let i = 0; i < d.length; i++) d[i] = Math.random() * 2 - 1;
    applyMode();
    return ac;
  }
  function applyMode() {
    if (!ac) return;
    musicBus.gain.value = mode === 'all' ? 1 : 0;
    fxBus.gain.value = mode === 'off' ? 0 : 1;
  }

  const hz = (midi) => 440 * 2 ** ((midi - 69) / 12);

  // A pitched note with a short attack and decay.
  function tone(bus, type, midiOrHz, t, dur, vol, slideTo) {
    const o = ac.createOscillator(), g = ac.createGain();
    o.type = type;
    const f = midiOrHz > 127 ? midiOrHz : hz(midiOrHz);
    o.frequency.setValueAtTime(f, t);
    if (slideTo) o.frequency.exponentialRampToValueAtTime(slideTo, t + dur);
    g.gain.setValueAtTime(0.0001, t);
    g.gain.exponentialRampToValueAtTime(vol, t + 0.005);
    g.gain.exponentialRampToValueAtTime(0.0001, t + dur);
    o.connect(g).connect(bus);
    o.start(t); o.stop(t + dur + 0.02);
  }

  // Filtered noise: hi-hats, snares and the crash.
  function hiss(bus, t, dur, vol, freq, type = 'highpass') {
    const s = ac.createBufferSource(), f = ac.createBiquadFilter(), g = ac.createGain();
    s.buffer = noise; f.type = type; f.frequency.value = freq;
    g.gain.setValueAtTime(vol, t); g.gain.exponentialRampToValueAtTime(0.0001, t + dur);
    s.connect(f).connect(g).connect(bus);
    s.start(t, Math.random() * 0.5); s.stop(t + dur + 0.02);
  }

  function kick(t) { tone(musicBus, 'sine', 150, t, 0.16, 0.5, 45); }

  // ---- effects ---------------------------------------------------------------------------
  const FX = {
    start(t) { [60, 64, 67, 72].forEach((n, i) => tone(fxBus, 'square', n + 12, t + i * 0.06, 0.09, 0.07)); },
    pause(t) { tone(fxBus, 'square', 76, t, 0.07, 0.06); tone(fxBus, 'square', 69, t + 0.07, 0.1, 0.06); },
    resume(t) { tone(fxBus, 'square', 69, t, 0.07, 0.06); tone(fxBus, 'square', 76, t + 0.07, 0.1, 0.06); },
    collect(t) {
      tone(fxBus, 'square', 84, t, 0.05, 0.07);
      tone(fxBus, 'square', 91, t + 0.045, 0.09, 0.07);
      tone(fxBus, 'triangle', 96, t + 0.045, 0.12, 0.05);
    },
    levelup(t) {
      [72, 76, 79, 84, 88].forEach((n, i) => tone(fxBus, 'square', n, t + i * 0.08, 0.12, 0.07));
      tone(fxBus, 'triangle', 48, t, 0.45, 0.12);
      tone(fxBus, 'triangle', 60, t + 0.32, 0.4, 0.1);
    },
    gameover(t) {
      hiss(fxBus, t, 0.35, 0.25, 900, 'lowpass');
      tone(fxBus, 'square', 600, t, 0.25, 0.08, 120);
      [67, 63, 60, 55].forEach((n, i) => tone(fxBus, 'triangle', n, t + 0.3 + i * 0.16, 0.18, 0.12));
    },
    won(t) {
      const tune = [[72, 0], [76, .1], [79, .2], [84, .3], [79, .45], [84, .55], [88, .7]];
      tune.forEach(([n, at]) => tone(fxBus, 'square', n, t + at, at === .7 ? .5 : .12, 0.07));
      [[48, 0], [55, .3], [60, .7]].forEach(([n, at]) => tone(fxBus, 'triangle', n, t + at, .4, 0.12));
    },
  };

  // ---- music: an original eight-bar loop in A minor (Am F C G) -------------------------
  // One string per bar, 16 sixteenth-note steps; numbers are MIDI notes, '.' is a rest.
  const N = (s) => s.trim().split(/\s+/).map((x) => (x === '.' ? null : +x));
  const LEAD = [
    '69 . 72 . 76 . 72 . 74 . 72 . 69 . 67 .',
    '65 . 69 . 72 . 69 . 77 . 76 . 72 . . .',
    '67 . 72 . 76 . 79 . 77 . 76 . 74 . 72 .',
    '71 . 74 . 79 . . . 77 . 74 . 71 . . .',
    '81 . 79 . 81 . 76 . 81 . 79 . 76 . 72 .',
    '77 . 76 . 77 . 72 . 69 . 72 . 77 . . .',
    '76 . 79 . 84 . 79 . 76 . 72 . 76 . 79 .',
    '83 . 79 . 74 . 79 . 83 . . . 86 . . .',
  ].map(N);
  const ROOTS = [45, 41, 48, 43, 45, 41, 48, 43]; // A F C G, twice
  const CHORD = { 45: [0, 3, 7], 41: [0, 4, 7], 48: [0, 4, 7], 43: [0, 4, 7] };

  let playing = false, timer = null, step = 0, nextT = 0, bpm = 112;

  function scheduleStep(i, t) {
    const bar = Math.floor(i / 16) % 8, s = i % 16, root = ROOTS[bar];
    const sixteenth = 60 / bpm / 4;
    // Bass: eighth notes, root and octave.
    if (s % 2 === 0) tone(musicBus, 'triangle', root + (s % 4 === 2 ? 12 : 0), t, sixteenth * 1.8, 0.16);
    // Arpeggio: quiet chord tones on every step, for the arcade shimmer.
    const ch = CHORD[root];
    tone(musicBus, 'square', root + 24 + ch[s % 3], t, sixteenth * 0.6, 0.012);
    // Lead.
    const n = LEAD[bar][s];
    if (n) tone(musicBus, 'square', n, t, sixteenth * 1.7, 0.035);
    // Drums: kick on 1 and 3, snare on 2 and 4, hats on the off-beats.
    if (s === 0 || s === 8 || (bar % 2 === 1 && s === 14)) kick(t);
    if (s === 4 || s === 12) hiss(musicBus, t, 0.12, 0.12, 1800, 'bandpass');
    if (s % 4 === 2) hiss(musicBus, t, 0.04, 0.05, 7000);
  }

  // Look-ahead scheduler: every 25 ms, queue the notes due in the next 120 ms.
  function tick() {
    const sixteenth = 60 / bpm / 4;
    while (nextT < ac.currentTime + 0.12) {
      if (mode === 'all') scheduleStep(step, nextT); // keep time even when muted
      step = (step + 1) % (16 * 8);
      nextT += sixteenth;
    }
  }

  function music(on, level = 1) {
    bpm = 112 + (Math.min(level, 10) - 1) * 6;
    if (on === playing) return;
    playing = on;
    if (on) {
      ctx();
      nextT = ac.currentTime + 0.05;
      timer = setInterval(tick, 25);
      tick();
    } else {
      clearInterval(timer); timer = null;
    }
  }

  window.icSound = {
    get mode() { return mode; },
    setMode(m) { mode = m; applyMode(); },
    sfx(name) { if (mode !== 'off' && FX[name]) { ctx(); FX[name](ac.currentTime + 0.01); } },
    music,
    restartMusic() { step = 0; },
  };
})();
