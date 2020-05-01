import React, { FC, ChangeEvent } from "react"
import Slider from "@material-ui/core/Slider"
import Typography from "@material-ui/core/Typography"
import red from "@material-ui/core/colors/red"
import indigo from "@material-ui/core/colors/indigo"
import makeStyles from "@material-ui/core/styles/makeStyles"
import { Weekday, sunday, saturday } from "../schedule"

export type OnUpdateValue = (value: number | number[] | null) => void
export type MergeChanceScheduleRange = [number, number]

interface WeekdayRangeSliderProps {
  readonly weekday: Weekday
  readonly scheduleRange: MergeChanceScheduleRange | null
  readonly onUpdateValue: OnUpdateValue
}

const useSyles = makeStyles({
  sundayLabel: {
    color: red[500],
  },
  saturdayLabel: {
    color: indigo[500],
  },
})

export const WeekdayRangeSlider: FC<WeekdayRangeSliderProps> = ({ weekday, onUpdateValue, scheduleRange }) => {
  const { sundayLabel, saturdayLabel } = useSyles()

  const colors: Partial<Record<Weekday, string>> = {
    [saturday]: saturdayLabel,
    [sunday]: sundayLabel,
  }
  const handleChange = (_: ChangeEvent<{}>, value: number | number[]): void => {
    onUpdateValue(value)
  }

  return (
    <>
      <Typography className={colors[weekday]} gutterBottom>
        {weekday}
      </Typography>
      <Slider
        marks
        valueLabelDisplay="auto"
        step={1}
        min={0}
        max={23}
        value={scheduleRange ?? [0, 23] /* TODO */}
        onChange={handleChange}
      />
    </>
  )
}
