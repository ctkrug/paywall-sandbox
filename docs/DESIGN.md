# Design direction — Tollgate landing page

Tollgate is a CLI, so `site/` is a marketing landing page, not an app build. This
file is the single source of truth for how that page looks. Product and page are
one brand.

## 1. Aesthetic direction

Tollgate is a technical blueprint rendered in ink: a near-black slate canvas,
precise monospace type for anything the machine says, and a single warm amber
accent that reads like a coin dropped at a gate. Calm, exact, developer-native.
No gradients-for-decoration, no stock icons.

## 2. Tokens

| Token          | Value       | Use                                  |
|----------------|-------------|--------------------------------------|
| `--bg`         | `#0e1116`   | page background (slate, not pure #000)|
| `--surface-1`  | `#161b22`   | cards, panels                        |
| `--surface-2`  | `#1e242d`   | raised/hover surfaces, code blocks   |
| `--border`     | `#2a313b`   | hairline separators                  |
| `--text`       | `#e8ebf0`   | body + headings                      |
| `--muted`      | `#9aa5b4`   | secondary text                       |
| `--accent`     | `#f2b544`   | wordmark mark, links, primary CTA    |
| `--accent-2`   | `#3fbfa3`   | "settled / 200" success cues         |
| `--danger`     | `#e5645b`   | "402 / challenged" cues              |

- **Type pairing:** display + code = **JetBrains Mono** (wordmark, headings,
  terminal); UI/body = **Inter**. System fallbacks: `ui-monospace, monospace`
  and `system-ui, sans-serif`.
- **Type scale:** ~1.25 ratio. Body 16px, up to a ~44px hero on desktop.
- **Spacing:** 8px scale (8/16/24/32/48/64).
- **Radius:** 10px cards, 8px buttons, 6px inline code.
- **Elevation:** soft `0 1px 0` hairline + a low shadow on cards; the CTA gets a
  subtle amber glow.
- **Motion:** UI transitions 160ms ease-out (hover, focus, press). Respect
  `prefers-reduced-motion`.

## 3. Layout intent

- **Hero (the star):** a real, colorized terminal transcript of the
  `request` command walking a 402 challenge to a settled 200. It sits beside the
  headline on desktop (two columns at 1440px) and stacks under it on phone. The
  transcript is the proof, so it gets real estate, not a thumbnail.
- **1440×900:** two-column hero, then a full-width feature band (3 cards), an
  install/usage strip, an SEO + FAQ section, footer. No dead background.
- **390×844:** single column, wordmark top-left, hero copy then transcript,
  everything else stacked. No horizontal scroll.

## 4. Signature detail

The hero transcript is syntax-colored the way the tool actually speaks: the
`402` line in `--danger`, the settled `200` line in `--accent-2`, headers and
descriptor fields dimmed. It reads like a screenshot of the CLI, but it is live
HTML, so it stays crisp at any zoom. The wordmark carries a small amber gate/coin
monogram in SVG.

## 5. Anti-slop

No em-dashes in copy. None of the banned buzzwords. Every feature names a real,
checkable capability of the tool. System-font-only, unstyled controls, and a tiny
widget adrift in empty space are all rejects.
