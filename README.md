# Remindme
Small cli tool to get desktop notifications for small reminders

# Usage
Currently it is only a command line tool.
It runs a http server locally on Port 3050 which listens for events and notifys you of reminders.
If you invoke the program without the server already running, it will act as the server,
meaning it is intended to be run on startup to start the server and then use the cli tool to interact with it.

For command usage simply invoke it with a running server and no arguments to get a list and explanation of all arguments.

# Example
  ```
  $remindme -after 1h20m -title "Workout" -msg "go to the gym"
  ```
  
# Installation
If you have Go installed simply run ```$go install github.com/bafto/remindme``` and configure the program in go/bin to run on startup.
If you do not have Go installed, download the program executable, add it to your PATH and configure it to run on startup.
