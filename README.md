## The Gojira JIRA client

This project is a product of me being bored out of my mind<br/>
because of Corona virus quarantine combined with Easter holidays.

Gojira is the Japanese word for Godzilla.

### Key Features:
  - Create new issues
  - Add and update comments
  - Use your favorite editor set by $EDITOR, defaults to vim
  - Change status
  - Assign to yourself or to others
  - Report time spent
  - Show comments, current status and the entire worklog
  - One view to show it all with the describe command
  - Display all unresolved issues assigned to you
  - Display the current sprint with all issues and statuses
  - Mark issue and/or board as active for less typing
  - Open issue in default browser
  

Gojira integrates with passwordstore and gpg to keep your password safe.

All commands have a short help text you can access by passing -h or --help.<br/>
Gojira has full command tab completion support, and most commands have<br/>
assigned aliases to their first letter for less typing.


### Build Instructions:
Install a Go version with support for Go modules >= 1.11<br/>
On Linux just install Go from your distributions package repositories.

Personally I like setting the [GOPATH](https://github.com/golang/go/wiki/SettingGOPATH)


Clone the repo then run `go build` and/or `go install`


### Install Instructions
After building make sure the Gojira executable is in your path.<br/>
Copy the `config-example.yaml` to $HOME/.config/gojira/config.yaml<br/>
and edit the JIRA server setting, username and password.<br/>
