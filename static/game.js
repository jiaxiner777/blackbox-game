/* ============================================================
   SKYLINE TYCOON — 前端游戏逻辑
   ============================================================ */

'use strict';

// ─── 全局状态 ───────────────────────────────────────────────
let gameState    = null;   // 来自后端的最新游戏状态
let selectedIdx  = null;   // 第一次点击选中的柱子索引
let lbPollTimer  = null;   // 排行榜轮询定时器

// ─── 工具函数 ───────────────────────────────────────────────
function $(id) { return document.getElementById(id); }

async function apiPost(path, body = {}) {
  const res = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

async function apiGet(path) {
  const res = await fetch(path);
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  return res.json();
}

// ─── 开始游戏 ───────────────────────────────────────────────
async function startGame() {
  const nameEl = $('player-name-input');
  // 名字优先取输入框，其次取当前状态里的名字（重开时）
  const name = (nameEl && nameEl.value.trim()) || (gameState && gameState.player_name) || '无名侠';
  try {
    gameState = await apiPost('/api/start', { name });
    showScreen('game-screen');
    renderGame();
    checkOverlay();   // ← 修复：初始排列就是最大值时立即弹窗
    startLbPolling();
  } catch (e) {
    alert('连接服务器失败，请刷新重试。');
  }
}

// ─── 点击柱子（选择/交换）──────────────────────────────────
async function onColumnClick(idx) {
  if (!gameState || gameState.status !== 'playing') return;

  if (selectedIdx === null) {
    // 第一次点击：选中该柱子
    selectedIdx = idx;
    renderColumns();
  } else if (selectedIdx === idx) {
    // 再次点击同一柱子：取消选中
    selectedIdx = null;
    renderColumns();
  } else {
    // 第二次点击不同柱子：执行交换
    const idx1 = selectedIdx;
    const idx2 = idx;
    selectedIdx = null;

    flashColumns(idx1, idx2);

    try {
      gameState = await apiPost('/api/swap', { idx1, idx2 });
      renderGame();

      if (gameState.status === 'won_level') {
        showWinOverlay();
      } else if (gameState.status === 'game_over') {
        showGameOverOverlay();
        refreshLeaderboard();
      }
    } catch (e) {
      console.error(e);
    }
  }
}

// ─── 下一关 ─────────────────────────────────────────────────
async function nextLevel() {
  try {
    gameState = await apiPost('/api/next');
    selectedIdx = null;
    hideOverlay();
    renderGame();
    checkOverlay();   // ← 修复：新关卡初始就通关时弹窗
  } catch (e) {
    console.error(e);
  }
}

// ─── 公共：检查状态并弹出对应弹窗 ──────────────────────────
function checkOverlay() {
  if (!gameState) return;
  if (gameState.status === 'won_level') {
    // 稍微延迟，让柱子揭晓动画先跑完
    setTimeout(showWinOverlay, 500);
  } else if (gameState.status === 'game_over') {
    setTimeout(() => { showGameOverOverlay(); refreshLeaderboard(); }, 500);
  }
}

// ─── 退出游戏 ────────────────────────────────────────────────
async function quitGame() {
  try {
    await apiPost('/api/quit');
  } catch (_) {}
  gameState = null;
  selectedIdx = null;
  stopLbPolling();
  hideOverlay();
  showScreen('start-screen');
  refreshStartLeaderboard();
}

// ─── 弹窗按钮回调 ─────────────────────────────────────────
function overlayPrimary() {
  if (!gameState) return;
  if (gameState.status === 'won_level') {
    nextLevel();
  } else if (gameState.status === 'game_over') {
    // 失败后重开：保持同一玩家名，重新请求 /api/start
    const name = gameState.player_name || '无名侠';
    hideOverlay();
    apiPost('/api/start', { name }).then(state => {
      gameState = state;
      selectedIdx = null;
      renderGame();
      checkOverlay();
    }).catch(console.error);
  }
}

function overlaySecondary() { quitGame(); }

// ─── 渲染整个游戏状态 ─────────────────────────────────────
function renderGame() {
  if (!gameState) return;

  // HUD
  $('hud-player').textContent = gameState.player_name;
  $('hud-level').textContent  = String(gameState.level).padStart(2, '0');

  const stepsEl = $('hud-steps');
  stepsEl.textContent = gameState.current_steps;
  stepsEl.className = 'hud-val neon-green';
  if (gameState.current_steps <= 3) {
    stepsEl.className = 'hud-val neon-pink steps-warn';
  } else if (gameState.current_steps <= 7) {
    stepsEl.className = 'hud-val neon-orange';
  }

  $('hud-water').textContent = gameState.current_water;

  renderFeedback(gameState.feedback);
  renderColumns();
}

// ─── 渲染反馈文字 ─────────────────────────────────────────
function renderFeedback(text) {
  const el = $('feedback-bar');
  if (!text) { el.textContent = ''; return; }

  let color = '#c8d0e0';
  if (text.includes('涨')) color = '#00ff88';
  else if (text.includes('跌')) color = '#ff2d78';
  else if (text.includes('目标')) color = '#ffe600';
  else if (text.includes('耗尽')) color = '#ff2d78';

  el.style.color = color;
  el.style.textShadow = `0 0 10px ${color}`;
  el.textContent = text;
}

// ─── 渲染柱子 ─────────────────────────────────────────────
function renderColumns() {
  if (!gameState) return;

  const area = $('columns-area');
  const { heights, n, revealed, status } = gameState;
  const MAX_H = 6;
  const BODY_H = 220; // px，与 CSS .col-body height 保持一致

  if (area.children.length !== n) {
    area.innerHTML = '';
    for (let i = 0; i < n; i++) {
      area.appendChild(makeColumnDOM(i));
    }
  }

  const cols = area.querySelectorAll('.col-wrap');
  cols.forEach((col, i) => {
    const isSelected = (i === selectedIdx);
    const isRevealed = revealed;
    const h = heights[i];
    const fillH = Math.round((h / MAX_H) * BODY_H);

    col.className = 'col-wrap';
    if (isSelected) col.classList.add('selected');
    if (isRevealed) col.classList.add('revealed');

    const fill = col.querySelector('.col-fill');
    fill.style.height = isRevealed ? `${fillH}px` : '0px';

    const question = col.querySelector('.col-question');
    const num      = col.querySelector('.col-num');
    if (isRevealed) {
      question.style.display = 'none';
      num.style.display = 'flex';
      num.textContent = h;
    } else {
      question.style.display = 'flex';
      num.style.display = 'none';
    }

    col.style.cursor = status !== 'playing' ? 'default' : 'pointer';
  });
}

// 创建单个柱子 DOM
function makeColumnDOM(i) {
  const wrap = document.createElement('div');
  wrap.className = 'col-wrap';
  wrap.dataset.idx = i;
  wrap.onclick = () => onColumnClick(i);

  wrap.innerHTML = `
    <div class="col-body">
      <div class="col-fill" style="height:0"></div>
      <div class="col-question">?</div>
      <div class="col-num" style="display:none"></div>
    </div>
    <div class="col-index">${i}</div>
  `;
  return wrap;
}

// ─── 胜利弹窗 ─────────────────────────────────────────────
function showWinOverlay() {
  $('ov-title').innerHTML = `<span class="neon-yellow">🎉 LEVEL CLEAR!</span>`;
  $('ov-sub').textContent = `Level ${gameState.level} 通关成功！`;
  $('ov-info').innerHTML  =
    `💧 目标水量: ${gameState.target_score_reveal}` +
    `<br>⚡ 剩余步数: ${gameState.current_steps}` +
    `<br>🎁 奖励: +5 步`;
  $('ov-btn-primary').textContent = '► NEXT LEVEL';
  document.querySelector('.overlay-box').className = 'overlay-box win';
  showOverlay();
}

// ─── 失败弹窗 ─────────────────────────────────────────────
function showGameOverOverlay() {
  const levelsCleared = gameState.level - 1;
  $('ov-title').innerHTML = `<span class="neon-pink">💀 GAME OVER</span>`;
  $('ov-sub').textContent = levelsCleared > 0
    ? `通过了 ${levelsCleared} 关，成绩已上榜！`
    : '还没有通关记录';
  $('ov-info').innerHTML  =
    `💧 本关目标: ${gameState.target_score_reveal}` +
    `<br>📊 最高通关: Level ${levelsCleared}`;
  $('ov-btn-primary').textContent = '► PLAY AGAIN';
  document.querySelector('.overlay-box').className = 'overlay-box lose';
  showOverlay();
}

function showOverlay() { $('overlay').classList.remove('hidden'); }
function hideOverlay()  { $('overlay').classList.add('hidden'); }

// ─── 屏幕切换 ─────────────────────────────────────────────
function showScreen(id) {
  document.querySelectorAll('.screen').forEach(s => {
    s.classList.toggle('active', s.id === id);
  });
}

// ─── 柱子交换闪烁动画 ────────────────────────────────────
function flashColumns(i1, i2) {
  const area = $('columns-area');
  const cols = area.querySelectorAll('.col-wrap');
  [i1, i2].forEach(i => {
    if (cols[i]) {
      cols[i].classList.add('swapping');
      setTimeout(() => cols[i].classList.remove('swapping'), 500);
    }
  });
}

// ─── 排行榜轮询 ───────────────────────────────────────────
function startLbPolling() {
  refreshLeaderboard();
  lbPollTimer = setInterval(refreshLeaderboard, 5000);
}

function stopLbPolling() {
  clearInterval(lbPollTimer);
  lbPollTimer = null;
}

async function refreshLeaderboard() {
  try {
    const entries = await apiGet('/api/leaderboard');
    renderLeaderboard($('game-lb-list'), entries);
  } catch (_) {}
}

async function refreshStartLeaderboard() {
  try {
    const entries = await apiGet('/api/leaderboard');
    renderLeaderboard($('start-lb-list'), entries);
  } catch (_) {}
}

function renderLeaderboard(container, entries) {
  if (!container) return;
  if (!entries || entries.length === 0) {
    container.innerHTML = '<div class="lb-empty">NO RECORDS YET</div>';
    return;
  }
  const medals = ['🥇', '🥈', '🥉', '4', '5'];
  container.innerHTML = entries.map((e, i) => `
    <div class="lb-entry rank-${i + 1}">
      <span class="lb-rank">${i < 3 ? medals[i] : (i + 1)}</span>
      <span class="lb-name"  title="${escHtml(e.name)}">${escHtml(truncate(e.name, 8))}</span>
      <span class="lb-level">LV.${e.max_level}</span>
    </div>
  `).join('');
}

function truncate(s, n) {
  return [...s].length > n ? [...s].slice(0, n - 1).join('') + '…' : s;
}

function escHtml(s) {
  return s.replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/>/g,'&gt;').replace(/"/g,'&quot;');
}

// ─── 初始化 ───────────────────────────────────────────────
(function init() {
  refreshStartLeaderboard();

  $('player-name-input').addEventListener('keydown', e => {
    if (e.key === 'Enter') startGame();
  });
})();
