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
// HEIGHT と PAD.top は凡例とプロットの間隔を広げるため元の 300 / 26 から +6 した。
// 同じだけ上げているので PLOT_H（プロットの高さ）は 224 のまま変わらない。
const HEIGHT = 306;
const PAD = { top: 32, right: 78, bottom: 50, left: 58 };
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
const FONT_FAMILY =
  "Geist, -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Helvetica Neue', Arial, sans-serif";

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
  const wd = WEEKDAYS[new Date(d.date + "T00:00:00").getDay()];
  return `${d.date} (${wd})\n$${Number(d.cost).toFixed(2)}  |  ${comma(d.pv)} PV`;
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

  const from = days[0].date;
  const to = days[n - 1].date;
  const caption = from === to ? from : `${from} → ${to} (${n} days)`;

  // 折れ線凡例の x。1つ目のラベル幅ぶん空ける。固定値だとフォントを変えたとき
  // ラベルに重なるので、文字数 × フォントサイズから幅をざっくり見積もって置く。
  const LEGEND_FONT = 16;
  const SWATCH = 15 + Math.ceil(BARS_NAME.length * LEGEND_FONT * 0.55) + 24;

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
      <!-- 凡例 -->
      <g
        transform="translate(${PAD.left},16)"
        font-size=${LEGEND_FONT}
        fill=${TEXT_COLOR}
      >
        <rect
          x="0"
          y="-7"
          width="9"
          height="9"
          rx="1.5"
          fill=${BAR_COLOR}
          opacity="0.85"
        />
        <text x="15" y="1">${BARS_NAME}</text>
        <line
          x1=${SWATCH}
          y1="-3"
          x2=${SWATCH + 22}
          y2="-3"
          stroke=${LINE_COLOR}
          stroke-width="2"
        />
        <circle cx=${SWATCH + 11} cy="-3" r="2.8" fill=${LINE_COLOR} />
        <text x=${SWATCH + 28} y="1">${LINE_NAME}</text>
      </g>

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
          <text
            x=${num(PLOT_W)}
            y=${num(PLOT_H + 40)}
            text-anchor="end"
            font-size="12"
            opacity="0.7"
          >
            ${caption}
          </text>
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
      <div>
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
                <span>${open ? "▾ " : "▸ "}</span>${d.date}
              </div>
              <div class="day__meta">
                <div>$${Number(d.cost).toFixed(2)}</div>
                <div>${comma(d.pv)} PV</div>
              </div>
            </button>
            ${
              open &&
              html`<ul class="day__list">
                ${
                  pages.length === 0
                    ? html`<li class="day__empty">上位ページなし</li>`
                    : pages.map(
                        (p) =>
                          html`<li key=${p.path}>
                            <a
                              href=${origin + p.path}
                              target="_blank"
                              rel="noopener"
                            >
                              <div class="day__path">${p.path}</div>
                              <div class="day__views">${comma(p.views)}</div>
                            </a>
                          </li>`,
                      )
                }
              </ul>`
            }
          </div>`;
        })}
      </div>
    </section>
  `;
}

function App({ title = "", range, days = [], origin = "" }) {
  const [selected, setSelected] = useState(null);
  const toggle = (date) => setSelected((s) => (s === date ? null : date));

  const period =
    range &&
    (range.from === range.to ? range.from : `${range.from} 〜 ${range.to}`);

  return html`
    <header class="pageHeader">
      ${title && html`<h1 class="pageHeader__title">PVVC Chart</h1>`}
      ${period && html`<p class="pageHeader__period">${period}</p>`}
    </header>
    <main class="main">
      <div class="chartWrapper">
        <${Graph} days=${days} selected=${selected} onSelect=${toggle} />
      </div>
    </main>
    <${PageList}
      days=${days}
      origin=${origin}
      selected=${selected}
      onSelect=${toggle}
    />
  `;
}

// --- マウント ---
const el = document.getElementById("__PVVC_DATA__");
const root = document.getElementById("root");
if (el && root) {
  const data = JSON.parse(el.textContent);
  render(html`<${App} ...${data} />`, root);
}

export { App, Graph, PageList, makeScale, xLabels };
