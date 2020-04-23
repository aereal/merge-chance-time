/* tslint:disable */
/* eslint-disable */
// @generated
// This file was automatically generated and should not be edited.

// ====================================================
// GraphQL query operation: GetInstalledRepos
// ====================================================

export interface GetInstalledRepos_visitor_installations_installedRepositories_owner {
  readonly __typename: "User" | "Organization";
  readonly login: string;
}

export interface GetInstalledRepos_visitor_installations_installedRepositories {
  readonly __typename: "Repository";
  readonly fullName: string;
  readonly name: string;
  readonly owner: GetInstalledRepos_visitor_installations_installedRepositories_owner;
}

export interface GetInstalledRepos_visitor_installations {
  readonly __typename: "Installation";
  readonly installedRepositories: ReadonlyArray<GetInstalledRepos_visitor_installations_installedRepositories>;
}

export interface GetInstalledRepos_visitor {
  readonly __typename: "Visitor";
  readonly installations: ReadonlyArray<GetInstalledRepos_visitor_installations>;
}

export interface GetInstalledRepos {
  readonly visitor: GetInstalledRepos_visitor;
}
