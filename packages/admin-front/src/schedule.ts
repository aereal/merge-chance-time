export const weekdays = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"] as const
export const [sunday, monday, tuesday, wednesday, thursday, friday, saturday] = weekdays
export type Weekday = typeof weekdays[number]

export const hours = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23] as const
export type Hour = typeof hours[number]
