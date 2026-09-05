CREATE TABLE IF NOT EXISTS program_options (id text PRIMARY KEY, options jsonb NOT NULL);

INSERT INTO program_options(id, options) VALUES (
    'xlsx-linear-cycle-v1',
    '{
  "weeks": [
    {
      "id": "week_1",
      "label": "Неделя 1"
    },
    {
      "id": "week_2",
      "label": "Неделя 2"
    },
    {
      "id": "week_3",
      "label": "Неделя 3"
    },
    {
      "id": "week_4",
      "label": "Неделя 4"
    },
    {
      "id": "week_5",
      "label": "Неделя 5"
    },
    {
      "id": "week_6",
      "label": "Неделя 6"
    },
    {
      "id": "week_7",
      "label": "Неделя 7"
    },
    {
      "id": "week_8",
      "label": "Неделя 8"
    }
  ],
  "variants": [
    {
      "id": "variant_1",
      "label": "Вариант 1"
    },
    {
      "id": "variant_2",
      "label": "Вариант 2"
    }
  ],
  "progressionSteps": [
    {
      "id": "step_4_percent",
      "label": "4% от 1ПМ"
    },
    {
      "id": "step_5_percent",
      "label": "5% от 1ПМ"
    }
  ],
  "assistance": {
    "deadlift": [
      {
        "id": "good_morning",
        "label": "Гуд-морнинг"
      },
      {
        "id": "romanian_deadlift",
        "label": "Румынская тяга"
      },
      {
        "id": "deficit_deadlift",
        "label": "Тяга из ямы"
      },
      {
        "id": "deadlift",
        "label": "Классическая становая тяга"
      },
      {
        "id": "sumo_deadlift",
        "label": "Становая тяга сумо"
      },
      {
        "id": "paused_deadlift",
        "label": "Становая тяга с паузами"
      }
    ],
    "bench": [
      {
        "id": "close_grip_bench",
        "label": "Жим узким хватом"
      },
      {
        "id": "reverse_grip_bench",
        "label": "Жим обратным хватом"
      },
      {
        "id": "incline_bench",
        "label": "Жим на наклонной скамье"
      },
      {
        "id": "dumbbell_bench",
        "label": "Жим гантелей лежа"
      },
      {
        "id": "dips",
        "label": "Брусья"
      }
    ],
    "squat": [
      {
        "id": "zercher_squat",
        "label": "Приседания Зерчера"
      },
      {
        "id": "front_squat",
        "label": "Приседания со штангой на груди"
      },
      {
        "id": "high_bar_squat",
        "label": "Приседания с высоким грифом"
      },
      {
        "id": "low_bar_squat",
        "label": "Приседания с низким грифом"
      },
      {
        "id": "bulgarian_split_squat",
        "label": "Болгарские сплит-приседания"
      }
    ]
  },
  "gpp": {
    "abs": [
      {
        "id": "",
        "label": "-"
      },
      {
        "id": "abs",
        "label": "Упражнение на пресс"
      }
    ],
    "triceps": [
      {
        "id": "",
        "label": "-"
      },
      {
        "id": "triceps",
        "label": "Трицепс"
      }
    ],
    "horizontalPull": [
      {
        "id": "",
        "label": "-"
      },
      {
        "id": "barbell_row",
        "label": "Тяга штанги в наклоне"
      },
      {
        "id": "cable_seated_row",
        "label": "Горизонтальный блок"
      },
      {
        "id": "dumbbell_row",
        "label": "Тяга гантели в наклоне"
      },
      {
        "id": "lever_horizontal_row",
        "label": "Рычажная горизонтальная тяга"
      }
    ],
    "biceps": [
      {
        "id": "",
        "label": "-"
      },
      {
        "id": "biceps",
        "label": "Бицепс"
      }
    ],
    "verticalPull": [
      {
        "id": "",
        "label": "-"
      },
      {
        "id": "pull_up",
        "label": "Подтягивания"
      },
      {
        "id": "lat_pulldown",
        "label": "Вертикальный блок"
      },
      {
        "id": "lever_vertical_row",
        "label": "Рычажная вертикальная тяга"
      }
    ],
    "overheadPress": [
      {
        "id": "",
        "label": "-"
      },
      {
        "id": "dumbbell_military_press",
        "label": "Армейский жим гантелей"
      },
      {
        "id": "handstand_push_up",
        "label": "Отжимания в стойке на руках"
      },
      {
        "id": "kettlebell_military_press",
        "label": "Армейский жим гирь"
      },
      {
        "id": "one_arm_military_press",
        "label": "Армейский жим одной рукой"
      },
      {
        "id": "barbell_military_press",
        "label": "Армейский жим штанги"
      }
    ]
  }
}'::jsonb
)
ON CONFLICT (id) DO NOTHING;
