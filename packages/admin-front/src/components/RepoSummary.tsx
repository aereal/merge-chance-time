import React, { FC } from "react"
import ListItem from "@material-ui/core/ListItem"
import ListItemText from "@material-ui/core/ListItemText"
import gql from "graphql-tag"
import { routes } from "../routes"
import { RepoSummary as RepoSummaryFragment } from "./__generated__/RepoSummary"

export const REPO_SUMMARY = gql`
  fragment RepoSummary on Repository {
    fullName
    name
    owner {
      login
    }
  }
`

interface RepoSummaryProps {
  readonly repo: RepoSummaryFragment
}

export const RepoSummary: FC<RepoSummaryProps> = ({ repo }) => (
  <ListItem button {...routes.repoDetail.link({ owner: repo.owner.login, name: repo.name })}>
    <ListItemText primary={repo.fullName} />
  </ListItem>
)
