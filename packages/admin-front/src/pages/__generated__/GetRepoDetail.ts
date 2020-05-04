/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: GetRepoDetail
// ====================================================

export interface GetRepoDetail_repository_owner {
  readonly __typename: "User" | "Organization";
  readonly login: string;
}

export interface GetRepoDetail_repository_config_schedules_sunday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface GetRepoDetail_repository_config_schedules_monday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface GetRepoDetail_repository_config_schedules_tuesday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface GetRepoDetail_repository_config_schedules_wednesday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface GetRepoDetail_repository_config_schedules_thursday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface GetRepoDetail_repository_config_schedules_friday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface GetRepoDetail_repository_config_schedules_saturday {
  readonly __typename: "MergeChanceSchedule";
  readonly startHour: number;
  readonly stopHour: number;
}

export interface GetRepoDetail_repository_config_schedules {
  readonly __typename: "MergeChanceSchedules";
  readonly sunday: GetRepoDetail_repository_config_schedules_sunday | null;
  readonly monday: GetRepoDetail_repository_config_schedules_monday | null;
  readonly tuesday: GetRepoDetail_repository_config_schedules_tuesday | null;
  readonly wednesday: GetRepoDetail_repository_config_schedules_wednesday | null;
  readonly thursday: GetRepoDetail_repository_config_schedules_thursday | null;
  readonly friday: GetRepoDetail_repository_config_schedules_friday | null;
  readonly saturday: GetRepoDetail_repository_config_schedules_saturday | null;
}

export interface GetRepoDetail_repository_config {
  readonly __typename: "RepositoryConfig";
  readonly mergeAvailable: boolean;
  readonly schedules: GetRepoDetail_repository_config_schedules;
}

export interface GetRepoDetail_repository {
  readonly __typename: "Repository";
  readonly owner: GetRepoDetail_repository_owner;
  readonly name: string;
  readonly config: GetRepoDetail_repository_config | null;
}

export interface GetRepoDetail {
  readonly repository: GetRepoDetail_repository | null;
}

export interface GetRepoDetailVariables {
  readonly owner: string;
  readonly name: string;
}
