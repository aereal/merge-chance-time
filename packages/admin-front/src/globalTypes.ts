/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

//==============================================================
// START Enums and Input Objects
//==============================================================

export interface MergeChanceScheduleToUpdate {
  readonly startHour: number;
  readonly stopHour: number;
}

export interface MergeChanceSchedulesToUpdate {
  readonly sunday?: MergeChanceScheduleToUpdate | null;
  readonly monday?: MergeChanceScheduleToUpdate | null;
  readonly tuesday?: MergeChanceScheduleToUpdate | null;
  readonly wednesday?: MergeChanceScheduleToUpdate | null;
  readonly thursday?: MergeChanceScheduleToUpdate | null;
  readonly friday?: MergeChanceScheduleToUpdate | null;
  readonly saturday?: MergeChanceScheduleToUpdate | null;
}

export interface RepositoryConfigToUpdate {
  readonly schedules: MergeChanceSchedulesToUpdate;
}

//==============================================================
// END Enums and Input Objects
//==============================================================
