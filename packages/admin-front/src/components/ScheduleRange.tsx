import React, { FC, useState } from "react"
import {
  weekdays,
  Weekday,
  sunday,
  monday,
  tuesday,
  wednesday,
  thursday,
  friday,
  saturday,
  wholeDay,
} from "../schedule"
import { WeekdayRangeSlider, MergeChanceScheduleRange, OnUpdateValue } from "./WeekdayRangeSlider"

type WeekdaySchedules = Record<Weekday, MergeChanceScheduleRange | null>

const normalizeValues = (value: number | number[]): MergeChanceScheduleRange => {
  if (typeof value === "number") {
    return [value, value]
  }
  if (value.length !== 2) {
    throw new Error("Invalid value length")
  }
  const [from, to] = value
  return [from, to]
}

export const ScheduleRange: FC = () => {
  const [schedules, setSchedules] = useState<WeekdaySchedules>({
    [sunday]: wholeDay,
    [monday]: wholeDay,
    [tuesday]: wholeDay,
    [wednesday]: wholeDay,
    [thursday]: wholeDay,
    [friday]: wholeDay,
    [saturday]: wholeDay,
  })

  const sliderHandler = (weekday: Weekday): OnUpdateValue => (value) => {
    setSchedules((prev) => ({
      ...prev,
      [weekday]: value === null ? null : normalizeValues(value),
    }))
  }
  const handlers = weekdays.reduce(
    (handlers, wd) => ({
      ...handlers,
      [wd]: sliderHandler(wd),
    }),
    {} as { [k: string]: OnUpdateValue }
  )

  return (
    <>
      {weekdays.map((wd) => (
        <WeekdayRangeSlider weekday={wd} scheduleRange={schedules[wd]} onUpdateValue={handlers[wd]} key={`wd${wd}`} />
      ))}
    </>
  )
}
