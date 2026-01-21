const snowflakeChars = ["❄", "❅", "❆", "✻", "✼", "❉"];
const sparkleChars = ["✦", "✧", "⭐", "✨", "💫", "⭑"];
const confettiColors = [
  "#ff6b6b",
  "#ffd700",
  "#74b9ff",
  "#a29bfe",
  "#55efc4",
  "#fd79a8",
  "#00b894",
];
let lightsOn = true;
let musicPlaying = false;
let audioContext = null;

// Create snowflakes
function createSnowflakes(count = 40) {
  const snow = document.getElementById("snow");
  for (let i = 0; i < count; i++) {
    const flake = document.createElement("div");
    flake.className = "snowflake";
    flake.innerHTML =
      snowflakeChars[Math.floor(Math.random() * snowflakeChars.length)];
    flake.style.left = Math.random() * 100 + "%";
    flake.style.fontSize = Math.random() * 10 + 8 + "px";
    flake.style.animationDuration = Math.random() * 5 + 5 + "s";
    flake.style.animationDelay = Math.random() * 10 + "s";
    flake.style.opacity = Math.random() * 0.6 + 0.4;
    snow.appendChild(flake);
  }
}

// Snowflakes burst from cursor
function createSnowBurst(x, y, count = 12) {
  const card = document.getElementById("card");
  const rect = card.getBoundingClientRect();
  const startX = x - rect.left;
  const startY = y - rect.top;

  for (let i = 0; i < count; i++) {
    const flake = document.createElement("div");
    flake.className = "snowflake-burst";
    flake.innerHTML =
      snowflakeChars[Math.floor(Math.random() * snowflakeChars.length)];
    flake.style.left = startX + "px";
    flake.style.top = startY + "px";
    flake.style.fontSize = Math.random() * 14 + 10 + "px";
    flake.style.opacity = Math.random() * 0.4 + 0.6;
    card.appendChild(flake);

    const spreadX = (Math.random() - 0.5) * 100;
    const fallDistance = 150 + Math.random() * 250;
    const duration = 1500 + Math.random() * 1500;
    const delay = Math.random() * 150;

    setTimeout(() => {
      flake.style.transition = `all ${duration}ms cubic-bezier(0.25, 0.46, 0.45, 0.94)`;
      flake.style.transform = `translate(${spreadX}px, ${fallDistance}px) rotate(${
        360 + Math.random() * 360
      }deg)`;
      flake.style.opacity = "0";
    }, delay);

    setTimeout(() => flake.remove(), duration + delay + 100);
  }
}

// Mouse sparkle trail
function setupSparkleTrail() {
  const card = document.getElementById("card");
  let lastSparkle = 0;

  card.addEventListener("mousemove", (e) => {
    const now = Date.now();
    if (now - lastSparkle < 50) return;
    lastSparkle = now;

    const rect = card.getBoundingClientRect();
    const sparkle = document.createElement("div");
    sparkle.className = "sparkle";
    sparkle.innerHTML =
      sparkleChars[Math.floor(Math.random() * sparkleChars.length)];
    sparkle.style.left = e.clientX - rect.left + "px";
    sparkle.style.top = e.clientY - rect.top + "px";
    sparkle.style.color =
      confettiColors[Math.floor(Math.random() * confettiColors.length)];
    card.appendChild(sparkle);

    setTimeout(() => sparkle.remove(), 1000);
  });
}

// Shooting stars
function createShootingStar() {
  const card = document.getElementById("card");
  const star = document.createElement("div");
  star.className = "shooting-star";
  star.style.left = Math.random() * 200 + "px";
  star.style.top = Math.random() * 100 + 20 + "px";
  star.style.animation = "shootingStar 1.5s ease-out forwards";
  card.appendChild(star);

  setTimeout(() => star.remove(), 1500);
}

function startShootingStars() {
  setInterval(() => {
    if (Math.random() > 0.7) createShootingStar();
  }, 3000);
}

// Santa flying across
function flySanta() {
  const santa = document.getElementById("santa");
  santa.style.animation = "flySanta 8s ease-in-out forwards";

  setTimeout(() => {
    santa.style.animation = "none";
  }, 8000);
}

function startSantaFlights() {
  flySanta();
  setInterval(flySanta, 20000);
}

// Tree ornaments
function createOrnaments() {
  const container = document.getElementById("treeContainer");
  const colors = ["red", "gold", "blue", "purple"];
  const positions = [
    { top: 70, left: -20 },
    { top: 70, left: 20 },
    { top: 110, left: -40 },
    { top: 110, left: 0 },
    { top: 110, left: 40 },
    { top: 160, left: -55 },
    { top: 160, left: -20 },
    { top: 160, left: 20 },
    { top: 160, left: 55 },
    { top: 200, left: -70 },
    { top: 200, left: -30 },
    { top: 200, left: 10 },
    { top: 200, left: 50 },
  ];

  positions.forEach((pos, i) => {
    const ornament = document.createElement("div");
    ornament.className = "ornament " + colors[i % colors.length];
    ornament.style.top = pos.top + "px";
    ornament.style.left = "calc(50% + " + pos.left + "px)";
    ornament.style.animationDelay = Math.random() * 2 + "s";
    container.appendChild(ornament);
  });
}

// Tree lights
function createLights() {
  const container = document.getElementById("treeContainer");
  const colors = [
    "#ff6b6b",
    "#ffd700",
    "#74b9ff",
    "#a29bfe",
    "#55efc4",
    "#fd79a8",
  ];

  for (let i = 0; i < 20; i++) {
    const light = document.createElement("div");
    light.className = "light";
    light.style.background = colors[Math.floor(Math.random() * colors.length)];
    light.style.color = light.style.background;
    light.style.top = 50 + Math.random() * 180 + "px";
    light.style.left = "calc(50% + " + (Math.random() * 120 - 60) + "px)";
    light.style.animationDelay = Math.random() * 1 + "s";
    light.style.animationDuration = Math.random() * 0.5 + 0.5 + "s";
    container.appendChild(light);
  }
}

// Shake tree
function setupTreeShake() {
  const tree = document.getElementById("treeContainer");

  tree.addEventListener("click", (e) => {
    e.stopPropagation();
    tree.classList.add("shake");
    createSnowBurst(e.clientX, e.clientY, 20);

    setTimeout(() => tree.classList.remove("shake"), 500);
  });
}

// Light switch
function setupLightSwitch() {
  const lightSwitch = document.getElementById("lightSwitch");
  const switchEl = document.getElementById("switch");
  const treeContainer = document.getElementById("treeContainer");

  lightSwitch.addEventListener("click", (e) => {
    e.stopPropagation();
    lightsOn = !lightsOn;
    switchEl.classList.toggle("on", lightsOn);
    treeContainer.classList.toggle("lights-off", !lightsOn);
  });
}

// Confetti burst
function createConfetti(x, y, count = 30) {
  const card = document.getElementById("card");
  const rect = card.getBoundingClientRect();
  const startX = x - rect.left;
  const startY = y - rect.top;

  for (let i = 0; i < count; i++) {
    const confetti = document.createElement("div");
    confetti.className = "confetti";
    confetti.style.left = startX + "px";
    confetti.style.top = startY + "px";
    confetti.style.background =
      confettiColors[Math.floor(Math.random() * confettiColors.length)];
    confetti.style.borderRadius = Math.random() > 0.5 ? "50%" : "2px";
    confetti.style.width = Math.random() * 8 + 5 + "px";
    confetti.style.height = Math.random() * 8 + 5 + "px";
    card.appendChild(confetti);

    const spreadX = (Math.random() - 0.5) * 150;
    const spreadY = -(Math.random() * 100 + 50);
    const fallY = Math.random() * 200 + 100;
    const duration = 1500 + Math.random() * 1000;

    confetti.animate(
      [
        { transform: "translate(0, 0) rotate(0deg)", opacity: 1 },
        {
          transform: `translate(${spreadX}px, ${spreadY}px) rotate(360deg)`,
          opacity: 1,
          offset: 0.3,
        },
        {
          transform: `translate(${spreadX}px, ${fallY}px) rotate(720deg)`,
          opacity: 0,
        },
      ],
      {
        duration,
        easing: "cubic-bezier(0.25, 0.46, 0.45, 0.94)",
        fill: "forwards",
      }
    );

    setTimeout(() => confetti.remove(), duration);
  }
}

// Open presents
function setupPresents() {
  const presents = document.querySelectorAll(".present");

  presents.forEach((present) => {
    present.addEventListener("dblclick", (e) => {
      e.stopPropagation();
      if (present.dataset.opened === "true") return;

      present.dataset.opened = "true";
      present.classList.add("opened");
      createConfetti(e.clientX, e.clientY, 40);

      setTimeout(() => present.classList.remove("opened"), 500);
    });
  });
}

// Draggable elements
function setupDraggable(selector) {
  const elements = document.querySelectorAll(selector);
  const card = document.getElementById("card");

  elements.forEach((el) => {
    let isDragging = false;
    let hasMoved = false;
    let offsetX, offsetY;

    const startDrag = (clientX, clientY) => {
      isDragging = true;
      hasMoved = false;
      el.classList.add("dragging");
      const rect = el.getBoundingClientRect();
      offsetX = clientX - rect.left;
      offsetY = clientY - rect.top;
    };

    const moveDrag = (clientX, clientY) => {
      if (!isDragging) return;
      hasMoved = true;
      const cardRect = card.getBoundingClientRect();
      let newX = clientX - cardRect.left - offsetX;
      let newY = clientY - cardRect.top - offsetY;
      const elRect = el.getBoundingClientRect();
      newX = Math.max(0, Math.min(newX, cardRect.width - elRect.width));
      newY = Math.max(0, Math.min(newY, cardRect.height - elRect.height));
      el.style.left = newX + "px";
      el.style.top = newY + "px";
      el.style.bottom = "auto";
    };

    const endDrag = () => {
      if (isDragging) {
        isDragging = false;
        el.classList.remove("dragging");
      }
    };

    el.addEventListener("mousedown", (e) => {
      startDrag(e.clientX, e.clientY);
      e.preventDefault();
    });
    document.addEventListener("mousemove", (e) =>
      moveDrag(e.clientX, e.clientY)
    );
    document.addEventListener("mouseup", endDrag);
    el.addEventListener("touchstart", (e) => {
      startDrag(e.touches[0].clientX, e.touches[0].clientY);
      e.preventDefault();
    });
    document.addEventListener("touchmove", (e) => {
      if (isDragging) moveDrag(e.touches[0].clientX, e.touches[0].clientY);
    });
    document.addEventListener("touchend", endDrag);
  });
}

// Click to snow
function setupClickToSnow() {
  const card = document.getElementById("card");
  card.addEventListener("click", (e) => {
    if (
      e.target.closest(".present") ||
      e.target.closest(".control-btn") ||
      e.target.closest(".tree-container") ||
      e.target.closest(".snowman") ||
      e.target.closest(".message-input")
    )
      return;
    createSnowBurst(e.clientX, e.clientY);
  });
}

// Custom message
function setupCustomMessage() {
  const input = document.getElementById("messageInput");
  const display = document.getElementById("customMessage");
  const defaultMessage =
    "Wishing you joy, peace, and happiness this holiday season ✨";

  input.addEventListener("input", () => {
    display.textContent = input.value || defaultMessage;
  });

  input.addEventListener("keydown", (e) => {
    if (e.key === "Enter") input.blur();
  });
}

// Christmas countdown
function updateCountdown() {
  const now = new Date();
  const year =
    now.getMonth() === 11 && now.getDate() > 25
      ? now.getFullYear() + 1
      : now.getFullYear();
  const christmas = new Date(year, 11, 25);
  const diff = christmas - now;

  const countdownEl = document.getElementById("countdown");

  if (diff <= 0) {
    countdownEl.textContent = "🎄 Merry Christmas! 🎄";
  } else {
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));

    if (days === 0) {
      countdownEl.textContent = `🎅 Christmas in ${hours} hours! 🎁`;
    } else if (days === 1) {
      countdownEl.textContent = "🎅 Christmas is tomorrow! 🎁";
    } else {
      countdownEl.textContent = `🎅 ${days} days until Christmas! 🎁`;
    }
  }
}

// Music (Web Audio API - simple jingle)
function setupMusic() {
  const btn = document.getElementById("musicBtn");

  btn.addEventListener("click", (e) => {
    e.stopPropagation();
    musicPlaying = !musicPlaying;
    btn.innerHTML = musicPlaying ? "🔊 Music" : "🔇 Music";
    btn.classList.toggle("playing", musicPlaying);

    if (musicPlaying) {
      playJingle();
    } else {
      stopJingle();
    }
  });
}

let jingleInterval = null;
function playJingle() {
  if (!audioContext) {
    audioContext = new (window.AudioContext || window.webkitAudioContext)();
  }

  const notes = [
    { freq: 330, dur: 0.2 },
    { freq: 330, dur: 0.2 },
    { freq: 330, dur: 0.4 },
    { freq: 330, dur: 0.2 },
    { freq: 330, dur: 0.2 },
    { freq: 330, dur: 0.4 },
    { freq: 330, dur: 0.2 },
    { freq: 392, dur: 0.2 },
    { freq: 262, dur: 0.2 },
    { freq: 294, dur: 0.2 },
    { freq: 330, dur: 0.6 },
  ];

  let noteIndex = 0;
  let time = audioContext.currentTime;

  function playNote() {
    if (!musicPlaying) return;

    const note = notes[noteIndex % notes.length];
    const osc = audioContext.createOscillator();
    const gain = audioContext.createGain();

    osc.connect(gain);
    gain.connect(audioContext.destination);

    osc.frequency.value = note.freq;
    osc.type = "sine";

    gain.gain.setValueAtTime(0.15, audioContext.currentTime);
    gain.gain.exponentialRampToValueAtTime(
      0.01,
      audioContext.currentTime + note.dur
    );

    osc.start(audioContext.currentTime);
    osc.stop(audioContext.currentTime + note.dur);

    noteIndex++;
    if (noteIndex >= notes.length) noteIndex = 0;
  }

  playNote();
  jingleInterval = setInterval(playNote, 300);
}

function stopJingle() {
  if (jingleInterval) {
    clearInterval(jingleInterval);
    jingleInterval = null;
  }
}

// Update year
document.querySelector(".year").textContent = new Date().getFullYear();

// Initialize everything
createSnowflakes();
createOrnaments();
createLights();
setupSparkleTrail();
setupTreeShake();
setupLightSwitch();
setupPresents();
setupDraggable(".present");
setupDraggable(".snowman");
setupClickToSnow();
setupCustomMessage();
setupMusic();
updateCountdown();
startShootingStars();
startSantaFlights();

setInterval(updateCountdown, 60000);
