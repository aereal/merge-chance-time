/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

import { RepositoryConfigToUpdate } from "./../../globalTypes";

// ====================================================
// GraphQL mutation operation: UpdateRepoConfig
// ====================================================

export interface UpdateRepoConfig {
  readonly updateRepositoryConfig: boolean;
}

export interface UpdateRepoConfigVariables {
  readonly owner: string;
  readonly name: string;
  readonly config: RepositoryConfigToUpdate;
}
