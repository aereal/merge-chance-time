/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL mutation operation: UpdateRepoConfig
// ====================================================

export interface UpdateRepoConfig {
  readonly updateRepositoryConfig: boolean;
}

export interface UpdateRepoConfigVariables {
  readonly owner: string;
  readonly name: string;
  readonly startSchedule?: string | null;
  readonly stopSchedule?: string | null;
}
