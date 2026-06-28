# Implementation Notes

## XLSX formulas ported

- Source values were taken from `docs/xlsx-analysis.md`, then checked against the `Программа` sheet ranges `G4:I21`, validation ranges `Z4:AF13`, and helper cells `Z15:AB17`.
- Heavy lift base percent:
  - `Вариант 1`: `0.82`
  - `Вариант 2`: `0.86`
- XLSX-compatible progression step is `4%` of the relevant 1RM, rounded with `MROUND(..., 2.5)`.
- Weekly heavy progression:
  - Week 1: `base - step * 4`
  - Week 2: `base - step * 3`
  - Week 3: `base - step * 2`
  - Week 4: `base - step`
  - Week 5: `base`
  - Week 6: `base + step`
  - Week 7: `base + step * 2`
  - Week 8: `1ПМ`
- Heavy lift reps:
  - Weeks 1-5: `5x5`
  - Week 6: `3x3`
  - Week 7: `2x2`
  - Week 8: `Тест`
- Light bench and light squat/deadlift percentages were copied from the visible formulas in rows `G6:I6`, `G12:I12`, and `G18:I18`.
- Assistance and GPP set/RPE rules were copied from rows `G7:I9`, `G13:I15`, and `G19:I21`.

## XLSX compatibility issues

- The workbook has a known `AA17` issue: helper formulas in `Z15:Z17` check `A23=AA17`, but `AA17` is empty. As a result, the 5% branch is never selected in the XLSX.
- This MVP keeps the progression step fixed at `4%` for XLSX compatibility. A future explicit `5%` setting should be a separate cycle/profile choice, not automatically tied to `Вариант 2`.
- The XLSX assistance defaults are prompt strings such as `Кликни чтобы выбрать доп. тягу`. The MVP defaults to concrete valid selections so the rendered plan is immediately usable.
- The XLSX contains the typo `Становя тяга с паузами`; the option is preserved as-is in the UI label.

## Dataset mapping limits

- `hasaneyldrm/exercises-dataset` was used only to verify program-relevant exercises. The client does not include a catalog.
- Confirmed exercise details are kept in `client/src/data/exerciseDetails.ts`.
- The XLSX-to-dataset alias layer is kept separately in `client/src/data/exerciseAliases.ts`.
- Mappings marked `needs_review` or `missing` intentionally show a calm fallback instead of a guessed GIF.
- Known unconfirmed or limited mappings:
  - `Тяга из ямы`: no exact deficit deadlift entry confirmed.
  - `Становя тяга с паузами`: no exact paused deadlift entry confirmed.
  - `Приседания с высоким грифом` / `Приседания с низком грифом`: dataset search did not confirm the high-bar vs low-bar distinction.
  - `Упражнение на пресс`: XLSX is generic and does not name a concrete exercise.
  - `Рычажная вертикальная тяга`: no exact lever vertical row entry confirmed.
  - Generic `Бицепс` and `Трицепс` are mapped to local/isolation examples from the dataset, but the XLSX does not specify exact movements.

## MVP persistence and auth boundary

- Settings are saved to `localStorage` under `linear-cycle-program-selection`.
- On app initialization the client sets a JavaScript-readable `init=1` cookie with `Path=/` and `SameSite=Lax`.
- `init=1` is only a temporary bootstrap marker. It is not authentication or authorization.
- There is no backend, database, router, account model, or real auth in this MVP.
