import React, { FC } from "react"
import List from "@material-ui/core/List"
import { RepoSummary } from "./RepoSummary"
import { RepoSummary as RepoSummaryFragment } from "./__generated__/RepoSummary"

interface ReposListProps {
  readonly repos: RepoSummaryFragment[]
}

export const ReposList: FC<ReposListProps> = ({ repos }) => (
  <List>
    {repos.map((repo) => (
      <RepoSummary repo={repo} key={repo.fullName} />
    ))}
  </List>
)
