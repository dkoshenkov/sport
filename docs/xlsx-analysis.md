# XLSX Analysis

Source file:

`/Users/dkoshenkov/Downloads/Telegram Lite/Линейный цикл для продвинутых.xlsx`

## Workbook

Sheets:

- `Инструкция`
- `Программа`

## Instruction Summary

The program is a classic American linear cycle.

User flow from the instruction sheet:

1. Enter current one-rep maximums.
2. Select assistance exercises for the main lifts.
3. Select GPP exercises, or skip GPP for a minimalist version where allowed.
4. Select the program variant.
5. Switch between training weeks.

Variant guidance:

- `Вариант 1` is recommended for most users and for first-time use.
- `Вариант 2` is intended for experienced athletes with low strength endurance.
- A higher variant number does not mean a better or more effective program.

GPP guidance:

- Abs are user choice.
- Arms should use local/isolation exercises.
- Dips and pull-ups are not suitable as biceps/triceps choices.

## Program Inputs

Main one-rep maximums:

| Cell | Meaning | Default |
| --- | --- | --- |
| `B4` | Deadlift 1RM | `225` |
| `B5` | Bench press 1RM | `125` |
| `B6` | Squat 1RM | `170` |

Selections:

| Cell | Meaning |
| --- | --- |
| `G3` | Selected week |
| `A10` | Deadlift assistance |
| `A11` | Bench assistance |
| `A12` | Squat assistance |
| `A15` | GPP horizontal/back row option |
| `A16` | GPP biceps option |
| `A17` | GPP abs option |
| `A18` | GPP triceps option |
| `A19` | GPP vertical/back row option |
| `A20` | GPP overhead/press option |
| `A23:B23` | Program variant |

## Selection Options

Weeks:

- `Неделя 1`
- `Неделя 2`
- `Неделя 3`
- `Неделя 4`
- `Неделя 5`
- `Неделя 6`
- `Неделя 7`
- `Неделя 8`

Variants:

- `Вариант 1`
- `Вариант 2`

Deadlift assistance:

- `Гуд-морнинг`
- `Румынская тяга`
- `Тяга из ямы`
- `Классическая становая тяга`
- `Становая тяга сумо`
- `Становя тяга с паузами`

Bench assistance:

- `Жим узким хватом`
- `Жим обратным хватом`
- `Жим на наклонной скамье`
- `Жим гантелей лежа`
- `Брусья`

Squat assistance:

- `Приседания Зерчера`
- `Приседания со штангой на груди`
- `Приседания с высоким грифом`
- `Приседания с низком грифом`
- `Болгарские сплит-приседания`

GPP groups:

- Abs: `-`, `Упражнение на пресс`
- Triceps: `-`, `Трицепс`
- Horizontal pull: `-`, `Тяга штанги в наклоне`, `Горизонтальный блок`, `Тяга гантели в наклоне`, `Рычажная горизонтальная тяга`
- Biceps: `-`, `Бицепс`
- Vertical pull: `-`, `Подтягивания`, `Вертикальный блок`, `Рычажная вертикальная тяга`
- Overhead press: `-`, `Армейский жим гантелей`, `Отжимания в стойке на руках`, `Армейский жим гирь`, `Армейский жим одной рукой`, `Армейский жим штанги`

## Training Layout

The rendered program area is on sheet `Программа`, columns `G:I`.

Day blocks:

| Day | Rows | Header |
| --- | --- | --- |
| Day 1 | `G4:I9` | `День1` |
| Day 2 | `G10:I15` | `День2` |
| Day 3 | `G16:I21` | `День3` |

Each exercise row has:

- exercise name;
- sets/reps;
- weight or RPE/comment;
- optional unit marker in column `J` for main calculated weights.

## Formula Notes

The spreadsheet uses `G3` as the selected week switch.

Main heavy lift progression:

- Day 1 heavy lift: deadlift, based on `B4`.
- Day 2 heavy lift: bench press, based on `B5`.
- Day 3 heavy lift: squat, based on `B6`.
- Weeks 1-7 progress around a base percentage.
- Week 8 is a test week and renders `1ПМ` instead of calculated working weights.

Base percentage:

- `Вариант 1`: `0.82`
- `Вариант 2`: `0.86`

Important compatibility rule:

- Variant changes only the base percentage.
- XLSX-compatible progression step is `4%` of the relevant 1RM, rounded to the nearest `2.5 kg`.
- `5%` progression can exist only as an explicit user/profile/cycle setting.
- Do not silently change `Вариант 2` to use a `5%` progression step.

Reference implementation:

```ts
function roundToNearest2_5(value: number): number {
  return Math.round(value / 2.5) * 2.5;
}

function baseWeight(oneRepMax: number, variant: "variant_1" | "variant_2"): number {
  const basePercent = variant === "variant_1" ? 0.82 : 0.86;
  return oneRepMax * basePercent;
}

function progressionStep(
  oneRepMax: number,
  step: "step_4_percent" | "step_5_percent" = "step_4_percent",
): number {
  const percent = step === "step_4_percent" ? 0.04 : 0.05;
  return roundToNearest2_5(oneRepMax * percent);
}
```

Examples:

| Lift | 1RM | Step |
| --- | ---: | ---: |
| Deadlift | `225 kg` | `10 kg` |
| Bench press | `125 kg` | `5 kg` |
| Squat | `170 kg` | `7.5 kg` |

Weekly progression pattern for heavy lift weights:

| Week | Formula shape |
| --- | --- |
| 1 | `base - step * 4` |
| 2 | `base - step * 3` |
| 3 | `base - step * 2` |
| 4 | `base - step` |
| 5 | `base` |
| 6 | `base + step` |
| 7 | `base + step * 2` |
| 8 | `1ПМ` |

Weights are rounded with Excel `MROUND(..., 2.5)`.

Known XLSX compatibility issue:

- Cells `Z15:Z17` use `IF(A23=AA17, ..., ...)`, but `AA17` is empty in the workbook.
- `AA17` is not hidden, not merged, not a named range, and not the source of the variant dropdown.
- The variant dropdown for `A23:B23` uses `$AA$15:$AA$16`.
- As written, the `5%` step branch is never selected and the formulas always use the `4%` branch.
- Therefore the XLSX-compatible implementation must use a `4%` progression step for both variants.
- A `5%` step is allowed only as an explicit profile/cycle setting.
- Any deliberate future change that makes `Вариант 2` imply `5%` automatically must be explicit, versioned, and covered by migration/recalculation notes.

## API/Domain Mapping

Recommended domain objects:

- `ProgramSelection`: 1RM values, week, variant, assistance choices, GPP choices.
- `ProgramCycle`: persisted user cycle with current settings and current week.
- `TrainingPlan`: calculated day/exercise rows from the current selection.
- `ExerciseDetails`: dataset metadata and GIF URL for a program-relevant exercise.
- `ProgressCheckpoint`: persisted user progress for a specific day/exercise/set or exercise row.

The backend should store cycle settings and progress checkpoints in a database. The XLSX formulas become deterministic domain calculations, not UI code.
