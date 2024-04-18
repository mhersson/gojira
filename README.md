# The Gojira JIRA client

This project is a product of me being bored out of my mind because of Corona
virus quarantine combined with Easter holidays.

Gojira is the Japanese word for Godzilla.

## Key Features

- Create issues
- Create and edit existing comments
- Create or edit worklogs for time reporting
- Import registered hours of colleagues to copy reporting (*)
- Import your own previously registered hours for reoccurring meetings (*)
- Show time reporting statistics (*)
- Update issue status and assignee
- Show comments, current status and the entire worklog
- One view to show it all with the describe command
- Display all unresolved issues assigned to you
- Display the current sprint with all issues and statuses
- Mark issue and/or board as active for less typing
- Use your favorite editor set by $EDITOR, defaults to vim
- Open issue in default browser
  
  (*) Only with timesheet plugin enabled

Gojira integrates with passwordstore and gpg to keep your password safe.

Gojira has full command tab completion support, and most commands have
assigned aliases to their first letter for less typing.

## Build Instructions

Gojira requires Go >= 1.16

Clone the repo then run `make build` and/or `make install`

## Install Instructions

After building make sure the Gojira executable is in your path.

Copy the `config-example.yaml` to $HOME/.config/gojira/config.yaml
and edit the JIRA server setting, username and password
