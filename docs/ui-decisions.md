# UI Decisions

## Profile vs Cycle

Use two levels of settings:

- Profile: long-lived athlete defaults.
- Cycle: concrete settings used for the current program run.

Profile should contain:

- deadlift 1RM;
- bench press 1RM;
- squat 1RM;
- preferred variant;
- preferred progression step: `4%` or `5%`;
- optional notes.

Cycle should contain a snapshot of:

- 1RM values used for this cycle;
- selected variant;
- selected progression step;
- assistance exercises;
- GPP exercises;
- current week.

Changing the profile should not silently rewrite an active or completed cycle. The UI can offer an explicit action such as "Apply profile values to current cycle".

## Progression Step

Default to `4%` because it matches the current XLSX behavior.

Allow `5%` only as an explicit setting. Do not automatically bind `5%` to `Вариант 2`.

Recommended UI labels:

- `4% - XLSX-compatible`
- `5% - more aggressive`

## Exercise Selection

Do not force the user to choose assistance/GPP exercises every week.

Recommended behavior:

1. User selects exercises when creating or editing the active cycle.
2. The same selections apply across weeks by default.
3. User can edit the active cycle before starting a new week.
4. Completed checkpoints keep prescribed snapshots, so past progress remains stable.

This matches the XLSX flow: exercise selection is a setup/control area, while week switching changes the rendered plan.

If weekly variation is needed later, add it as a separate feature:

- cycle-level default exercises;
- optional per-week overrides;
- clear UI indicator when a week differs from the cycle default.

Do not add per-week overrides in the MVP unless the requirements explicitly need them.
