# PomoGo Focus Companion Implementation Pipeline

> Build the product described in `docs/00-product-vision.md` and researched in
> `docs/07-product-research.md`: a beautiful terminal deep-work companion where Pomodoro is
> the internal engine, not the product.

Status legend: `[ ]` todo · `[x]` done · `⚠` blocked

## Working Rules

1. Work top to bottom.
2. Finish one task ID at a time.
3. A task is done only after its verify command passes.
4. Keep the app fast: no daemon, no network, no idle busy loop.
5. Add visual variants only when they have a distinct purpose.

---

## Phase F0 — Product Semantics

**Goal:** make the two-mode model exact before adding more visuals.

- [x] **F0.1 — Quick Focus auto-cycle semantics**
  - Do: Quick Focus defaults to a continuing 25/5 cycle. Add `[quick_focus].auto_advance`
    with default `true`; keep it configurable. Ensure the UI keeps ticking after an
    auto-advanced segment.
  - Verify: `go test ./internal/session ./internal/config ./internal/ui` and `go test ./...`.

- [x] **F0.2 — Deep Focus config semantics**
  - Do: add `[deep_focus].default_duration`, use deep-specific work/break durations when
    building deep blocks, and make config init write the modern sections.
  - Verify: config round-trip tests and manual start of a configured 2h block.

- [x] **F0.3 — Active block identity in statefile**
  - Do: write the real active `block_id` to statefile v2 and restore it so a resumed block can
    finish the correct database row.
  - Verify: restore test asserts `currentBlockID` survives a deep block restart.

---

## Phase F1 — Visual Direction

**Goal:** make the app screenshot-worthy without visual junk.

- [ ] **F1.1 — Visual target sheet**
  - Do: add `docs/08-visual-targets.md` with layout jobs, density rules, glyph rules, and
    screenshots/GIF requirements.
  - Verify: document exists and every planned new layout has a reason to exist.

- [ ] **F1.2 — Width-safe icon text**
  - Do: add one width helper for emoji/nerd-font alignment before adding more icon-heavy UI.
  - Verify: golden test with emoji project icons and plain ASCII names.

- [ ] **F1.3 — Theme expansion**
  - Do: add 5 researched palettes from published specs. Keep names stable and cite palette
    sources in comments.
  - Verify: theme tests and layout sweep render every theme.

- [ ] **F1.4 — New layout variants**
  - Do: add distinct layouts only: `dashboard`, `monolith`, `tinybar`, and one experimental
    terminal-rice layout if it passes the visual target sheet.
  - Verify: golden zen/running/idle frames at 80x24 and sweep test at 120x32.

---

## Phase F2 — Polish & Proof

**Goal:** prove the product feels good and stays lightweight.

- [ ] **F2.1 — Demo capture pipeline**
  - Do: update `contrib/demo.tape` for Quick Focus, Deep Focus, theme/layout cycling, and
    screenshot mode.
  - Verify: VHS capture succeeds locally.

- [ ] **F2.2 — Performance gate**
  - Do: run `contrib/bench.sh` with effects off and on; record final numbers.
  - Verify: RSS < 15 MB, CPU < 1%, cold start < 50 ms.

- [ ] **F2.3 — README refresh**
  - Do: rewrite README around Quick Focus / Deep Focus / screenshot identity with real captures.
  - Verify: links and local images resolve.
