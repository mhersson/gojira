## The Gojira JIRA client

This project is a product of me being bored out of my mind<br/>
because of Corona virus quarantine combined with Easter holidays.

Gojira is the Japanese word for Godzilla.

### Key Features:
  - Create issues
  - Add and update comments
  - Update status and assignee
  - Add work logs to report time spent
  - One view to show it all with the describe command
  - Display all unresolved issues currently assigned to you
  - Use your favorite editor to create and update issues and comments. Gojira uses
$EDITOR to determine which editor to use (defaults to vim).
  - Integrates with passwordstore and gpg to keep your password safe

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
