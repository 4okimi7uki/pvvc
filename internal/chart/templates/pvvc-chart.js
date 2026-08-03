// PVVC チャートのクライアント描画。
//
// 元データは HTML 内の <script id="__PVVC_DATA__"> の JSON。
// 形は internal/chart.PageData（range / days[].{date,cost,pv,topPages[]} / origin）。
//
// 描画ロジックは Go 側の internal/chart（svg.go / scale.go / theme.go / format.go）の
// 移植。--svg で吐く静的 SVG と見た目を合わせてある。こちらはブラウザで触るための
// インタラクティブ版で、バーをクリックするとその日の上位ページを下に展開する。
import { h, render } from "https://esm.sh/preact";
import { useState, useEffect, useRef } from "https://esm.sh/preact/hooks";
import htm from "https://esm.sh/htm";

const html = htm.bind(h);

// --- 寸法・色（theme.go の既定値）---
const WIDTH = 1369;
// 凡例は SVG の外（カードのヘッダー）に出したので、PAD.top は右軸の "PV" ラベルと
// ピーク PV の吹き出しが収まるぶんだけ。PLOT_H（プロットの高さ）は 224。
// "PV" は baseline がプロット上端の -18、font-size 14 なので上に 29px 必要。
const HEIGHT = 292;
const PAD = { top: 34, right: 78, bottom: 34, left: 58 };
const TICKS = 5;
const BAR_RATIO = 0.56;
const MAX_LINE_DOTS = 14;

const PLOT_W = WIDTH - PAD.left - PAD.right; // 1233
const PLOT_H = HEIGHT - PAD.top - PAD.bottom; // 224

const BAR_COLOR = "#0072f5";
const LINE_COLOR = "#e06c00";
const GRID_COLOR = "#ebebeb";
const TEXT_COLOR = "#8f8f8f";
const DOT_FILL = "#fff";
// 軸ラベルは等幅（page.tmpl.html の --font-mono と同じ）。数字が揃う。
const FONT_FAMILY =
  "'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, monospace";

const BARS_NAME = "Vercel cost / day (USD)";
const LINE_NAME = "GA4 pageviews";

// --- フォーマッタ（format.go の移植）---

// plain は指数表記にしない素の数値。Go の FormatFloat('f', -1) 相当。
const plain = (v) => {
  const s = String(v);
  return s.includes("e") ? v.toFixed(6).replace(/\.?0+$/, "") : s;
};

const usd = (v) => "$" + plain(v);

// comma は "859,241" 形式。
const comma = (v) => {
  const neg = v < 0;
  const s = String(Math.round(Math.abs(v)));
  let out = "";
  for (let i = 0; i < s.length; i++) {
    if (i > 0 && (s.length - i) % 3 === 0) out += ",";
    out += s[i];
  }
  return (neg ? "-" : "") + out;
};

// compact は "200k" / "1M" 形式。
const compact = (v) =>
  v >= 1_000_000
    ? plain(v / 1_000_000) + "M"
    : v >= 1_000
      ? plain(v / 1_000) + "k"
      : plain(v);

// num は座標値。整数なら小数を落とし、そうでなければ 2 桁（svg.go の num）。
const num = (v) => (Number.isInteger(v) ? String(v) : Number(v.toFixed(2)));

const MONTHS = [
  "Jan",
  "Feb",
  "Mar",
  "Apr",
  "May",
  "Jun",
  "Jul",
  "Aug",
  "Sep",
  "Oct",
  "Nov",
  "Dec",
];
const WEEKDAYS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

// "2026-07-22" -> "Jul 22"
const monthDay = (iso) => {
  const [, m, d] = iso.split("-").map(Number);
  return `${MONTHS[m - 1]} ${d}`;
};

// --- スケール（scale.go の移植）---
const NICE = [1, 1.5, 2, 2.5, 5, 7.5, 10];

function makeScale(dataMax, ticks, height) {
  if (ticks <= 0) ticks = 5;
  let step, max;
  if (!(dataMax > 0)) {
    max = 1;
    step = 1 / ticks;
  } else {
    const raw = dataMax / ticks;
    const mag = Math.pow(10, Math.floor(Math.log10(raw)));
    step = 10 * mag;
    for (const m of NICE) {
      if (raw <= m * mag) {
        step = m * mag;
        break;
      }
    }
    max = step * ticks;
  }
  return {
    max,
    step,
    y: (v) => height - (v / max) * height,
    values: () => Array.from({ length: ticks + 1 }, (_, i) => step * i),
  };
}

const seriesMax = (arr) =>
  arr.reduce((m, v) => (isFinite(v) && v > m ? v : m), 0);
const plotValue = (v) => (v < 0 ? 0 : v);

// xLabels は等間隔に最大 limit 個。両端は必ず含む（report/chart.go の移植）。
function xLabels(days, limit = 8) {
  const n = days.length;
  if (n === 0) return [];
  if (limit < 2 || n <= limit) {
    return days.map((d, i) => ({ at: i, text: monthDay(d.date) }));
  }
  const step = (n - 1) / (limit - 1);
  const out = [];
  let prev = -1;
  for (let k = 0; k < limit; k++) {
    const i = Math.round(k * step);
    if (i === prev) continue;
    prev = i;
    out.push({ at: i, text: monthDay(days[i].date) });
  }
  return out;
}

// ホバー時の <title>（slotTitles の移植）。
function slotTitle(d) {
  return `${dayWithWeekday(d.date)}\n$${Number(d.cost).toFixed(2)}  |  ${comma(d.pv)} PV`;
}

// --- チャート本体 ---
function Graph({ days, selected, onSelect }) {
  const n = days.length;
  if (n === 0) return null;

  const slot = PLOT_W / n;
  const barW = slot * BAR_RATIO;
  const centerX = (i) => i * slot + slot / 2;

  const left = makeScale(seriesMax(days.map((d) => d.cost)), TICKS, PLOT_H);
  const right = makeScale(seriesMax(days.map((d) => d.pv)), TICKS, PLOT_H);

  const labels = xLabels(days, 8);
  const selIdx =
    selected == null ? -1 : days.findIndex((d) => d.date === selected);

  // 折れ線の点と、最大 PV の点。
  const pts = days.map((d, i) => ({
    x: centerX(i),
    y: right.y(plotValue(d.pv)),
    v: d.pv,
  }));
  let top = pts[0];
  for (const p of pts) if (p.v > top.v) top = p;

  return html`
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 ${WIDTH} ${HEIGHT}"
      width=${WIDTH}
      height=${HEIGHT}
      font-family=${FONT_FAMILY}
      role="img"
      aria-label="Vercel daily cost with GA4 pageviews overlaid"
    >
      <g transform="translate(${PAD.left},${PAD.top})">
        <!-- 選択中の日の帯 -->
        ${
          selIdx >= 0 &&
          html`<rect
            x=${num(selIdx * slot)}
            y="0"
            width=${num(slot)}
            height=${num(PLOT_H)}
            fill=${LINE_COLOR}
            opacity="0.07"
          />`
        }

        <!-- 左軸（コスト）: グリッド + 目盛り -->
        <g font-size="14" fill=${BAR_COLOR}>
          ${left.values().map(
            (v) =>
              html`<g>
                <line
                  x1="0"
                  y1=${num(left.y(v))}
                  x2=${num(PLOT_W)}
                  y2=${num(left.y(v))}
                  stroke=${GRID_COLOR}
                  stroke-width="1"
                  shape-rendering="crispEdges"
                />
                <text x="-10" y=${num(left.y(v) + 4)} text-anchor="end"
                  >${usd(v)}</text
                >
              </g>`,
          )}
        </g>

        <!-- 右軸（PV） -->
        <g font-size="14" fill=${LINE_COLOR}>
          <text x=${num(PLOT_W + 11)} y="-18" font-size="14" opacity="0.8">
            PV
          </text>
          ${right.values().map(
            (v) =>
              html`<g>
                <line
                  x1=${num(PLOT_W)}
                  y1=${num(right.y(v))}
                  x2=${num(PLOT_W + 5)}
                  y2=${num(right.y(v))}
                  stroke=${LINE_COLOR}
                  stroke-width="1"
                  opacity="0.5"
                />
                <text x=${num(PLOT_W + 11)} y=${num(right.y(v) + 4)}
                  >${compact(v)}</text
                >
              </g>`,
          )}
        </g>

        <!-- 棒（コスト） -->
        <g>
          ${days.map((d, i) => {
            const x = i * slot + (slot - barW) / 2;
            const y = left.y(plotValue(d.cost));
            const op =
              selected == null ? 0.45 : d.date === selected ? 0.9 : 0.2;
            return html`<rect
              x=${num(x)}
              y=${num(y)}
              width=${num(barW)}
              height=${num(PLOT_H - y)}
              fill=${BAR_COLOR}
              fill-opacity=${op}
            />`;
          })}
        </g>

        <!-- 折れ線（PV） -->
        <polyline
          points=${pts.map((p) => `${num(p.x)},${num(p.y)}`).join(" ")}
          fill="none"
          stroke=${LINE_COLOR}
          stroke-width="2"
          stroke-linejoin="round"
          stroke-linecap="round"
        />
        ${
          n <= MAX_LINE_DOTS &&
          pts.map(
            (p) =>
              html`<circle
                cx=${num(p.x)}
                cy=${num(p.y)}
                r="3"
                fill=${DOT_FILL}
                stroke=${LINE_COLOR}
                stroke-width="2"
              />`,
          )
        }
        <text
          x=${num(top.x)}
          y=${num(top.y - 10)}
          text-anchor="middle"
          font-size="15"
          font-weight="600"
          fill=${LINE_COLOR}
          stroke="white"
          stroke-width="2"
          paint-order="stroke fill"
        >
          ${comma(top.v)}
        </text>

        <!-- X 軸 -->
        <g font-size="15" fill=${TEXT_COLOR}>
          <line
            x1="0"
            y1=${num(PLOT_H)}
            x2=${num(PLOT_W)}
            y2=${num(PLOT_H)}
            stroke=${GRID_COLOR}
            shape-rendering="crispEdges"
          />
          ${labels.map(
            (l) =>
              html`<text
                x=${num(centerX(l.at))}
                y=${num(PLOT_H + 20)}
                text-anchor="middle"
              >
                ${l.text}
              </text>`,
          )}
        </g>

        <!-- クリック用の当たり判定（列の全高） -->
        <g>
          ${days.map(
            (d, i) =>
              html`<rect
                x=${num(i * slot)}
                y="0"
                width=${num(slot)}
                height=${num(PLOT_H)}
                fill="transparent"
                style="cursor:pointer"
                onClick=${() => onSelect(d.date)}
              >
                <title>${slotTitle(d)}</title>
              </rect>`,
          )}
        </g>
      </g>
    </svg>
  `;
}

// --- 下部の上位ページ（アコーディオン）---
function PageList({ days, origin, selected, onSelect }) {
  const openRef = useRef(null);

  // 選択が変わったら、展開された行をリスト内でスクロールして見せる。
  // scrollIntoView は縦方向にページ全体（window）まで動かしてしまうので使わない。
  // 縦スクロール可能な直近の祖先を自前で探し、その容器の scrollTop だけを動かす。
  useEffect(() => {
    const activeRow = openRef.current;
    if (!selected || !activeRow) return;

    // overflow-y が auto/scroll で、かつ実際に溢れている祖先を探す。
    // overflow だけだと、溢れてはいるがスクロールしない内側コンテナを掴んで
    // 動かないことがあるため scrollHeight > clientHeight まで見る。
    // body に達したら打ち切り = ページ自体はスクロールさせない。
    let box = activeRow.parentElement;
    while (box && box !== document.body) {
      const { overflowY } = window.getComputedStyle(box);
      if (
        (overflowY === "auto" || overflowY === "scroll") &&
        box.scrollHeight > box.clientHeight
      ) {
        break;
      }
      box = box.parentElement;
    }
    if (!box || box === document.body) return;

    const boxRect = box.getBoundingClientRect();
    const rowRect = activeRow.getBoundingClientRect();
    const rowOffsetTop = rowRect.top - boxRect.top + box.scrollTop;

    // 容器の中央に来るように寄せる。
    box.scrollTo({
      top: rowOffsetTop,
      behavior: "instant",
    });
  }, [selected]);

  return html`
    <section class="pages">
      <div class="pages__head">
        <p class="pages__title">日別データ</p>
        <div class="pages__keys">
          <span class="cost">Cost</span>
          <span class="pv">Pageviews</span>
        </div>
      </div>
      <div class="pages__body">
        ${[...days].reverse().map((d) => {
          const open = selected === d.date;
          const pages = d.topPages || [];
          return html`<div
            ref=${open ? openRef : null}
            class=${"day" + (open ? " is-open" : "")}
            key=${d.date}
          >
            <button class="day__head" onClick=${() => onSelect(d.date)}>
              <div class="day__date">
                <span class="day__caret">▶</span>
                <span class="day__dateText">${d.date}</span>
                ${
                  pages.length > 0 &&
                  html`<span class="day__tag">Top ${pages.length}</span>`
                }
              </div>
              <div class="day__meta">
                <span class="cost">$${Number(d.cost).toFixed(2)}</span>
                <span class="pv">${comma(d.pv)} PV</span>
              </div>
            </button>
            ${
              open &&
              html`<div class="day__list">
                ${
                  pages.length === 0
                    ? html`<p class="day__empty">ページ別データなし</p>`
                    : html`
                        <div class="day__listHead">
                          <span>Path</span><span>Views</span>
                        </div>
                        <ul>
                          ${pages.map(
                            (p) =>
                              html`<li key=${p.path}>
                                <a
                                  href=${origin + p.path}
                                  target="_blank"
                                  rel="noopener"
                                >
                                  <span class="day__path">${p.path}</span>
                                  <span class="day__views"
                                    >${comma(p.views)}</span
                                  >
                                </a>
                              </li>`,
                          )}
                        </ul>
                      `
                }
              </div>`
            }
          </div>`;
        })}
      </div>
    </section>
  `;
}

// --- 上部の統計カード ---

// weekday は "2026-06-08" -> "Mon"。
const weekday = (iso) => WEEKDAYS[new Date(iso + "T00:00:00").getDay()];

// dayWithWeekday は "2026-06-08 (Mon)"。カードの補足行に使う。
const dayWithWeekday = (iso) => `${iso} (${weekday(iso)})`;

// summarize は days から統計カードの中身を組み立てる。
function summarize(days, range) {
  const n = days.length;
  if (n === 0) return [];

  const totalCost = days.reduce((s, d) => s + Number(d.cost), 0);
  const peakPv = days.reduce((a, d) => (Number(d.pv) > Number(a.pv) ? d : a));
  const peakCost = days.reduce((a, d) =>
    Number(d.cost) > Number(a.cost) ? d : a,
  );
  const from = range ? range.from : days[0].date;
  const to = range ? range.to : days[n - 1].date;

  return [
    {
      label: "計測期間",
      value: comma(n),
      unit: "日",
      sub: `${from} → ${to}`,
    },
    {
      label: "トータルコスト",
      value: "$" + comma(totalCost),
      sub: `${comma(n)}日間合計`,
    },
    {
      label: "ピーク PV",
      value: comma(peakPv.pv),
      sub: dayWithWeekday(peakPv.date),
    },
    {
      label: "ピークコスト",
      value: "$" + Number(peakCost.cost).toFixed(2),
      unit: "/day",
      sub: dayWithWeekday(peakCost.date),
    },
  ];
}

function Stats({ days, range }) {
  const cards = summarize(days, range);
  if (cards.length === 0) return null;

  return html`
    <div class="stats">
      ${cards.map(
        (s) =>
          html`<div class="stat" key=${s.label}>
            <p class="stat__label">${s.label}</p>
            <p class="stat__value">
              ${s.value}${s.unit && html`<span class="stat__unit">${s.unit}</span>`}
            </p>
            <p class="stat__sub">${s.sub}</p>
          </div>`,
      )}
    </div>
  `;
}

function App({ range, days = [], origin = "" }) {
  const [selected, setSelected] = useState(null);
  const toggle = (date) => setSelected((s) => (s === date ? null : date));

  const period =
    range &&
    (range.from === range.to ? range.from : `${range.from} 〜 ${range.to}`);

  const avgCost = days.length
    ? days.reduce((s, d) => s + Number(d.cost), 0) / days.length
    : 0;

  return html`
    <div class="container">
      <header class="pageHeader">
        <div>
          <h1 class="pageHeader__title">PVVC Chart</h1>
          ${period && html`<p class="pageHeader__period">${period}</p>`}
        </div>
      </header>

      <${Stats} days=${days} range=${range} />

      <section class="card">
        <div class="card__head">
          <div class="legend">
            <span><span class="legend__bar"></span>${BARS_NAME}</span>
            <span>
              <svg width="20" height="8" viewBox="0 0 20 8" aria-hidden="true">
                <line
                  x1="0"
                  y1="4"
                  x2="20"
                  y2="4"
                  stroke=${LINE_COLOR}
                  stroke-width="2"
                />
                <circle cx="10" cy="4" r="2.5" fill=${LINE_COLOR} />
              </svg>
              ${LINE_NAME}
            </span>
          </div>
          <span class="mono">avg $${avgCost.toFixed(2)}/day</span>
        </div>
        <div class="chartWrapper">
          <${Graph} days=${days} selected=${selected} onSelect=${toggle} />
        </div>
        <div class="card__foot">
          <p>棒グラフ: 左軸 (USD) ／ 折れ線: 右軸 (PV数)</p>
          <p class="mono">Source: Vercel billing + GA4</p>
        </div>
      </section>

      <${PageList}
        days=${days}
        origin=${origin}
        selected=${selected}
        onSelect=${toggle}
      />
    </div>
  `;
}

// --- マウント ---
const el = document.getElementById("__PVVC_DATA__");
const root = document.getElementById("root");
if (el && root) {
  const data = JSON.parse(el.textContent);
  render(html`<${App} ...${data} />`, root);
}

export { App, Graph, PageList, Stats, makeScale, summarize, xLabels };
